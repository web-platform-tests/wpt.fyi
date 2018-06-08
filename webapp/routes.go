// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"html/template"

	"github.com/web-platform-tests/wpt.fyi/api"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func init() {
	// webapp.RegisterRoutes has a catch-all, so needs to go last.
	api.RegisterRoutes()
	RegisterRoutes()
}

// RegisterRoutes adds the route handlers for the webapp.
func RegisterRoutes() {
	// About wpt.fyi
	shared.AddRoute("/about", "about", aboutHandler)

	// Lists of test run results which have poor interoperability
	shared.AddRoute("/anomalies", "anomaly", anomalyHandler)

	// Test run results, viewed by pass-rate across the browsers
	shared.AddRoute("/interop/", "interop", interopHandler)
	shared.AddRoute("/interop/{path:.*}", "interop", interopHandler)

	// List of all test runs, by SHA[0:10]
	shared.AddRoute("/test-runs", "test-runs", testRunsHandler)

	// Admin-only manual results upload.
	shared.AddRoute("/admin/results/upload", "admin-results-upload", adminUploadHandler)

	// Test run results, viewed by browser (default view)
	// For run results diff view, 'before' and 'after' params can be given.
	shared.AddRoute("/results/", "results", testResultsHandler)
	shared.AddRoute("/results/{path:.*}", "results", testResultsHandler)

	// Legacy wildcard match
	shared.AddRoute("/", "results-legacy", testResultsHandler)
	shared.AddRoute("/{path:.*}", "results-legacy", testResultsHandler)
}
