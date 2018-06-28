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
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// This handler is responsible for all pages that display test results.
// It fetches the latest TestRun for each browser then renders the HTML
// page with the TestRuns encoded as JSON. The Polymer app picks those up
// and loads the summary files based on each entity's TestRun.ResultsURL.
//
// The browsers initially displayed to the user are defined in browsers.json.
// The JSON property "initially_loaded" is what controls this.
func testResultsHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect legacy paths.
	path := mux.Vars(r)["path"]
	var redir string
	if path == "results" {
		redir = "/results/"
	} else if strings.Index(r.URL.Path, "/results/") != 0 {
		redir = fmt.Sprintf("/results/%s", path)
	}
	if redir != "" {
		params := ""
		if r.URL.RawQuery != "" {
			params = "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, redir+params, http.StatusTemporaryRedirect)
		return
	}

	query := r.URL.Query()
	runSHA, err := shared.ParseSHAParam(r)
	if err != nil {
		http.Error(w, "Invalid query params", http.StatusBadRequest)
		return
	}

	var testRunSources []string
	var testRuns []shared.TestRun
	if testRunSources, testRuns, err = getTestRunsAndSources(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := struct {
		TestRuns       string
		TestRunSources string
		SHA            string
		Diff           bool
		Filter         string
		Labels         string
	}{
		SHA:    runSHA,
		Filter: r.URL.Query().Get("filter"),
	}

	labels := shared.ToStringSlice(shared.ParseLabelsParam(r))
	if labels != nil {
		var marshaled []byte
		if marshaled, err = json.Marshal(labels); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data.Labels = string(marshaled)
	}

	_, diff := query["diff"]
	data.Diff = diff || query.Get("before") != "" || query.Get("after") != ""

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

// getTestRunsAndSources gets source urls (api endpoints), and any placeholder TestRun entities from the parameters
// for the current request.
// When diffing, 'before' and 'after' parameters can be test-run specs (i.e. [product]@[sha]), or base64 encoded
// TestRun JSON blobs for the results summaries.
func getTestRunsAndSources(r *http.Request) (testRunSources []string, testRuns []shared.TestRun, err error) {
	before := r.URL.Query().Get("before")
	after := r.URL.Query().Get("after")
	if before != "" || after != "" {
		if before == "" {
			return nil, nil, errors.New("after param provided, but before param missing")
		} else if after == "" {
			return nil, nil, errors.New("before param provided, but after param missing")
		}

		const singleRunURL = `/api/run?sha=%s&product=%s`

		if beforeDecoded, err := base64.URLEncoding.DecodeString(before); err == nil {
			var run shared.TestRun
			if err = json.Unmarshal([]byte(beforeDecoded), &run); err != nil {
				return nil, nil, err
			}
			testRuns = append(testRuns, run)
		} else {
			beforeSpec, err := shared.ParseProductSpec(before)
			if err != nil {
				return nil, nil, errors.New("invalid before param")
			}
			testRunSources = append(testRunSources, fmt.Sprintf(singleRunURL, beforeSpec.Revision, beforeSpec.Product.String()))
		}

		if afterDecoded, err := base64.URLEncoding.DecodeString(after); err == nil {
			var run shared.TestRun
			if err = json.Unmarshal([]byte(afterDecoded), &run); err != nil {
				return nil, nil, err
			}
			testRuns = append(testRuns, run)
		} else {
			afterSpec, err := shared.ParseProductSpec(after)
			if err != nil {
				return nil, nil, errors.New("invalid after param")
			}
			testRunSources = append(testRunSources, fmt.Sprintf(singleRunURL, afterSpec.Revision, afterSpec.Product.String()))
		}
	} else {
		sourceURL, _ := url.Parse("/api/runs")
		f, err := shared.ParseTestRunFilterParams(r)
		if err != nil {
			return nil, nil, err
		}
		sourceURL.RawQuery = f.ToQuery(true).Encode()
		testRunSources = []string{sourceURL.String()}
	}
	return testRunSources, testRuns, nil
}
