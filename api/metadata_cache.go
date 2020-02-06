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

func updateMetadataFromMemcache(ctx context.Context, log shared.Logger, client *http.Client) {
	cached, err := memcache.Get(ctx, metadataCacheKey)

	if err != nil && err != memcache.ErrCacheMiss {
		log.Errorf("Error from getting Metadata in memcache: %s", err.Error())
		return
	}

	if err == nil {
		// Caches hit; update Metadata.
		var metadataMap map[string][]byte
		err = json.Unmarshal(cached.Value, &metadataMap)
		if err != nil {
			log.Errorf("Error from unmarshaling Metadata in memcache: %s", err.Error())
			return
		}

		shared.UpdatedMetadataInWebapp(metadataMap)
		return
	}

	// Caches missed.
	metadataByteMap, err := shared.UpdatedMetadata(client, log)
	if err != nil {
		log.Errorf("Error from UpdatedMetadata in a cache miss: %s", err.Error())
		return
	}

	body, err := json.Marshal(metadataByteMap)
	if err != nil {
		log.Errorf("Error from marshaling metadataByteMap in a cache miss: %s", err.Error())
		return
	}

	item := &memcache.Item{
		Key:        metadataCacheKey,
		Value:      body,
		Expiration: time.Minute * 10,
	}
	memcache.Set(ctx, item)
}
