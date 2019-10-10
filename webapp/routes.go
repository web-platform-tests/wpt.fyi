// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate packr2

package webapp

import (
	"html/template"

	"github.com/gobuffalo/packr/v2"
	"github.com/web-platform-tests/wpt.fyi/shared"
	_ "google.golang.org/appengine/remote_api" // Registers the remote API in init
)

var templates *template.Template

func init() {
	box := packr.New("html templates", "./templates/")
	templates = template.New("all.html")
	for _, t := range box.List() {
		tmpl := templates.New(t)
		body, err := box.FindString(t)
		if err != nil {
			panic(err)
		} else if _, err = tmpl.Parse(body); err != nil {
			panic(err)
		}
	}
}

// RegisterRoutes adds the route handlers for the webapp.
func RegisterRoutes() {
	// About wpt.fyi
	shared.AddRoute("/about", "about", aboutHandler)

	// Reftest analyzer
	shared.AddRoute("/analyzer", "analyzer", analyzerHandler)

	// Feature flags for wpt.fyi
	shared.AddRoute("/flags", "flags", flagsHandler)
	shared.AddRoute("/components/wpt-env-flags.js", "flags-component", flagsComponentHandler)

	shared.AddRoute("/node_modules/{path:.*}", "components", componentsHandler)

	// Test run results, viewed by pass-rate across the browsers
	shared.AddRoute("/interop/", "interop", interopHandler)
	shared.AddRoute("/interop/{path:.*}", "interop", interopHandler)

	// A list of useful/insightful queries
	shared.AddRoute("/insights", "insights", insightsHandler)

	// List of all pending/in-flight runs
	shared.AddRoute("/status", "processor-status", processorStatusHandler)

	// List of all test runs, by SHA[0:10]
	shared.AddRoute("/runs", "test-runs", testRunsHandler)
	shared.AddRoute("/test-runs", "test-runs", testRunsHandler) // Legacy name

	shared.AddRoute("/service-worker.js", "service-worker", serviceWorkerHandler)

	// Admin-only manual results upload.
	shared.AddRoute("/admin/results/upload", "admin-results-upload", adminUploadHandler)

	// Admin-only manual cache flush.
	shared.AddRoute("/admin/cache/flush", "admin-cache-flush", adminCacheFlushHandler)

	// Admin-only environment flag management
	shared.AddRoute("/admin/flags", "admin-flags", adminFlagsHandler)

	// Test run results, viewed by browser (default view)
	// For run results diff view, 'before' and 'after' params can be given.
	shared.AddRoute("/results/", "results", testResultsHandler)
	shared.AddRoute("/results/{path:.*}", "results", testResultsHandler)

	// Legacy wildcard match
	shared.AddRoute("/", "results-legacy", testResultsHandler)
	shared.AddRoute("/{path:.*}", "results-legacy", testResultsHandler)
}
