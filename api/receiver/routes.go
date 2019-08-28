// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import "github.com/web-platform-tests/wpt.fyi/shared"

// RegisterRoutes adds all the result receiver route handlers.
func RegisterRoutes() {
	// PROTECTED API endpoint for receiving test results (wptreport) from runners.
	// This API is authenticated. Runners have credentials.
	shared.AddRoute("/api/results/upload", "api-results-upload", apiResultsUploadHandler)

	// PRIVATE API endpoint for creating a test run in Datastore.
	// This API is authenticated. Only this AppEngine project has the credential.
	shared.AddRoute("/api/results/create", "api-results-create", apiResultsCreateHandler)

	// PRIVATE API endpoint for updating the status of a pending test run
	shared.AddRoute("/api/status/{id:[0-9]+}", "api-pending-test-run-update", apiPendingTestRunUpdateHandler)
}
