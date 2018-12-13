// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/api/manifest"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/memcache"
)

func apiManifestHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	sha, err := shared.ParseSHAParamFull(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	paths := shared.ParsePathsParam(q)

	ctx := shared.NewAppEngineContext(r)
	manifestAPI := manifest.NewAPI(ctx)
	sha, manifest, err := getManifest(ctx, manifestAPI, sha, paths)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Add("wpt-sha", sha)
	w.Header().Add("Content-Type", "application/json")
	w.Write(manifest)
}

func getManifest(ctx context.Context, manifestAPI manifest.API, sha string, paths []string) (string, []byte, error) {
	if sha == "" {
		sha = "latest"
	}
	fetchedSHA := sha
	var body []byte
	cached, err := memcache.Get(ctx, manifestCacheKey(sha))

	if err != nil && err != memcache.ErrCacheMiss {
		return "", nil, err
	} else if cached != nil {
		// "latest" caches which SHA is latest; Return the manifest for that SHA.
		if sha == "latest" {
			return getManifest(ctx, manifestAPI, string(cached.Value), paths)
		}
		body = cached.Value
	} else {
		fetchedSHA, body, err = manifestAPI.GetManifestForSHA(sha)
		if paths != nil {
			if body, err = manifest.Filter(body, paths); err != nil {
				return fetchedSHA, nil, err
			}
		}
	}

	item := &memcache.Item{
		Key:   manifestCacheKey(fetchedSHA),
		Value: body,
	}
	memcache.Set(ctx, item)

	// Shorter expiry for latest SHA, to keep it current.
	if shared.IsLatest(sha) {
		latestSHAItem := &memcache.Item{
			Key:        manifestCacheKey("latest"),
			Value:      []byte(fetchedSHA),
			Expiration: time.Minute * 5,
		}
		memcache.Set(ctx, latestSHAItem)
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return fetchedSHA, nil, err
	}
	body, err = ioutil.ReadAll(gzReader)
	return fetchedSHA, body, err
}

func manifestCacheKey(sha string) string {
	return fmt.Sprintf("MANIFEST-%s", sha)
}
