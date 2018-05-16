// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"net/http"
)

func init() {
	// API endpoint for diff of two test run summary JSON blobs.
	http.HandleFunc("/api/diff", apiDiffHandler)

	// API endpoint for fetching a manifest for a commit SHA.
	http.HandleFunc("/api/manifest", apiManifestHandler)

	// API endpoint for listing all test runs for a given SHA.
	http.HandleFunc("/api/runs", apiTestRunsHandler)

	// API endpoint for a single test run.
	http.HandleFunc("/api/run", apiTestRunHandler)

	// API endpoint for redirecting to a run's summary JSON blob.
	http.HandleFunc("/api/results", apiResultsRedirectHandler)

	// API endpoint for receiving test results (wptreport) from runners.
	http.HandleFunc("/api/results/upload", apiResultsReceiveHandler)
}
