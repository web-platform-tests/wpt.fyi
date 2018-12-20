// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// PRsHandler is an http.Handler for the /api/prs endpoint.
type PRsHandler struct {
	ctx context.Context
}

// apiPRsHandler executes a search for PRs for a particular directory.
func apiPRsHandler(w http.ResponseWriter, r *http.Request) {
	// Serve cached with 5 minute expiry. Delegate to PRsHandler on cache miss.
	ctx := shared.NewAppEngineContext(r)
	shared.NewCachingHandler(
		ctx,
		PRsHandler{ctx},
		shared.NewGZReadWritable(shared.NewMemcacheReadWritable(ctx, 5*time.Minute)),
		shared.AlwaysCachable,
		shared.URLAsCacheKey,
		shared.CacheStatusOK).ServeHTTP(w, r)
}

func (h PRsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	paths := shared.ParsePathsParam(q)
	ctx := h.ctx

	aeAPI := shared.NewAppEngineAPI(ctx)
	prs, err := getPRsByPaths(aeAPI, paths...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to run search: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	marshalled, err := json.Marshal(prs)
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

func getPRsByPaths(aeAPI shared.AppEngineAPI, paths ...string) ([]github.Issue, error) {
	client, err := aeAPI.GetGitHubClient()
	q := fmt.Sprintf("type:pr user:web-platform-tests repo:wpt state:open %s", strings.Join(paths, " "))
	prs, _, err := client.Search.Issues(aeAPI.Context(), q, &github.SearchOptions{
		Order: "desc",
		Sort:  "updated",
	})
	return prs.Issues, err
}
