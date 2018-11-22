// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
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

	testRunFilter, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get runs from a month ago, onward, by default.
	if testRunFilter.IsDefaultQuery() {
		aWeekAgo := time.Now().Truncate(time.Hour*24).AddDate(0, 0, -7)
		testRunFilter.From = &aWeekAgo
	} else if testRunFilter.MaxCount == nil {
		oneHundred := 100
		testRunFilter.MaxCount = &oneHundred
	}

	filter := convertTestRunUIFilter(testRunFilter)

	data := struct {
		Filter testRunUIFilter
	}{
		Filter: filter,
	}

	if err := templates.ExecuteTemplate(w, "test-runs.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
