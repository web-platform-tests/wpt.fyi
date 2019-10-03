// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package backfill

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"

	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/monitor"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

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

// GetDatastore constructs a shared.Datastore interface that loads runs from Datastore
// in reverse cronological order, by shared.TestRun.TimeStart.
func GetDatastore(projectID string, gcpCredentialsFile *string, logger shared.Logger) (shared.Datastore, error) {
	ctx := context.WithValue(context.Background(), shared.DefaultLoggerCtxKey(), logrus.StandardLogger())
	var client *datastore.Client
	var err error
	if gcpCredentialsFile != nil && *gcpCredentialsFile != "" {
		client, err = datastore.NewClient(ctx, projectID, option.WithCredentialsFile(*gcpCredentialsFile))
	} else {
		client, err = datastore.NewClient(ctx, projectID)
	}
	if err != nil {
		return nil, err
	}
	return shared.NewCloudDatastore(ctx, client), nil
}

func (i *backfillIndex) EvictRuns(percent float64) (int, error) {
	i.backfilling = false
	return i.ProxyIndex.EvictRuns(percent)
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
func FillIndex(store shared.Datastore, logger shared.Logger, rt monitor.Runtime, interval time.Duration, maxIngestedRuns uint, maxBytes uint64, evictionPercent float64, idx index.Index) (monitor.Monitor, error) {
	if idx == nil {
		return nil, errNilIndex
	}

	bfIdx := &backfillIndex{
		ProxyIndex:  index.NewProxyIndex(idx),
		backfilling: true,
	}
	idxMon, err := monitor.NewIndexMonitor(logger, rt, interval, maxIngestedRuns, maxBytes, evictionPercent, bfIdx)
	if err != nil {
		return nil, err
	}
	bfMon := &backfillMonitor{
		ProxyMonitor: monitor.NewProxyMonitor(idxMon),
		idx:          bfIdx,
	}

	err = startBackfillMonitor(store, logger, maxBytes, bfMon)
	if err != nil {
		return nil, err
	}

	return bfMon, nil
}

func startBackfillMonitor(store shared.Datastore, logger shared.Logger, maxBytes uint64, m *backfillMonitor) error {
	// FetchRuns will return at most N runs for each product, so divide the upper bound by the number of products.
	limit := int(maxBytes/bytesPerRun) / len(shared.GetDefaultProducts())
	runsByProduct, err := store.TestRunQuery().LoadTestRuns(shared.GetDefaultProducts(), nil, nil, nil, nil, &limit, nil)
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
