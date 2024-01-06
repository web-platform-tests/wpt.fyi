// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// SetWebFeaturesDataCache safely swaps the data cache.
// Currently, the separate goroutine in the poll folder will use this.
func SetWebFeaturesDataCache(newData shared.WebFeaturesData) {
	webFeaturesDataCacheLock.Lock()
	webFeaturesDataCache = newData
	defer webFeaturesDataCacheLock.Unlock()
}

// webFeaturesDataCache is the local cache of Web Features data in searchcache. Zero value is nil.
var webFeaturesDataCache shared.WebFeaturesData // nolint:gochecknoglobals // TODO: Fix gochecknoglobals lint error
var webFeaturesDataCacheLock sync.RWMutex       // nolint:gochecknoglobals // TODO: Fix gochecknoglobals lint error

type searchcacheWebFeaturesManifestFetcher struct{}

func (f searchcacheWebFeaturesManifestFetcher) Fetch() (shared.WebFeaturesData, error) {
	logger := shared.GetLogger(context.Background())
	logger.Infof("starting manifest fetch")
	webFeaturesDataCacheLock.RLock()
	defer webFeaturesDataCacheLock.RUnlock()
	logger.Infof("checking for manifest cache in fetch")
	if webFeaturesDataCache != nil {
		logger.Infof("manifest exists in cache")
		return webFeaturesDataCache, nil
	}
	logger.Infof("fetch could not find manifest cache")

	netClient := &http.Client{
		Timeout: time.Second * 5,
	}
	ctx := context.Background()
	gitHubClient, err := shared.NewAppEngineAPI(ctx).GetGitHubClient()
	if err != nil {
		shared.GetLogger(ctx).Warningf("unable to get github client for searchcache")

		return nil, err
	}
	downloader := shared.NewGitHubWebFeaturesManifestDownloader(netClient, gitHubClient)

	data, err := shared.GetWPTWebFeaturesManifest(ctx, downloader, shared.WebFeaturesManifestJSONParser{})
	if err != nil {
		shared.GetLogger(ctx).Errorf("unable to fetch web features manifest during query. %s", err.Error())

		return nil, err
	}
	logger.Infof("retrieved manifest data for cache")

	return data, nil
}
