// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
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

	var browserNames []string
	if browserNames, err = shared.ParseBrowsersParam(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	labels := shared.ParseLabelsParam(r)
	experimentalBrowsers := labels != nil && labels.Contains(shared.ExperimentalLabel)

	var testRuns [][]shared.TestRun
	var limit int
	if limit, err = shared.ParseMaxCountParam(r); err != nil {
		http.Error(w, "Invalid 'max-count' param: "+err.Error(), http.StatusBadRequest)
		return
	}
	baseQuery := datastore.
		NewQuery("TestRun").
		Order("-CreatedAt").
		Limit(limit)

	for _, browserName := range browserNames {
		var testRunResults []shared.TestRun
		if experimentalBrowsers {
			browserName = browserName + "-" + shared.ExperimentalLabel
		}
		query := baseQuery.Filter("BrowserName =", browserName)
		if runSHA != "" && runSHA != "latest" {
			query = query.Filter("Revision =", runSHA)
		}
		if _, err := query.GetAll(ctx, &testRunResults); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		testRuns = append(testRuns, testRunResults)
	}

	var allRuns []shared.TestRun
	if limit > 1 {
		// Crop at `limit` runs - whichever is the youngest `limit`th run.
		youngestRun := time.Unix(0, 0)
		for _, runs := range testRuns {
			if len(runs) == limit && runs[limit-1].CreatedAt.After(youngestRun) {
				youngestRun = runs[limit-1].CreatedAt
			}
		}
		for i, runs := range testRuns {
			for len(runs) > 0 && runs[len(runs)-1].CreatedAt.Before(youngestRun) {
				runs = runs[0 : len(runs)-1]
			}
			testRuns[i] = runs
		}
	}
	for _, runs := range testRuns {
		allRuns = append(allRuns, runs...)
	}

	testRunsBytes, err := json.Marshal(allRuns)
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
