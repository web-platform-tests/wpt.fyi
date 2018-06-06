// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
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
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// TODO(lukebjerring): Consolidate with shared.ParseProductParam & shared.LoadTestRuns
	product := params.Get("product")
	if product == "" {
		http.Error(w, "Param 'product' missing", http.StatusBadRequest)
		return
	}

	runSHA := params.Get("sha")
	if runSHA == "" {
		// Legacy name, in case still present in scripts/local stores.
		runSHA = params.Get("run")
	}
	if runSHA == "" {
		runSHA = "latest"
	}

	run, err := getRun(r, runSHA, product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if run == nil {
		http.Error(w, fmt.Sprintf("404 - Test run '%s' not found", runSHA), http.StatusNotFound)
		return
	}

	test := params.Get("test")
	resultsURL := getResultsURL(*run, test)

	http.Redirect(w, r, resultsURL, http.StatusFound)
}

func getRun(r *http.Request, run string, product string) (latest *shared.TestRun, err error) {
	productPieces := strings.Split(product, "-")
	if len(productPieces) < 1 || len(productPieces) > 4 {
		err = errors.New("Invalid path")
		return
	}

	ctx := appengine.NewContext(r)
	baseQuery := datastore.NewQuery("TestRun").Order("-CreatedAt").Limit(1)

	var testRunResults []shared.TestRun
	query := baseQuery.Filter("BrowserName =", productPieces[0])
	if run != "" && run != "latest" {
		query = query.Filter("Revision =", run)
	}
	if len(productPieces) > 1 {
		query = shared.VersionPrefix(query, "BrowserVersion", productPieces[1], true)
	}
	if len(productPieces) > 2 {
		query = query.Filter("OSName =", productPieces[2])
	}
	if len(productPieces) > 3 {
		query = shared.VersionPrefix(query, "OSVersion", productPieces[3], true)
	}
	keys, err := query.GetAll(ctx, &testRunResults)
	if err != nil {
		return
	}
	// Append the keys as ID
	for i, key := range keys {
		testRunResults[i].ID = key.IntID()
	}
	if len(testRunResults) > 0 {
		latest = &testRunResults[0]
	}
	return latest, err
}

func getResultsURL(run shared.TestRun, testFile string) (resultsURL string) {
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
