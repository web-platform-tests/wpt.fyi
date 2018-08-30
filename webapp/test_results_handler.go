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
	"strings"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

type testRunUIFilter struct {
	Products      string
	Labels        string
	SHA           string
	Complete      bool
	MaxCount      *int
	BeforeSpec    *shared.ProductSpec
	AfterSpec     *shared.ProductSpec
	Diff          bool
	BeforeTestRun *shared.TestRun
	AfterTestRun  *shared.TestRun
}

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

	filter, err := parseTestRunUIFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := struct {
		// TestRuns are inlined marshalled JSON for arbitrary test runs (e.g. when
		// diffing), for runs which aren't fetchable via a URL or the api.
		TestRuns string
		Filter   testRunUIFilter
	}{
		Filter: filter,
	}

	// Runs by base64-encoded param or spec param.
	if filter.BeforeTestRun != nil && filter.AfterTestRun != nil {
		runs := []shared.TestRun{
			*(filter.BeforeTestRun),
			*(filter.AfterTestRun),
		}
		filter.Diff = true

		var marshaled []byte
		if marshaled, err = json.Marshal(runs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data.TestRuns = string(marshaled)
	}

	if err := templates.ExecuteTemplate(w, "results.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// parseTestRunUIFilter parses the standard TestRunFilter, as well as the extra
// diff params (diff, before, after).
func parseTestRunUIFilter(r *http.Request) (filter testRunUIFilter, err error) {
	testRunFilter, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		return filter, err
	}

	before := r.URL.Query().Get("before")
	after := r.URL.Query().Get("after")
	if before != "" || after != "" {
		if before == "" {
			return filter, errors.New("after param provided, but before param missing")
		} else if after == "" {
			return filter, errors.New("before param provided, but after param missing")
		}

		const singleRunURL = `/api/run?sha=%s&product=%s`

		if beforeDecoded, err := base64.URLEncoding.DecodeString(before); err == nil {
			var run shared.TestRun
			if err = json.Unmarshal([]byte(beforeDecoded), &run); err != nil {
				return filter, err
			}
			filter.BeforeTestRun = &run
		} else {
			beforeSpec, err := shared.ParseProductSpec(before)
			if err != nil {
				return filter, errors.New("invalid before param")
			}
			filter.BeforeSpec = &beforeSpec
		}

		if afterDecoded, err := base64.URLEncoding.DecodeString(after); err == nil {
			var run shared.TestRun
			if err = json.Unmarshal([]byte(afterDecoded), &run); err != nil {
				return filter, err
			}
			filter.AfterTestRun = &run
		} else {
			afterSpec, err := shared.ParseProductSpec(after)
			if err != nil {
				return filter, errors.New("invalid after param")
			}
			filter.AfterSpec = &afterSpec
		}
	}

	if testRunFilter.Labels != nil {
		data, _ := json.Marshal(testRunFilter.Labels.ToSlice())
		filter.Labels = string(data)
	}
	if !shared.IsLatest(testRunFilter.SHA) {
		filter.SHA = testRunFilter.SHA
	}
	if !testRunFilter.IsDefaultProducts() {
		data, _ := json.Marshal(testRunFilter.Products.Strings())
		filter.Products = string(data)
	}
	filter.MaxCount = testRunFilter.MaxCount
	filter.Complete = testRunFilter.Complete != nil && *testRunFilter.Complete

	diff, err := shared.ParseBooleanParam(r, "diff")
	if err != nil {
		return filter, err
	}
	filter.Diff = diff != nil && *diff
	return filter, nil
}
