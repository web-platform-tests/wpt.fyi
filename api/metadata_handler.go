// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// MetadataHandler is an http.Handler for the /api/metadata endpoint.
type MetadataHandler struct {
	ctx context.Context
}

// apiMetadataHandler searches Metadata for given products.
func apiMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Invalid HTTP method", http.StatusBadRequest)
		return
	}
	// Serve cached with 5 minute expiry. Delegate to MetadataHandler on cache miss.
	ctx := shared.NewAppEngineContext(r)
	shared.NewCachingHandler(
		ctx,
		MetadataHandler{ctx},
		shared.NewGZReadWritable(shared.NewMemcacheReadWritable(ctx, 5*time.Minute)),
		shared.AlwaysCachable,
		shared.URLAsCacheKey,
		shared.CacheStatusOK).ServeHTTP(w, r)
}

func (h MetadataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var productSpecs shared.ProductSpecs
	productSpecs, err := shared.ParseProductOrBrowserParams(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if len(productSpecs) == 0 {
		http.Error(w, fmt.Sprintf("Missing required 'product' param"), http.StatusBadRequest)
		return
	}

	ctx := h.ctx
	client := shared.NewAppEngineAPI(ctx).GetHTTPClient()
	logger := shared.GetLogger(ctx)

	MetadataResponse := shared.GetMetadataResponseOnProducts(productSpecs, client, logger)
	marshalled, err := json.Marshal(MetadataResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(marshalled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
