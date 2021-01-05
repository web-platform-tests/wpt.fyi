// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"net/http"

	"github.com/google/go-github/v32/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

type webappMetadataFetcher struct {
	ctx          context.Context
	httpClient   *http.Client
	gitHubClient *github.Client
	forceUpdate  bool
}

func (f webappMetadataFetcher) Fetch() (sha *string, res map[string][]byte, err error) {
	log := shared.GetLogger(f.ctx)
	mCache := shared.NewJSONObjectCache(f.ctx, shared.NewMemcacheReadWritable(f.ctx, shared.MetadataCacheExpiry))
	if !f.forceUpdate {
		sha, metadataMap, err := shared.GetMetadataFromMemcache(mCache)
		if err == nil {
			return sha, metadataMap, nil
		}
		log.Debugf("Metadata cache missed: %v", err)
	}

	sha, err = shared.GetWPTMetadataMasterSHA(f.ctx, f.gitHubClient)
	if err != nil {
		log.Errorf("Error getting HEAD SHA of wpt-metadata: %v", err)
		return nil, nil, err
	}

	res, err = shared.GetWPTMetadataArchive(f.httpClient, sha)
	if err != nil {
		log.Errorf("Error getting archive of wpt-metadata: %v", err)
		return nil, nil, err
	}

	if err := fillMetadataToMemcache(mCache, *sha, res); err != nil {
		// This is not a fatal failure.
		log.Errorf("Error storing metadata to cache: %v", err)
	}

	return sha, res, nil
}

func fillMetadataToMemcache(cache shared.ObjectCache, sha string, metadataByteMap map[string][]byte) error {
	metadataSHAMap := make(map[string]map[string][]byte)
	metadataSHAMap[sha] = metadataByteMap

	return cache.Put(shared.MetadataCacheKey, metadataSHAMap)
}
