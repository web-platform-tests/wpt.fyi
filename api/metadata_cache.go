// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/memcache"
)

const metadataCacheKey = "WPT-METADATA"

type webappMetadataFetcher struct {
	ctx    context.Context
	client *http.Client
	log    shared.Logger
	url    string
}

func (f webappMetadataFetcher) Fetch() (res map[string][]byte, err error) {
	metadataMap, err := getMetadataFromMemcache(f.ctx, f.log, f.client, f.url)
	if err == nil && metadataMap != nil {
		return metadataMap, nil
	}

	return shared.CollectMetadataWithURL(f.client, f.url)
}

func getMetadataFromMemcache(ctx context.Context, log shared.Logger, client *http.Client, url string) (res map[string][]byte, err error) {
	cached, err := memcache.Get(ctx, metadataCacheKey)

	if err != nil && err != memcache.ErrCacheMiss {
		log.Errorf("Error from getting Metadata in memcache: %s", err.Error())
		return nil, err
	}

	if err == nil && cached != nil {
		// Caches hit; update Metadata.
		var metadataMap map[string][]byte
		err = json.Unmarshal(cached.Value, &metadataMap)
		if err != nil {
			log.Errorf("Error from unmarshaling Metadata in memcache: %s", err.Error())
			return nil, err
		}

		return metadataMap, nil
	}

	// Caches missed.
	metadataByteMap, err := shared.CollectMetadataWithURL(client, url)
	if err != nil {
		log.Errorf("Error from CollectMetadataWithURL in a cache miss: %s", err.Error())
		return nil, err
	}

	body, err := json.Marshal(metadataByteMap)
	if err != nil {
		log.Errorf("Error from marshaling metadataByteMap in a cache miss: %s", err.Error())
		return metadataByteMap, nil
	}

	item := &memcache.Item{
		Key:        metadataCacheKey,
		Value:      body,
		Expiration: time.Minute * 10,
	}
	memcache.Set(ctx, item)
	return metadataByteMap, nil
}
