// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import "github.com/web-platform-tests/wpt.fyi/shared"

// RegisterRoutes adds all the api route handlers.
func RegisterRoutes() {
	// API endpoint for diff of two test run summary JSON blobs.
	shared.AddRoute("/api/diff", "api-diff", shared.WrapPermissiveCORS(apiDiffHandler))

	// API endpoint for fetching interoperability metadata.
	shared.AddRoute("/api/interop", "api-interop", shared.WrapPermissiveCORS(apiInteropHandler))

	// API endpoint for fetching a manifest for a commit SHA.
	shared.AddRoute("/api/manifest", "api-manifest", shared.WrapPermissiveCORS(apiManifestHandler))

	// API endpoint for listing test runs for a set of products, labels, and SHA.
	shared.AddRoute("/api/runs", "api-test-runs", shared.WrapPermissiveCORS(apiTestRunsHandler))

	// API endpoints for a single test run, by its ID:
	shared.AddRoute("/api/runs/{id}", "api-test-run", shared.WrapPermissiveCORS(apiTestRunHandler))

	// API endpoint for listing SHAs for the test runs.
	shared.AddRoute("/api/shas", "api-shas", shared.WrapPermissiveCORS(apiSHAsHandler))

	// API endpoint for redirecting to a run's summary JSON blob.
	shared.AddRoute("/api/results", "api-results", shared.WrapPermissiveCORS(apiResultsRedirectHandler))

	// PROTECTED API endpoint for receiving test results (wptreport) from runners.
	// This API is authenticated. Runners have credentials.
	shared.AddRoute("/api/results/upload", "api-results-upload", apiResultsUploadHandler)

	// PRIVATE API endpoint for creating a test run in Datastore.
	// This API is authenticated. Only this AppEngine project has the credential.
	shared.AddRoute("/api/results/create", "api-results-create", apiResultsCreateHandler)
}
