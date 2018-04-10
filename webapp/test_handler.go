// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	models "github.com/w3c/wptdashboard/shared"
)

// This handler is responsible for all pages that display test results.
// It fetches the latest TestRun for each browser then renders the HTML
// page with the TestRuns encoded as JSON. The Polymer app picks those up
// and loads the summary files based on each entity's TestRun.ResultsURL.
//
// The browsers initially displayed to the user are defined in browsers.json.
// The JSON property "initially_loaded" is what controls this.
func testHandler(w http.ResponseWriter, r *http.Request) {
	runSHA, err := ParseSHAParam(r)
	if err != nil {
		http.Error(w, "Invalid query params", http.StatusBadRequest)
		return
	}

	var testRunSources []string
	var testRuns []models.TestRun
	if testRunSources, testRuns, err = getTestRunsAndSources(r, runSHA); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := struct {
		TestRuns       string
		TestRunSources string
		SHA            string
	}{
		SHA: runSHA,
	}

	// Run source URLs
	if testRunSources != nil && len(testRunSources) > 0 {
		var marshaled []byte
		if marshaled, err = json.Marshal(testRunSources); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data.TestRunSources = string(marshaled)
	}

	// Runs by base64-encoded param or spec param.
	if testRuns != nil && len(testRuns) > 0 {
		var marshaled []byte
		if marshaled, err = json.Marshal(testRuns); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data.TestRuns = string(marshaled)
	}

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getTestRunsAndSources gets the arrays of source urls and placeholder-TestRun-models from the parameters for the
// current request. When diffing, 'before' and 'after' parameters can be test-run specs (i.e. [platform]@[sha]), or
// base64 encoded TestRun JSON blobs for the results summaries.
func getTestRunsAndSources(r *http.Request, runSHA string) (testRunSources []string, testRuns []models.TestRun, err error) {
	before := r.URL.Query().Get("before")
	after := r.URL.Query().Get("after")
	filter := r.URL.Query().Get("filter")
	if before != "" || after != "" {
		if before == "" {
			return nil, nil, errors.New("after param provided, but before param missing")
		} else if after == "" {
			return nil, nil, errors.New("before param provided, but after param missing")
		}

		const singleRunURL = `/api/run?sha=%s&browser=%s`

		if beforeDecoded, err := base64.URLEncoding.DecodeString(before); err == nil {
			var run models.TestRun
			if err = json.Unmarshal([]byte(beforeDecoded), &run); err != nil {
				return nil, nil, err
			}
			testRuns = append(testRuns, run)
		} else {
			var beforeSpec platformAtRevision
			if beforeSpec, err = parsePlatformAtRevisionSpec(before); err != nil {
				return nil, nil, errors.New("invalid before param")
			}
			testRunSources = append(testRunSources, fmt.Sprintf(singleRunURL, beforeSpec.Revision, beforeSpec.Platform))
		}

		if afterDecoded, err := base64.URLEncoding.DecodeString(after); err == nil {
			var run models.TestRun
			if err = json.Unmarshal([]byte(afterDecoded), &run); err != nil {
				return nil, nil, err
			}
			testRuns = append(testRuns, run)
		} else {
			var afterSpec platformAtRevision
			if afterSpec, err = parsePlatformAtRevisionSpec(after); err != nil {
				return nil, nil, errors.New("invalid after param")
			}
			testRunSources = append(testRunSources, fmt.Sprintf(singleRunURL, afterSpec.Revision, afterSpec.Platform))
		}
	} else {
		const sourceURL = `/api/runs?sha=%s`
		testRunSources = []string{fmt.Sprintf(sourceURL, runSHA)}
	}

	if before != "" || after != "" {
		const diffRunURL = `/api/diff?before=%s&after=%s`
		resultsURL := fmt.Sprintf(diffRunURL, before, after)
		if filter == "" {
			filter = "ACDU" // Added, Changed, Deleted, Unchanged
		}
		resultsURL += "&filter=" + filter
		diffRun := models.TestRun{
			Revision:    "diff",
			BrowserName: "diff",
			ResultsURL:  resultsURL,
		}
		testRuns = append(testRuns, diffRun)
	}
	return testRunSources, testRuns, nil
}
