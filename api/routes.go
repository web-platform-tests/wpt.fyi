// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import "github.com/web-platform-tests/wpt.fyi/shared"

func init() {
	// API endpoint for diff of two test run summary JSON blobs.
	shared.AddRoute("/api/diff", apiDiffHandler)

	// API endpoint for fetching a manifest for a commit SHA.
	shared.AddRoute("/api/manifest", apiManifestHandler)

	// API endpoint for listing all test runs for a given SHA.
	shared.AddRoute("/api/runs", apiTestRunsHandler)

	// API endpoint for a single test run.
	shared.AddRoute("/api/run", apiTestRunHandler)

	// API endpoint for redirecting to a run's summary JSON blob.
	shared.AddRoute("/api/results", apiResultsRedirectHandler)

	// API endpoint for receiving test results (wptreport) from runners.
	shared.AddRoute("/api/results/upload", apiResultsReceiveHandler)
}
