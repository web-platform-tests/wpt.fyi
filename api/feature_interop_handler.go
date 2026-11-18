// Copyright 2026 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api //nolint:revive

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// FeatureInteropHandler is an http.Handler for the /api/interop-features endpoint.
type FeatureInteropHandler struct {
	fetcher shared.FetchFeatureInterop
}

// apiFeatureInteropHandler fetches per-web-feature interop scores based on the URL params.
func apiFeatureInteropHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Serve cached with 60-minute expiry. Delegate to FeatureInteropHandler on cache miss.
	shared.NewCachingHandler(
		ctx,
		FeatureInteropHandler{shared.NewFetchFeatureInterop()},
		shared.NewGZReadWritable(shared.NewRedisReadWritable(ctx, 60*time.Minute)),
		shared.AlwaysCachable,
		shared.URLAsCacheKey,
		shared.CacheStatusOK).ServeHTTP(w, r)
}

func (h FeatureInteropHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	isExperimental := false
	val, _ := shared.ParseBooleanParam(q, "experimental")
	if val != nil {
		isExperimental = *val
	}

	mode := q.Get("mode")
	scope := q.Get("scope")
	tier := q.Get("tier")
	if mode != "" || scope != "" || tier != "" {
		if bundled, ok := shared.LoadBundledFeatureInteropData(isExperimental, mode, scope, tier); ok {
			marshalled, err := json.Marshal(bundled)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := w.Write(marshalled); err != nil {
				shared.GetLogger(ctx).Warningf("Failed to write data: %s", err.Error())
			}
			return
		}
	}

	// Two fetches, as the breakdown is published per date and the trendline is
	// what says which date is newest.
	trendline, breakdown, err := h.fetcher.Fetch(ctx, isExperimental)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	marshalled, err := json.Marshal(shared.ExtractFeatureInteropData(trendline, breakdown))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if _, err := w.Write(marshalled); err != nil {
		shared.GetLogger(ctx).Warningf("Failed to write data: %s", err.Error())
	}
}

// apiWebFeaturesManifestHandler serves the cached WEB_FEATURES_MANIFEST and
// WebStatus maturity tier mapping used by the results page.
func apiWebFeaturesManifestHandler(w http.ResponseWriter, r *http.Request) {
	raw, err := shared.LoadBundledWebFeaturesManifest()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(raw); err != nil {
		shared.GetLogger(r.Context()).Warningf("Failed to write manifest data: %s", err.Error())
	}
}
