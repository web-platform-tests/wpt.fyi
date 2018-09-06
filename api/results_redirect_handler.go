// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// apiResultsRedirectHandler is responsible for redirecting to the Google Cloud Storage API
// JSON blob for the given SHA (or latest) shared.TestRun for the given browser.
//
// URL format:
// /results
//
// Params:
//   product: Browser (and OS) of the run, e.g. "chrome-63.0" or "safari"
//   (optional) run: SHA[0:10] of the test run, or "latest" (latest is the default)
//   (optional) test: Path of the test, e.g. "/css/css-images-3/gradient-button.html"
func apiResultsRedirectHandler(w http.ResponseWriter, r *http.Request) {
	filters, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	one := 1
	testRuns, err := shared.LoadTestRuns(ctx, filters.Products, filters.Labels, filters.SHA, nil, nil, &one)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(testRuns) == 0 {
		http.Error(w, fmt.Sprintf("404 - Test run '%s' not found", filters.SHA), http.StatusNotFound)
		return
	}

	test := r.URL.Query().Get("test")
	resultsURL := GetResultsURL(testRuns[0], test)

	http.Redirect(w, r, resultsURL, http.StatusFound)
}

// GetResultsURL constructs the URL to the result of a single test file in the given run.
func GetResultsURL(run shared.TestRun, testFile string) (resultsURL string) {
	resultsURL = run.ResultsURL
	if testFile != "" && testFile != "/" {
		// Assumes that result files are under a directory named SHA[0:10].
		resultsBase := strings.SplitAfter(resultsURL, "/"+run.Revision)[0]
		resultsPieces := strings.Split(resultsURL, "/")
		re := regexp.MustCompile("(-summary)?\\.json\\.gz$")
		product := re.ReplaceAllString(resultsPieces[len(resultsPieces)-1], "")
		resultsURL = fmt.Sprintf("%s/%s/%s", resultsBase, product, testFile)
	}
	return resultsURL
}
