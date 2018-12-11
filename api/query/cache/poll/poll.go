// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package poll

import (
	"time"

	"github.com/web-platform-tests/wpt.fyi/api/query/cache/backfill"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// KeepRunsUpdated implements updates to an index.Index via simple polling every
// interval duration for at most limit runs loaded from fetcher.
func KeepRunsUpdated(fetcher backfill.RunFetcher, logger shared.Logger, interval time.Duration, limit int, idx index.Index) {
	// Start by waiting polling interval. This reduces the chance of false alarms
	// from log monitoring when KeepRunsUpdated is invoked around the same time as
	// index backfilling.
	logger.Infof("Starting index update via polling; waiting polling interval first...")
	time.Sleep(interval)
	logger.Infof("Index update via polling started")

	lastLoadTime := time.Now()
	for {
		start := time.Now()

		runs, err := fetcher.FetchRuns(limit)
		if err != nil {
			logger.Errorf("Error fetching runs for update: %v", err)
			wait(start, interval)
			continue
		}

		found := false
		for i, run := range runs {
			err := idx.IngestRun(run)
			if err != nil {
				if err == index.ErrRunExists() {
					logger.Infof("Not updating run (already exists): %v", run)
				} else if err == index.ErrRunLoading() {
					logger.Infof("Not updating run (already loading): %v", run)
				} else {
					logger.Errorf("Error ingesting run: %v: %v", run, err)
				}
			} else {
				logger.Infof("Updated run index; new run: %v", run)

				if i != 0 && !found {
					logger.Errorf("Runs loaded out of order: Skipped %d runs: %v, then loaded new run: %v", i, runs[:i], run)
				}
				found = true
				lastLoadTime = time.Now()
			}
		}

		if !found {
			logger.Warningf("No runs loaded throughout polling iteration. Last run update was at %v", lastLoadTime)
		}

		wait(start, interval)
	}
}

func wait(start time.Time, total time.Duration) {
	t := total - time.Now().Sub(start)
	if t > 0 {
		time.Sleep(t)
	}
}
