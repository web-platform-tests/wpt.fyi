// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package backfill

import (
	"context"
	"errors"
	"time"

	"google.golang.org/api/option"

	"cloud.google.com/go/datastore"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/monitor"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// RunFetcher provides an interface for loading a limited number of runs
// suitable for backfilling an index.
type RunFetcher interface {
	FetchRuns(limit int) (shared.TestRunsByProduct, error)
}

type datastoreRunFetcher struct {
	projectID          string
	gcpCredentialsFile *string
	logger             shared.Logger
}

type backfillIndex struct {
	index.ProxyIndex

	backfilling bool
}

type backfillMonitor struct {
	monitor.ProxyMonitor

	idx *backfillIndex
}

// bytesPerRun is a slight over estimate of the memory requirements for one WPT
// run's indexed data. This value was determined experimentally in the early
// phases of search cache development.
const bytesPerRun = uint64(6.5e+7)

var errNilIndex = errors.New("Index to backfill is nil")

// NewDatastoreRunFetcher constructs a RunFetcher that loads runs from Datastore
// in reverse cronological order, by shared.TestRun.TimeStart.
func NewDatastoreRunFetcher(projectID string, gcpCredentialsFile *string, logger shared.Logger) RunFetcher {
	return datastoreRunFetcher{projectID, gcpCredentialsFile, logger}
}

func (f datastoreRunFetcher) FetchRuns(limit int) (shared.TestRunsByProduct, error) {
	ctx := context.Background()
	var client *datastore.Client
	var err error
	if f.gcpCredentialsFile != nil && *f.gcpCredentialsFile != "" {
		client, err = datastore.NewClient(ctx, f.projectID, option.WithCredentialsFile(*f.gcpCredentialsFile))
	} else {
		client, err = datastore.NewClient(ctx, f.projectID)
	}
	if err != nil {
		return nil, err
	}
	store := shared.NewCloudDatastore(ctx, client)

	// Query Datastore for latest maxBytes/bytesPerRun test runs.
	runs, err := shared.LoadTestRuns(store, nil, nil, nil, nil, nil, &limit, nil)
	return runs, nil
}

func (i *backfillIndex) EvictAnyRun() error {
	i.backfilling = false
	return i.ProxyIndex.EvictAnyRun()
}

func (m *backfillMonitor) Stop() error {
	m.idx.backfilling = false
	return m.ProxyMonitor.Stop()
}

func (*backfillIndex) Bind([]shared.TestRun, query.ConcreteQuery) (query.Plan, error) {
	return nil, nil
}

// FillIndex starts backfilling an index given a series of configuration
// parameters for run fetching and index monitoring. The backfilling process
// will halt either:
// The first time a run is evicted from the index.Index via EvictAnyRun(), OR
// the first time the returned monitor.Monitor is stopped via Stop().
func FillIndex(fetcher RunFetcher, logger shared.Logger, rt monitor.Runtime, interval time.Duration, maxBytes uint64, idx index.Index) (monitor.Monitor, error) {
	if idx == nil {
		return nil, errNilIndex
	}

	bfIdx := &backfillIndex{
		ProxyIndex:  index.NewProxyIndex(idx),
		backfilling: true,
	}
	bfMon := &backfillMonitor{
		ProxyMonitor: monitor.NewProxyMonitor(monitor.NewIndexMonitor(logger, rt, interval, maxBytes, bfIdx)),
		idx:          bfIdx,
	}

	err := startBackfillMonitor(fetcher, logger, maxBytes, bfMon)
	if err != nil {
		return nil, err
	}

	return bfMon, nil
}

func startBackfillMonitor(fetcher RunFetcher, logger shared.Logger, maxBytes uint64, m *backfillMonitor) error {
	runsByProduct, err := fetcher.FetchRuns(int(maxBytes/bytesPerRun) / 4)
	if err != nil {
		return err
	}
	if len(runsByProduct.AllRuns()) < 1 {
		return nil
	}

	// Start the monitor to ensure that memory pressure is tracked.
	go m.Start()

	// Backfill index until its backfilling parameter is set to false, or
	// collection of test runs is exhausted.
	go func() {
		most := 0
		for _, productRuns := range runsByProduct {
			if most < len(productRuns.TestRuns) {
				most = len(productRuns.TestRuns)
			}
		}
		for i := 0; i < most && m.idx.backfilling; i++ {
			for _, productRuns := range runsByProduct {
				if !m.idx.backfilling {
					logger.Warningf("Backfilling halted mid-iteration")
					break
				} else if i >= len(productRuns.TestRuns) {
					continue
				}
				run := productRuns.TestRuns[i]
				logger.Infof("Backfilling index with run %v", run)
				err = m.idx.IngestRun(run)
				if err != nil {
					logger.Errorf("Failed to ingest run during backfill: %v: %v", run, err)
				} else {
					logger.Infof("Backfilled index with run %v", run)
				}
			}
			logger.Infof("Backfilling complete")
		}
	}()

	return nil
}
