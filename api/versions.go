// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api //nolint:revive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// VersionsHandler is an http.Handler for the /api/versions endpoint.
type VersionsHandler struct {
	ctx context.Context // nolint:containedctx // TODO: Fix containedctx lint error
}

// apiVersionsHandler is responsible for emitting just the browser versions for the test runs.
func apiVersionsHandler(w http.ResponseWriter, r *http.Request) {
	// Serve cached with 5 minute expiry. Delegate to VersionsHandler on cache
	// miss.
	ctx := r.Context()
	shared.NewCachingHandler(
		ctx,
		VersionsHandler{ctx},
		shared.NewGZReadWritable(shared.NewRedisReadWritable(ctx, 5*time.Minute)),
		shared.AlwaysCachable,
		shared.URLAsCacheKey,
		shared.CacheStatusOK,
	).ServeHTTP(w, r)
}

func (h VersionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	product, err := shared.ParseProductParam(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	} else if product == nil {
		http.Error(w, fmt.Sprintf("Invalid product param: %s", r.URL.Query().Get("product")), http.StatusBadRequest)

		return
	}

	ctx := h.ctx
	store := shared.NewAppEngineDatastore(ctx, false)
	query := store.NewQuery("TestRun").Filter("BrowserName =", product.BrowserName)
	if product.Labels != nil {
		for label := range product.Labels.Iter() {
			query = query.Filter("Labels =", label)
		}
	}
	distinctQuery := query.Project("BrowserVersion").Distinct()
	var queries []shared.Query
	if product.BrowserVersion == "" {
		queries = []shared.Query{distinctQuery}
	} else {
		queries = []shared.Query{
			query.Filter("BrowserVersion =", product.BrowserVersion).Limit(1),
			shared.VersionPrefix(distinctQuery, "BrowserVersion", product.BrowserVersion, false /*desc*/),
		}
	}

	var runs shared.TestRuns
	for _, query := range queries {
		var someRuns shared.TestRuns
		if _, err := store.GetAll(query, &someRuns); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
		runs = append(runs, someRuns...)
	}

	if len(runs) < 1 {
		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write([]byte("[]"))
		if err != nil {
			logger := shared.GetLogger(ctx)
			logger.Warningf("Failed to write data in api/versions handler: %s", err.Error())
		}

		return
	}

	versions := make([]string, len(runs))
	for i := range runs {
		versions[i] = runs[i].BrowserVersion
	}
	// nolint:godox // TODO(lukebjerring): Fix this, it will put 100 before 11..., etc.
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))

	versionsBytes, err := json.Marshal(versions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	_, err = w.Write(versionsBytes)
	if err != nil {
		logger := shared.GetLogger(ctx)
		logger.Warningf("Failed to write data in api/versions handler: %s", err.Error())
	}
}
