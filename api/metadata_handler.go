// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// MetadataHandler is an http.Handler for /api/metadata endpoint.
type MetadataHandler struct {
	logger      shared.Logger
	httpClient  *http.Client
	metadataURL string
}

// apiMetadataHandler searches Metadata for given products.
func apiMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" && r.Method != "PATCH" {
		http.Error(w, "Invalid HTTP method", http.StatusBadRequest)
		return
	}

	if r.Method == "PATCH" {
		apiMetadataTriageHandler(w, r)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	client := shared.NewAppEngineAPI(ctx).GetHTTPClient()
	logger := shared.GetLogger(ctx)
	metadataURL := shared.MetadataArchiveURL
	delegate := MetadataHandler{logger, client, metadataURL}

	// Serve cached with 5 minute expiry. Delegate to Metadata Handler on cache miss.
	shared.NewCachingHandler(
		ctx,
		delegate,
		shared.NewGZReadWritable(shared.NewMemcacheReadWritable(ctx, 5*time.Minute)),
		shared.AlwaysCachable,
		cacheKey,
		shared.CacheStatusOK).ServeHTTP(w, r)
}

func apiMetadataTriageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	client := shared.NewAppEngineAPI(ctx).GetHTTPClient()
	logger := shared.GetLogger(ctx)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read PATCH request body", http.StatusInternalServerError)
	}
	err = r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to finish reading request body", http.StatusInternalServerError)
	}

	var metadata shared.MetadataResults
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		http.Error(w, "Failed to parse JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Process it to a digestable method and send it to Github
	// Verify cookies
	// Verify users access

	w.WriteHeader(http.StatusCreated)
}

func (h MetadataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var abstractLink query.AbstractLink
	if r.Method == "POST" {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		}
		err = r.Body.Close()
		if err != nil {
			http.Error(w, "Failed to finish reading request body", http.StatusInternalServerError)
		}

		var ae query.AbstractExists
		err = json.Unmarshal(data, &ae)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var isLinkQuery = false
		if len(ae.Args) == 1 {
			abstractLink, isLinkQuery = ae.Args[0].(query.AbstractLink)
		}

		if !isLinkQuery {
			h.logger.Errorf("Error from request: non Link search query %s for api/metadata", ae)
			http.Error(w, "Error from request: non Link search query for api/metadata", http.StatusBadRequest)
			return
		}
	}

	q := r.URL.Query()
	productSpecs, err := shared.ParseProductOrBrowserParams(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if len(productSpecs) == 0 {
		http.Error(w, fmt.Sprintf("Missing required 'product' param"), http.StatusBadRequest)
		return
	}

	metadataResponse, err := shared.GetMetadataResponseOnProducts(productSpecs, h.httpClient, h.logger, h.metadataURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.Method == "POST" {
		metadataResponse = filterMetadata(abstractLink, metadataResponse)
	}
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

// TODO(kyleju): Refactor this part to shared package.
var cacheKey = func(r *http.Request) interface{} {
	if r.Method == "GET" {
		return shared.URLAsCacheKey(r)
	}

	body := r.Body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Failed to read non-GET request body for generating cache key: %v", err)
		shared.GetLogger(shared.NewAppEngineContext(r)).Errorf(msg)
		panic(msg)
	}
	defer body.Close()

	// Ensure that r.Body can be read again by other request handling routines.
	r.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	return fmt.Sprintf("%s#%s", r.URL.String(), string(data))
}
