// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// MetadataHandler is an http.Handler for GET method of the /api/metadata endpoint.
type MetadataHandler struct {
	logger     shared.Logger
	httpClient *http.Client
}

// MetadataSearchHandler is an http.Handler for POST method of the /api/metadata endpoint.
type MetadataSearchHandler struct {
	logger     shared.Logger
	httpClient *http.Client
}

// apiMetadataHandler searches Metadata for given products.
func apiMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Invalid HTTP method", http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	client := shared.NewAppEngineAPI(ctx).GetHTTPClient()
	logger := shared.GetLogger(ctx)

	var delegate http.Handler
	if r.Method == "GET" {
		delegate = MetadataHandler{logger, client}
	} else {
		delegate = MetadataSearchHandler{logger, client}
	}

	// Serve cached with 5 minute expiry. Delegate to MetadataHandler on cache miss.
	shared.NewCachingHandler(
		ctx,
		delegate,
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

	MetadataResponse := shared.GetMetadataResponseOnProducts(productSpecs, h.httpClient, h.logger)
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

func (h MetadataSearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
	}
	err = r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to finish reading request body", http.StatusInternalServerError)
	}

	var rq query.RunQuery
	err = json.Unmarshal(data, &rq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var abstractLink query.AbstractLink
	var isLinkQuery = false
	if exists, isExists := rq.AbstractQuery.(query.AbstractExists); isExists && len(exists.Args) == 1 {
		abstractLink, isLinkQuery = exists.Args[0].(query.AbstractLink)
	}

	if !isLinkQuery {
		h.logger.Errorf("Error from request: non Link search query %s for api/metadata", rq.AbstractQuery)
		http.Error(w, "Error from request: non Link search query for api/metadata", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	var productSpecs shared.ProductSpecs
	productSpecs, err = shared.ParseProductOrBrowserParams(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if len(productSpecs) == 0 {
		http.Error(w, fmt.Sprintf("Missing required 'product' param"), http.StatusBadRequest)
		return
	}

	metadata := shared.GetMetadataResponseOnProducts(productSpecs, h.httpClient, h.logger)
	metadataResponse := filterMetadata(abstractLink, metadata)
	marshalled, err := json.Marshal(metadataResponse)
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

func filterMetadata(linkQuery query.AbstractLink, metadata shared.MetadataResults) shared.MetadataResults {
	res := shared.MetadataResults{}

	for _, data := range metadata {
		for _, url := range data.URLs {
			if strings.Contains(url, linkQuery.Pattern) {
				res = append(res, data)
				break
			}
		}
	}
	return res
}
