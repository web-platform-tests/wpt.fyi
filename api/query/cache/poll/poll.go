// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package poll

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v77/github"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// KeepRunsUpdated implements updates to an index.Index via simple polling every
// interval duration for at most limit runs loaded from fetcher.
// nolint:gocognit // TODO: Fix gocognit lint error
func KeepRunsUpdated(store shared.Datastore, logger shared.Logger, interval time.Duration, limit int, idx index.Index) {
	// Start by waiting polling interval. This reduces the chance of false alarms
	// from log monitoring when KeepRunsUpdated is invoked around the same time as
	// index backfilling.
	logger.Infof("Starting index update via polling; waiting polling interval first...")
	time.Sleep(interval)
	logger.Infof("Index update via polling started")

	lastLoadTime := time.Now()
	for {
		start := time.Now()

		runs, err := store.TestRunQuery().LoadTestRuns(shared.GetDefaultProducts(), nil, nil, nil, nil, &limit, nil)
		if err != nil {
			logger.Errorf("Error fetching runs for update: %v", err)
			wait(start, interval)

			continue
		}
		if len(runs) == 0 {
			logger.Errorf("Fetcher produced no runs for update")
			wait(start, interval)

			continue
		}

		errs := make([]error, len(runs))
		found := false
		for i, browserRuns := range runs {
			for _, run := range browserRuns.TestRuns {
				err := idx.IngestRun(run)
				errs[i] = err
				if err != nil {
					if errors.Is(err, index.ErrRunExists()) {
						logger.Debugf("Not updating run (already exists): %v", run)
					} else if errors.Is(err, index.ErrRunLoading()) {
						logger.Debugf("Not updating run (already loading): %v", run)
					} else {
						logger.Errorf("Error ingesting run: %v: %v", run, err)
					}
				} else {
					logger.Debugf("Updated run index; new run: %v", run)
					found = true
					lastLoadTime = time.Now()
				}
			}
		}

		if !found {
			logger.Infof("No runs loaded throughout polling iteration. Last run update was at %v", lastLoadTime)
		} else {
			next := errs[1:]
			for i := range next {
				if errs[i] != nil && next[i] == nil {
					logger.Errorf("Ingested run after skipping %d runs; ingest run attempt errors: %v", i, errs)

					break
				}
			}
		}

		wait(start, interval)
	}
}

func wait(start time.Time, total time.Duration) {
	t := total - time.Since(start)
	if t > 0 {
		time.Sleep(t)
	}
}

// StartMetadataPollingService performs metadata-related services via simple polling every
// interval duration.
func StartMetadataPollingService(ctx context.Context, logger shared.Logger, interval time.Duration) {
	logger.Infof("Starting Metadata polling service.")
	toBeRemovedPRs := make([]string, 0)
	netClient := &http.Client{Timeout: time.Second * 5}
	cacheSet := shared.NewRedisSet()
	gitHubClient, err := shared.NewAppEngineAPI(ctx).GetGitHubClient()
	if err != nil {
		logger.Infof("Unable to get GitHub client: %v", err)
	}

	for {
		keepMetadataUpdated(netClient, logger)
		if gitHubClient != nil {
			cleanOrphanedPendingMetadata(ctx, gitHubClient, cacheSet, logger, &toBeRemovedPRs)
		} else {
			logger.Infof("GitHub client is not initialized, skipping cleanOrphanedPendingMetadata.")
		}
		time.Sleep(interval)
	}
}

// keepMetadataUpdated fetches a new copy of the wpt-metadata repo and updates metadataMapCached.
func keepMetadataUpdated(client *http.Client, logger shared.Logger) {
	logger.Infof("Running keepMetadataUpdated...")
	metadataCache, err := shared.GetWPTMetadataArchive(client, nil)
	if err != nil {
		logger.Infof("Error fetching Metadata for update: %v", err)

		return
	}

	if metadataCache != nil {
		query.MetadataMapCached = metadataCache
	}
}

// cleanOrphanedPendingMetadata cleans and removes orphaned pending metadata in Redis.
func cleanOrphanedPendingMetadata(
	ctx context.Context,
	ghClient *github.Client,
	cacheSet shared.RedisSet,
	logger shared.Logger,
	toBeRemovedPRs *[]string,
) {
	logger.Infof("Running cleanOrphanedPendingMetadata...")

	for _, pr := range *toBeRemovedPRs {
		logger.Infof("Removing PR %s and its pending metadata from Redis", pr)
		err := cacheSet.Remove(shared.PendingMetadataCacheKey, pr)
		if err != nil {
			logger.Warningf("Error removing %s from RedisSet: %s", pr, err.Error())
		}
		err = shared.DeleteCache(shared.PendingMetadataCachePrefix + pr)
		if err != nil {
			logger.Warningf("Error removing %s from Redis: %s", pr, err.Error())
		}
	}

	prs, err := cacheSet.GetAll(shared.PendingMetadataCacheKey)
	if err != nil {
		logger.Infof("Error fetching pending PRs from cacheSet: %v", err)

		return
	}
	logger.Infof("Pending PR numbers in cacheSet are: %v", prs)

	newRemovePRs := make([]string, 0)
	for _, pr := range prs {
		// Parse PR string into integer
		prInt, err := strconv.Atoi(pr)
		if err != nil {
			logger.Infof("Error parsing %s into integer in cleanOrphanedPendingMetadata", pr)
			// Not an integer; remove it.
			newRemovePRs = append(newRemovePRs, pr)

			continue
		}

		res, _, err := ghClient.PullRequests.Get(ctx, shared.SourceOwner, shared.SourceRepo, prInt)
		if err != nil {
			logger.Infof("Error getting information for PR %s: %v", pr, err)

			continue
		}

		if res.State == nil || *res.State != "closed" {
			continue
		}
		newRemovePRs = append(newRemovePRs, pr)
	}
	*toBeRemovedPRs = newRemovePRs
}

// StartWebFeaturesManifestPollingService performs web features manifest related
// services via simple polling every interval duration.
func StartWebFeaturesManifestPollingService(ctx context.Context, logger shared.Logger, interval time.Duration) {
	logger.Infof("Starting web features manifest polling service.")
	gitHubClient, err := shared.NewAppEngineAPI(ctx).GetGitHubClient()
	if err != nil {
		logger.Infof("Unable to get GitHub client: %v", err)
	}

	for {
		if gitHubClient != nil {
			keepWebFeaturesManifestUpdated(
				ctx,
				logger,
				shared.NewGitHubWebFeaturesClient(gitHubClient))
		} else {
			logger.Infof("GitHub client is not initialized, skipping keepWebFeaturesManifestUpdated.")
		}
		time.Sleep(interval)
	}
}

// webFeaturesGetter provides a thin interface to get web features data.
type webFeaturesGetter interface {
	Get(context.Context) (shared.WebFeaturesData, error)
}

// keepWebFeaturesManifestUpdated fetches a new copy of the web features data
// and updates the local cache.
func keepWebFeaturesManifestUpdated(
	ctx context.Context,
	logger shared.Logger,
	featuresGetter webFeaturesGetter) {
	logger.Infof("Running keepWebFeaturesManifestUpdated...")

	data, err := featuresGetter.Get(ctx)
	if err != nil {
		logger.Errorf("unable to fetch web features manifest during query. %s", err.Error())

		return
	}
	if data != nil {
		query.SetWebFeaturesDataCache(data)
	}
}
