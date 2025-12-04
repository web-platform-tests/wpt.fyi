// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api //nolint:revive

import "github.com/web-platform-tests/wpt.fyi/shared"

// RegisterRoutes adds all the api route handlers.
func RegisterRoutes() {
	// API endpoint for diff of two test run summary JSON blobs.
	shared.AddRoute("/api/diff", "api-diff",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(apiDiffHandler)))

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
	shared.AddRoute("/api/status/{filter:pending|invalid|empty|duplicate}", "api-pending-test-runs", pendingTestRuns)

	// API endpoint for redirecting to a run's summary JSON blob.
	shared.AddRoute("/api/results", "api-results", shared.WrapPermissiveCORS(apiResultsRedirectHandler))

	// API endpoint for redirecting to a screenshot png blob.
	shared.AddRoute(
		"/api/screenshot/{screenshot:.*}",
		"api-screenshot",
		shared.WrapPermissiveCORS(apiScreenshotRedirectHandler),
	)

	// API endpoint for searching Metadata for the products.
	shared.AddRoute(
		"/api/metadata",
		"api-metadata",
		shared.WrapApplicationJSON(shared.WrapPermissiveCORS(apiMetadataHandler)),
	)

	// API endpoint for searching pending Metadata stored in memory.
	shared.AddRoute(
		"/api/metadata/pending",
		"api-pending-metadata",
		shared.WrapApplicationJSON(shared.WrapPermissiveCORS(apiPendingMetadataHandler)),
	)

	// API endpoint for modifying Metadata.
	shared.AddRoute(
		"/api/metadata/triage",
		"api-metadata-triage",
		shared.WrapTrustedCORS(apiMetadataTriageHandler, CORSList, []string{"PATCH"}),
	)

	// API endpoint for checking a user's login status.
	shared.AddRoute(
		"/api/user",
		"api-user",
		shared.WrapApplicationJSON(shared.WrapTrustedCORS(apiUserHandler, CORSList, nil)),
	)

	// API endpoint for fetching browser-specific failure data.
	shared.AddRoute(
		"/api/bsf",
		"api-bsf",
		shared.WrapApplicationJSON(shared.WrapPermissiveCORS(apiBSFHandler)),
	)

	// API endpoint for fetching historical data of a specific test for each of the four major browsers.
	shared.AddRoute("/api/history", "api-history",
		shared.WrapApplicationJSON(
			shared.WrapPermissiveCORS(testHistoryHandler)))
}
