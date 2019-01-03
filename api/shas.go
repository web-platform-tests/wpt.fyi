// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// SHAsHandler is an http.Handler for the /api/shas endpoint.
type SHAsHandler struct {
	ctx context.Context
}

// apiSHAsHandler is responsible for emitting just the revision SHAs for test runs.
func apiSHAsHandler(w http.ResponseWriter, r *http.Request) {
	// Serve cached with 5 minute expiry. Delegate to SHAsHandler on cache miss.
	ctx := shared.NewAppEngineContext(r)
	shared.NewCachingHandler(ctx, SHAsHandler{ctx}, shared.NewGZReadWritable(shared.NewMemcacheReadWritable(ctx, 5*time.Minute)), shared.AlwaysCachable, shared.URLAsCacheKey, shared.CacheStatusOK).ServeHTTP(w, r)
}

func (h SHAsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters, err := shared.ParseTestRunFilterParams(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := h.ctx
	store := shared.NewAppEngineDatastore(ctx)

	var shas []string
	products := filters.GetProductsOrDefault()
	if filters.Aligned != nil && *filters.Aligned {
		if shas, _, err = shared.GetAlignedRunSHAs(store, products, filters.Labels, filters.From, filters.To, filters.MaxCount, filters.Offset); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		testRuns, err := store.LoadTestRuns(products, filters.Labels, nil, filters.From, filters.To, filters.MaxCount, filters.Offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		seen := mapset.NewSet()
		for _, testRun := range testRuns.AllRuns() {
			if !seen.Contains(testRun.Revision) {
				shas = append(shas, testRun.Revision)
				seen.Add(testRun.Revision)
			}
		}
	}
	if len(shas) < 1 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("[]"))
		return
	}

	shasBytes, err := json.Marshal(shas)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(shasBytes)
}
