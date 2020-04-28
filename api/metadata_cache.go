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
	ctx        context.Context
	client     *http.Client
	url        string
	gitHubUtil shared.GitHubUtil
}

func (f webappMetadataFetcher) Fetch() (sha *string, res map[string][]byte, err error) {
	sha, metadataMap, err := getMetadataFromMemcache(f.ctx, f.client, f.url, f.gitHubUtil)
	if err == nil && metadataMap != nil && sha != nil {
		return sha, metadataMap, nil
	}

	shaKey, err := f.gitHubUtil.GetWPTMetadataMasterSHA()
	if err != nil {
		return nil, nil, err
	}

	// CollectMetadataWithURL retrives the content of the wpt-metadata repo from master by default.
	res, err = shared.CollectMetadataWithURL(f.client, f.url)
	return shaKey, res, err
}

func getMetadataFromMemcache(ctx context.Context, client *http.Client, url string, gitHubUtil shared.GitHubUtil) (sha *string, res map[string][]byte, err error) {
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

		shaKey := keys[0]
		return &shaKey, metadataSHAMap[shaKey], nil
	}

	// Caches missed.
	shaKey, err := gitHubUtil.GetWPTMetadataMasterSHA()
	if err != nil {
		log.Errorf("Error from getWPTMetadataMasterSHA in a cache miss: %s", err.Error())
		return nil, nil, err
	}

	metadataByteMap, err := shared.CollectMetadataWithURL(client, url)
	if err != nil {
		log.Errorf("Error from CollectMetadataWithURL in a cache miss: %s", err.Error())
		return nil, nil, err
	}

	var metadataSHAMap = make(map[string]map[string][]byte)
	metadataSHAMap[*shaKey] = metadataByteMap
	body, err := json.Marshal(metadataSHAMap)
	if err != nil {
		log.Errorf("Error from marshaling metadataSHAMap in a cache miss: %s", err.Error())
		return shaKey, metadataByteMap, nil
	}

	item := &memcache.Item{
		Key:        metadataCacheKey,
		Value:      body,
		Expiration: time.Minute * 10,
	}
	memcache.Set(ctx, item)
	return shaKey, metadataByteMap, nil
}
