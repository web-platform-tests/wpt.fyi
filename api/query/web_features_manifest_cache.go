// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"context"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// SetWebFeaturesDataCache safely swaps the data cache.
// Currently, the separate goroutine in the poll folder will use this.
func SetWebFeaturesDataCache(newData shared.WebFeaturesData) {
	shared.GetLogger(context.Background()).Infof("setting data cache for manifest")
	webFeaturesDataCache = newData
}

// webFeaturesDataCache is the local cache of Web Features data in searchcache. Zero value is nil.
var webFeaturesDataCache shared.WebFeaturesData // nolint:gochecknoglobals // TODO: Fix gochecknoglobals lint error

type searchcacheWebFeaturesManifestFetcher struct{}

func (f searchcacheWebFeaturesManifestFetcher) Fetch() (shared.WebFeaturesData, error) {
	if webFeaturesDataCache != nil {
		return webFeaturesDataCache, nil
	}

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

	return data, nil
}
