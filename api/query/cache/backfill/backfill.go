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
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/monitor"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// RunFetcher provides an interface for loading a limited number of runs
// suitable for backfilling an index.
type RunFetcher interface {
	FetchRuns(limit int) ([]shared.TestRun, error)
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

const bytesPerRun = uint64(1.45e+8)

var errNilIndex = errors.New("Index to backfill is nil")

// NewDatastoreRunFetcher constructs a RunFetcher that loads runs from Datastore
// in reverse cronological order, by shared.TestRun.TimeStart.
func NewDatastoreRunFetcher(projectID string, gcpCredentialsFile *string, logger shared.Logger) RunFetcher {
	return datastoreRunFetcher{projectID, gcpCredentialsFile, logger}
}

func (f datastoreRunFetcher) FetchRuns(limit int) ([]shared.TestRun, error) {
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

	// Query Datastore for latest maxBytes/bytesPerRun test runs.
	q := datastore.NewQuery("TestRun").Order("-TimeStart").Limit(limit)
	var runs []shared.TestRun
	keys, err := client.GetAll(ctx, q, &runs)
	if err != nil {
		return nil, err
	}

	// Ensure that runs contain IDs corresponding to Datastore keys.
	for i := range keys {
		runs[i].ID = keys[i].ID
	}

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

// FillIndex starts backfilling an index given a series of configuration
// parameters for run fetching and index monitoring. The backfilling process
// will halt either:
// The first time a run is evicted from the index.Index via EvictAnyRun(), OR
// the first time the returned monitor.Monitor is stopped via Stop().
func FillIndex(fetcher RunFetcher, logger shared.Logger, rt monitor.Runtime, freq time.Duration, maxBytes uint64, idx index.Index) (monitor.Monitor, error) {
	if idx == nil {
		return nil, errNilIndex
	}

	bfIdx := &backfillIndex{
		ProxyIndex:  index.NewProxyIndex(idx),
		backfilling: true,
	}
	bfMon := &backfillMonitor{
		ProxyMonitor: monitor.NewProxyMonitor(monitor.NewIndexMonitor(logger, rt, freq, maxBytes, bfIdx)),
		idx:          bfIdx,
	}

	err := startBackfillMonitor(fetcher, logger, maxBytes, bfMon)
	if err != nil {
		return nil, err
	}

	return bfMon, nil
}

func startBackfillMonitor(fetcher RunFetcher, logger shared.Logger, maxBytes uint64, m *backfillMonitor) error {
	runs, err := fetcher.FetchRuns(int(maxBytes / bytesPerRun))
	if err != nil {
		return err
	}

	// Start the monitor to ensure that memory pressure is tracked.
	go m.Start()

	// Backfill index until its backfilling parameter is set to false, or
	// collection of test runs is exhausted.
	go func() {
		for _, run := range runs {
			if !m.idx.backfilling {
				logger.Warningf("Backfilling halted mid-iteration")
				break
			}
			logger.Infof("Backfilling index with run %v", run)
			err = m.idx.IngestRun(run)
			if err != nil {
				logger.Errorf("Failed to ingest run during backfill: %v: %v", run, err)
			} else {
				logger.Infof("Backfilled index with run %v", run)
			}
		}
		logger.Infof("Backfilling complete")
	}()

	return nil
}
