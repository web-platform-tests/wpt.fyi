// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import "github.com/web-platform-tests/wpt.fyi/shared"

// RegisterRoutes adds all the api route handlers.
func RegisterRoutes() {
	// API endpoint for diff of two test run summary JSON blobs.
	shared.AddRoute("/api/diff", "api-diff",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(apiDiffHandler)))

	// API endpoint for fetching interoperability metadata.
	shared.AddRoute("/api/interop", "api-interop",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(apiInteropHandler)))

	// API endpoint for fetching all labels.
	shared.AddRoute("/api/labels", "api-labels",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(apiLabelsHandler)))

	// API endpoint for fetching a manifest for a commit SHA.
	shared.AddRoute("/api/manifest", "api-manifest",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(apiManifestHandler)))

	// API endpoint for listing all test runs for a given SHA.
	shared.AddRoute("/api/runs", "api-test-runs",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(apiTestRunsHandler)))

	// API endpoint for listing SHAs for the test runs.
	shared.AddRoute("/api/shas", "api-shas",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(apiSHAsHandler)))

	// API endpoint for searching PRs for the test runs.
	shared.AddRoute("/api/prs", "api-prs",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(apiPRsHandler)))

	// API endpoint for listing SHAs for the test runs.
	shared.AddRoute("/api/versions", "api-versions",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(apiVersionsHandler)))

	// API endpoints for a single test run, by
	// ID:
	shared.AddRoute("/api/runs/{id}", "api-test-run",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(apiTestRunHandler)))
	// 'product' param & 'sha' param:
	shared.AddRoute("/api/run", "api-test-run",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(apiTestRunHandler)))

	// API endpoint for listing pending test runs
	pendingTestRuns := shared.WrapApplicationJSON(
		shared.WrapPermissiveCORS(apiPendingTestRunsHandler))
	shared.AddRoute("/api/status", "api-pending-test-runs", pendingTestRuns)
	shared.AddRoute("/api/status/{filter:pending}", "api-pending-test-runs", pendingTestRuns)

	// API endpoint for redirecting to a run's summary JSON blob.
	shared.AddRoute("/api/results", "api-results", shared.WrapPermissiveCORS(apiResultsRedirectHandler))

	// API endpoint for redirecting to a screenshot png blob.
	shared.AddRoute("/api/screenshot/{screenshot:.*}", "api-screenshot", shared.WrapPermissiveCORS(apiScreenshotRedirectHandler))

	// API endpoint for searching Metadata for the products.
	shared.AddRoute("/api/metadata", "api-metadata", shared.WrapPermissiveCORS(apiMetadataHandler))
}
