// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api //nolint:revive

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
	ctx context.Context // nolint:containedctx // TODO: Fix containedctx lint error
}

// apiSHAsHandler is responsible for emitting just the revision SHAs for test runs.
func apiSHAsHandler(w http.ResponseWriter, r *http.Request) {
	// Serve cached with 5 minute expiry. Delegate to SHAsHandler on cache miss.
	ctx := r.Context()
	shared.NewCachingHandler(
		ctx,
		SHAsHandler{ctx},
		shared.NewGZReadWritable(shared.NewRedisReadWritable(ctx, 5*time.Minute)),
		shared.AlwaysCachable,
		shared.URLAsCacheKey,
		shared.CacheStatusOK,
	).ServeHTTP(w, r)
}

func (h SHAsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters, err := shared.ParseTestRunFilterParams(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	ctx := h.ctx
	store := shared.NewAppEngineDatastore(ctx, true)
	q := store.TestRunQuery()

	var shas []string
	products := filters.GetProductsOrDefault()
	if filters.Aligned != nil && *filters.Aligned {
		shas, _, err = q.GetAlignedRunSHAs(
			products,
			filters.Labels,
			filters.From,
			filters.To,
			filters.MaxCount,
			filters.Offset,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	} else {
		testRuns, err := q.LoadTestRuns(
			products,
			filters.Labels,
			nil,
			filters.From,
			filters.To,
			filters.MaxCount,
			filters.Offset,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		seen := mapset.NewSet()
		for _, testRun := range testRuns.AllRuns() {
			// nolint:staticcheck // TODO: Fix staticcheck lint error (SA1019).
			if !seen.Contains(testRun.Revision) {
				shas = append(shas, testRun.Revision)
				seen.Add(testRun.Revision)
			}
		}
	}
	if len(shas) < 1 {
		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write([]byte("[]"))
		if err != nil {
			logger := shared.GetLogger(ctx)
			logger.Warningf("Failed to write data in api/shas handler: %s", err.Error())
		}

		return
	}

	shasBytes, err := json.Marshal(shas)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	_, err = w.Write(shasBytes)
	if err != nil {
		logger := shared.GetLogger(r.Context())
		logger.Warningf("Failed to write data in api/shas handler: %s", err.Error())
	}
}
