// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// RegisterRoutes adds the route handlers for the webapp.
func RegisterRoutes() {
	// GitHub OAuth login
	shared.AddRoute("/login", "login", loginHandler)
	shared.AddRoute("/logout", "logout", logoutHandler)
	shared.AddRoute("/oauth", "oauth", oauthHandler)

	// About wpt.fyi
	shared.AddRoute("/about", "about", aboutHandler)

	// Reftest analyzer
	shared.AddRoute("/analyzer", "analyzer", analyzerHandler)

	// Feature flags for wpt.fyi
	shared.AddRoute("/flags", "flags", flagsHandler)
	shared.AddRoute("/dynamic-components/wpt-env-flags.js", "flags-component", flagsComponentHandler)

	shared.AddRoute("/node_modules/{path:.*}", "components", componentsHandler)

	// A list of useful/insightful queries
	shared.AddRoute("/insights", "insights", insightsHandler)

	// List of all pending/in-flight runs
	shared.AddRoute("/status", "processor-status", processorStatusHandler)

	// List of all test runs, by SHA[0:10]
	shared.AddRoute("/runs", "test-runs", testRunsHandler)
	shared.AddRoute("/test-runs", "test-runs", testRunsHandler) // Legacy name

	// Dashboard for the interop effort, by year.
	shared.AddRoute("/{name:(?:compat|interop-)}{year:[0-9]+}", "interop-dashboard", interopHandler)

	// Redirect to current year's interop effort.
	shared.AddRoute("/interop", "interop-dashboard", interopHandler)

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

	// Service readiness and liveness check handlers.
	shared.AddRoute("/_ah/liveness_check", "liveness-check", livenessCheckHandler)
	shared.AddRoute("/_ah/readiness_check", "readiness-check", readinessCheckHandler)

	// Legacy wildcard match
	shared.AddRoute("/", "results-legacy", testResultsHandler)
	shared.AddRoute("/{path:.*}", "results-legacy", testResultsHandler)
}

func livenessCheckHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Alive"))
	if err != nil {
		logger := shared.GetLogger(r.Context())
		logger.Warningf("Failed to write data in liveness check handler: %s", err.Error())
	}
}

func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Ready"))
	if err != nil {
		logger := shared.GetLogger(r.Context())
		logger.Warningf("Failed to write data in readiness check handler: %s", err.Error())
	}
}
