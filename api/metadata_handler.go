// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// MetadataHandler is an http.Handler for /api/metadata endpoint.
type MetadataHandler struct {
	logger  shared.Logger
	fetcher shared.MetadataFetcher
}

// apiMetadataHandler searches Metadata for given products.
func apiMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Invalid HTTP method", http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	logger := shared.GetLogger(ctx)
	client := aeAPI.GetHTTPClient()
	gitHubClient, err := shared.NewAppEngineAPI(ctx).GetGitHubClient()
	if err != nil {
		logger.Errorf("Unable to get GitHub client: %e", err)
		http.Error(w, "Unable to get GitHub client", http.StatusInternalServerError)
		return
	}

	fetcher := webappMetadataFetcher{ctx: ctx, httpClient: client, gitHubClient: gitHubClient}
	MetadataHandler{logger, fetcher}.ServeHTTP(w, r)
}

func apiMetadataTriageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	ds := shared.NewAppEngineDatastore(ctx, false)

	user, token := shared.GetUserFromCookie(ctx, ds, r)
	if user == nil {
		http.Error(w, "User is not logged in", http.StatusUnauthorized)
		return
	}

	aeAPI := shared.NewAppEngineAPI(ctx)
	logger := shared.GetLogger(ctx)
	botClient, err := aeAPI.GetGitHubClient()
	if err != nil {
		logger.Errorf("Unable to get GitHub client: %e", err)
		http.Error(w, "Unable to get GitHub client", http.StatusInternalServerError)
		return
	}

	fetcher := webappMetadataFetcher{
		ctx:          ctx,
		httpClient:   aeAPI.GetHTTPClient(),
		gitHubClient: botClient,
		forceUpdate:  true}
	tm := shared.NewTriageMetadata(ctx, botClient, user.GitHubHandle, user.GitHubEmail, fetcher)

	gac, err := shared.NewGitHubAccessControl(ctx, ds, botClient, user, token)
	if err != nil {
		http.Error(w, "Unable to create GitHub OAuth client: "+err.Error(), http.StatusInternalServerError)
		return
	}
	handleMetadataTriage(ctx, gac, tm, w, r)
}

func handleMetadataTriage(ctx context.Context, gac shared.GitHubAccessControl, tm shared.TriageMetadata, w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Invalid HTTP method; only accept PATCH", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Invalid content-type: %s"+contentType, http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read PATCH request body", http.StatusInternalServerError)
		return
	}

	err = r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to finish reading request body", http.StatusInternalServerError)
		return
	}

	var metadata shared.MetadataResults
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		http.Error(w, "Failed to parse JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	valid, err := gac.IsValidWPTMember()
	if err != nil {
		http.Error(w, "Failed to validate web-platform-tests membership: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Logged-in user must be a member of the web-platform-tests GitHub organization. To join, please contact wpt.fyi team members.", http.StatusForbidden)
		return
	}

	// TODO(kyleju): Check github client permission levels for auto merge.
	pr, err := tm.Triage(metadata)
	if err != nil {
		http.Error(w, "Unable to triage metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(pr))
}

func (h MetadataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var abstractLink query.AbstractLink
	if r.Method == "POST" {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		err = r.Body.Close()
		if err != nil {
			http.Error(w, "Failed to finish reading request body", http.StatusInternalServerError)
			return
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

	metadataResponse, err := shared.GetMetadataResponseOnProducts(productSpecs, h.logger, h.fetcher)
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

// filterMetadata filters the given metadata down to entries where the value (links) contain
// at least one link where the URL contains the substring provided in the "link" search atom.
func filterMetadata(linkQuery query.AbstractLink, metadata shared.MetadataResults) shared.MetadataResults {
	res := make(shared.MetadataResults)
	for test, links := range metadata {
		for _, link := range links {
			if strings.Contains(link.URL, linkQuery.Pattern) {
				res[test] = links
				break
			}
		}
	}
	return res
}
