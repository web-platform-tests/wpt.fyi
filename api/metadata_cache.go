// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api //nolint:revive

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/go-github/v79/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

const metadataCacheKey = "WPT-METADATA"
const metadataCacheExpiry = time.Minute * 10

type webappMetadataFetcher struct {
	ctx          context.Context // nolint:containedctx // TODO: Fix containedctx lint error
	httpClient   *http.Client
	gitHubClient *github.Client
	forceUpdate  bool
}

func (f webappMetadataFetcher) Fetch() (sha *string, res map[string][]byte, err error) {
	log := shared.GetLogger(f.ctx)
	mCache := shared.NewJSONObjectCache(f.ctx, shared.NewRedisReadWritable(f.ctx, metadataCacheExpiry))
	if !f.forceUpdate {
		sha, metadataMap, err := getMetadataFromRedis(mCache)
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

	if err := fillMetadataToRedis(mCache, *sha, res); err != nil {
		// This is not a fatal failure.
		log.Errorf("Error storing metadata to cache: %v", err)
	}

	return sha, res, nil
}

func getMetadataFromRedis(cache shared.ObjectCache) (sha *string, res map[string][]byte, err error) {
	var metadataSHAMap map[string]map[string][]byte
	err = cache.Get(metadataCacheKey, &metadataSHAMap)
	if err != nil {
		return nil, nil, err
	}

	// Caches hit; update Metadata.
	keys := make([]string, 0, len(metadataSHAMap))
	for key := range metadataSHAMap {
		keys = append(keys, key)
	}

	if len(keys) != 1 {
		return nil, nil, errors.New("error from getting the wpt-metadata SHA in metadataSHAMap")
	}

	sha = &keys[0]

	return sha, metadataSHAMap[*sha], nil
}

func fillMetadataToRedis(cache shared.ObjectCache, sha string, metadataByteMap map[string][]byte) error {
	metadataSHAMap := make(map[string]map[string][]byte)
	metadataSHAMap[sha] = metadataByteMap

	return cache.Put(metadataCacheKey, metadataSHAMap)
}
