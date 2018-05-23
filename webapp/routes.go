// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"html/template"
	"net/http"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func init() {
	routes := map[string]http.HandlerFunc{
		// Test run results, viewed by browser (default view)
		// For run results diff view, 'before' and 'after' params can be given.
		"/": testResultsHandler,
		"/results": testResultsHandler, // Prevent default redirect
		"/results/": testResultsHandler,

		// About wpt.fyi
		"/about": aboutHandler,

		// Test run results, viewed by pass-rate across the browsers
		"/interop/": interopHandler,

		// Lists of test run results which have poor interoperability
		"/interop/anomalies": anomalyHandler,

		// List of all test runs, by SHA[0:10]
		"/test-runs": testRunsHandler,

		// Admin-only manual results upload.
		"/admin/results/upload": adminUploadHandler,
	}

	for route, handler := range routes {
		http.HandleFunc(route, wrapHSTS(handler))
	}
}

func wrapHSTS(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := "max-age=31536000; preload"
		w.Header().Add("Strict-Transport-Security", value)
		h(w, r)
	})
}
