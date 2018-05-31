// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// apiTestRunsHandler is responsible for emitting test-run JSON for all the runs at a given SHA.
//
// URL Params:
//     sha: SHA[0:10] of the repo when the tests were executed (or 'latest')
func apiTestRunsHandler(w http.ResponseWriter, r *http.Request) {
	runSHA, err := shared.ParseSHAParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)
	// When ?complete=true, make sure to show results for the same complete run (executed for all browsers).
	if complete, err := strconv.ParseBool(r.URL.Query().Get("complete")); err == nil && complete {
		if runSHA == "latest" {
			runSHA, err = getLastCompleteRunSHA(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	var products []shared.Product
	if products, err = shared.GetProductsForRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var limit int
	if parsed, err := shared.ParseMaxCountParam(r); err != nil {
		http.Error(w, "Invalid 'max-count' param: "+err.Error(), http.StatusBadRequest)
		return
	} else if parsed != nil {
		limit = *parsed
	}
	var from *time.Time
	if from, err = shared.ParseFromParam(r); err != nil {
		http.Error(w, fmt.Sprintf("Invalid 'from' param: %s", err.Error()), http.StatusBadRequest)
		return
	}
	if limit == 0 && from == nil {
		// Default to a single, latest run when from & max-count both empty.
		limit = 1
	}

	testRuns, err := shared.LoadTestRuns(ctx, products, runSHA, from, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	testRunsBytes, err := json.Marshal(testRuns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(testRunsBytes)
}

// getLastCompleteRunSHA returns the SHA[0:10] for the most recent run that exists for all initially-loaded browser
// names (see GetBrowserNames).
func getLastCompleteRunSHA(ctx context.Context) (sha string, err error) {
	baseQuery := datastore.
		NewQuery("TestRun").
		Order("-CreatedAt").
		Limit(100).
		Project("Revision")

	// Map is sha -> browser -> seen yet?  - this prevents over-counting dupes.
	runSHAs := make(map[string]map[string]bool)
	var browserNames []string
	if browserNames, err = shared.GetBrowserNames(); err != nil {
		return sha, err
	}

	for _, browser := range browserNames {
		it := baseQuery.Filter("BrowserName = ", browser).Run(ctx)
		for {
			var testRun shared.TestRun
			_, err := it.Next(&testRun)
			if err == datastore.Done {
				break
			}
			if err != nil {
				return "latest", err
			}
			if _, ok := runSHAs[testRun.Revision]; !ok {
				runSHAs[testRun.Revision] = make(map[string]bool)
			}
			browsersSeen := runSHAs[testRun.Revision]
			browsersSeen[browser] = true
			if len(browsersSeen) == len(browserNames) {
				return testRun.Revision, nil
			}
		}
	}
	return "latest", nil
}
