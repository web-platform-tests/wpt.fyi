// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// BSFHandler is an http.Handler for the /api/bsf endpoint.
type BSFHandler struct {
	fetcher shared.FetchBSF
}

// apiBSFHandler fetches browser-specific failure data based on the URL params.
func apiBSFHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Serve cached with 5 minute expiry. Delegate to BSFHandler on cache miss.
	shared.NewCachingHandler(
		r.Context(),
		BSFHandler{shared.NewFetchBSF()},
		shared.NewGZReadWritable(shared.NewMemcacheReadWritable(ctx, 5*time.Minute)),
		shared.AlwaysCachable,
		shared.URLAsCacheKey,
		shared.CacheStatusOK).ServeHTTP(w, r)
}

func (b BSFHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	q := r.URL.Query()

	var from *time.Time
	if from, err = shared.ParseDateTimeParam(q, "from"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var to *time.Time
	if to, err = shared.ParseDateTimeParam(q, "to"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isExperimental := false
	val, _ := shared.ParseBooleanParam(q, "experimental")
	if val != nil {
		isExperimental = *val
	}

	lines, err := b.fetcher.Fetch(isExperimental)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bsfData := shared.FilterandExtractBSFData(lines, from, to)
	marshalled, err := json.Marshal(bsfData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(marshalled)
}
