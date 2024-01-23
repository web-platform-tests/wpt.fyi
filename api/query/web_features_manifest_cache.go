// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"context"
	"sync"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// SetWebFeaturesDataCache safely swaps the data cache.
// Currently, the separate goroutine in the poll folder will use this.
func SetWebFeaturesDataCache(newData shared.WebFeaturesData) {
	webFeaturesDataCacheLock.Lock()
	defer webFeaturesDataCacheLock.Unlock()
	webFeaturesDataCache = newData
}

// GetWebFeaturesDataCache safely retrieves the data cache.
func GetWebFeaturesDataCache() shared.WebFeaturesData {
	webFeaturesDataCacheLock.RLock()
	defer webFeaturesDataCacheLock.RUnlock()

	return webFeaturesDataCache
}

// webFeaturesDataCache is the local cache of Web Features data in searchcache. Zero value is nil.
var webFeaturesDataCache shared.WebFeaturesData // nolint:gochecknoglobals // TODO: Fix gochecknoglobals lint error
var webFeaturesDataCacheLock sync.RWMutex       // nolint:gochecknoglobals // TODO: Fix gochecknoglobals lint error

type searchcacheWebFeaturesManifestFetcher struct{}

func (f searchcacheWebFeaturesManifestFetcher) Fetch() (shared.WebFeaturesData, error) {
	cache := GetWebFeaturesDataCache()
	if cache != nil {
		return cache, nil
	}

	ctx := context.Background()
	gitHubClient, err := shared.NewAppEngineAPI(ctx).GetGitHubClient()
	if err != nil {
		shared.GetLogger(ctx).Warningf("unable to get github client for searchcache")

		return nil, err
	}

	featuresClient := shared.NewGitHubWebFeaturesClient(gitHubClient)
	data, err := featuresClient.Get(ctx)
	if err != nil {
		shared.GetLogger(ctx).Warningf("github client unable to get features for searchcache")

		return nil, err
	}

	return data, nil
}
