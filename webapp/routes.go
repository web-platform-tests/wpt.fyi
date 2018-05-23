// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"html/template"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func init() {
	// Test run results, viewed by browser (default view)
	// For run results diff view, 'before' and 'after' params can be given.
	shared.AddRoute("/", testResultsHandler)
	shared.AddRoute("/results", testResultsHandler) // Prevent default redirect
	shared.AddRoute("/results/", testResultsHandler)

	// About wpt.fyi
	shared.AddRoute("/about", aboutHandler)

	// Test run results, viewed by pass-rate across the browsers
	shared.AddRoute("/interop/", interopHandler)

	// Lists of test run results which have poor interoperability
	shared.AddRoute("/interop/anomalies", anomalyHandler)

	// List of all test runs, by SHA[0:10]
	shared.AddRoute("/test-runs", testRunsHandler)

	// Admin-only manual results upload.
	shared.AddRoute("/admin/results/upload", adminUploadHandler)
}
