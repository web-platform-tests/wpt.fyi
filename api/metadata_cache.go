// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/memcache"
)

const metadataCacheKey = "WPT-METADATA"

type webappMetadataFetcher struct {
	ctx           context.Context
	client        *http.Client
	url           string
	gitHubUtil    shared.GitHubUtil
	forceUpdate bool
}

func (f webappMetadataFetcher) Fetch() (sha *string, res map[string][]byte, err error) {
	if !f.forceUpdate {
		sha, metadataMap, err := getMetadataFromMemcache(f.ctx)
		if err == nil && metadataMap != nil && sha != nil {
			return sha, metadataMap, nil
		}
	}

	sha, err = f.gitHubUtil.GetWPTMetadataMasterSHA()
	if err != nil {
		return nil, nil, err
	}

	res, err = shared.CollectMetadataWithURL(f.client, f.url, sha)
	if err != nil {
		return nil, nil, err
	}

	// Caches missed.
	fillMetadataToMemcache(f.ctx, *sha, res)
	return sha, res, err
}

func getMetadataFromMemcache(ctx context.Context) (sha *string, res map[string][]byte, err error) {
	log := shared.GetLogger(ctx)
	cached, err := memcache.Get(ctx, metadataCacheKey)

	if err != nil && err != memcache.ErrCacheMiss {
		log.Errorf("Error from getting Metadata in memcache: %s", err.Error())
		return nil, nil, err
	}

	if err == nil && cached != nil {
		// Caches hit; update Metadata.
		var metadataSHAMap map[string]map[string][]byte
		err = json.Unmarshal(cached.Value, &metadataSHAMap)
		if err != nil {
			log.Errorf("Error from unmarshaling Metadata in memcache: %s", err.Error())
			return nil, nil, err
		}

		var keys []string
		for key := range metadataSHAMap {
			keys = append(keys, key)
		}

		if len(keys) != 1 {
			log.Errorf("Error from getting the wpt-metadata SHA in metadataSHAMap")
			return nil, nil, errors.New("Error from getting the wpt-metadata SHA in metadataSHAMap")
		}

		sha = &keys[0]
		return sha, metadataSHAMap[*sha], nil
	}

	return nil, nil, memcache.ErrCacheMiss
}

func fillMetadataToMemcache(ctx context.Context, sha string, metadataByteMap map[string][]byte) {
	log := shared.GetLogger(ctx)

	var metadataSHAMap = make(map[string]map[string][]byte)
	metadataSHAMap[sha] = metadataByteMap
	body, err := json.Marshal(metadataSHAMap)
	if err != nil {
		log.Errorf("Error from marshaling metadataSHAMap in a cache miss: %s", err.Error())
	}

	item := &memcache.Item{
		Key:        metadataCacheKey,
		Value:      body,
		Expiration: time.Minute * 10,
	}
	err = memcache.Set(ctx, item)
	if err != nil {
		log.Errorf("Error from memcache.Set in a cache miss: %s", err.Error())
	}
}
