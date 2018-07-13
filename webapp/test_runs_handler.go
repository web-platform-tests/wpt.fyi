// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// testRunsHandler handles GET/POST requests to /test-runs
func testRunsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET is supported.", http.StatusMethodNotAllowed)
		return
	}

	filter, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get runs from 3 months ago onward by default.
	if filter.From == nil {
		threeMonthsAgo := time.Now().Truncate(time.Hour*24).AddDate(0, -3, 0)
		filter.From = &threeMonthsAgo
	}

	query := filter.ToQuery(false)
	sourceURL := "/api/runs?" + query.Encode()

	// Serialize the data + pipe through the test-runs.html template.
	testRunSourcesBytes, err := json.Marshal([]string{sourceURL})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		TestRunSources string
	}{
		string(testRunSourcesBytes),
	}

	if err := templates.ExecuteTemplate(w, "test-runs.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
