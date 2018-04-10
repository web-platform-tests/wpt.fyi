// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"html/template"
	"net/http"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func init() {
	// Test run results, viewed by browser (default view)
	// For run results diff view, 'before' and 'after' params can be given.
	http.HandleFunc("/", testHandler)

	// About wpt.fyi
	http.HandleFunc("/about", aboutHandler)

	// Test run results, viewed by pass-rate across the browsers
	http.HandleFunc("/interop/", interopHandler)

	// Lists of test run results which have poor interoperability
	http.HandleFunc("/interop/anomalies", anomalyHandler)

	// List of all test runs, by SHA[0:10]
	http.HandleFunc("/test-runs", testRunsHandler)

	// API endpoint for diff of two test run summary JSON blobs.
	http.HandleFunc("/api/diff", apiDiffHandler)

	// API endpoint for listing all test runs for a given SHA.
	http.HandleFunc("/api/runs", apiTestRunsHandler)

	// API endpoint for a single test run.
	http.HandleFunc("/api/run", apiTestRunHandler)

	// API endpoint for redirecting to a run's summary JSON blob.
	http.HandleFunc("/results", resultsRedirectHandler)
}
