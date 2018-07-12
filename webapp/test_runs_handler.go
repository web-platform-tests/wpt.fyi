// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// testRunsHandler handles GET/POST requests to /test-runs
func testRunsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		http.Error(w, "POST /test-runs is no longer supported.", http.StatusMethodNotAllowed)
	} else if r.Method == "GET" {
		handleTestRunGet(w, r)
	} else {
		http.Error(w, "This endpoint only supports GET and POST.", http.StatusMethodNotAllowed)
	}
}

func handleTestRunGet(w http.ResponseWriter, r *http.Request) {
	from, err := shared.ParseDateTimeParam(r, "from")
	if err != nil {
		http.Error(w, "Invalid from param: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Get runs from 3 months ago onward.
	if from == nil {
		threeMonthsAgo := time.Now().Truncate(time.Hour*24).AddDate(0, -3, 0)
		from = &threeMonthsAgo
	}
	sourceURL := fmt.Sprintf(`/api/runs?from=%s`, from.Format(time.RFC3339))

	to, err := shared.ParseDateTimeParam(r, "to")
	if err != nil {
		http.Error(w, "Invalid to param: "+err.Error(), http.StatusBadRequest)
		return
	}
	if to != nil {
		sourceURL = sourceURL + "&to=" + to.Format(time.RFC3339)
	}

	labels := shared.ParseLabelsParam(r)
	if labels != nil {
		for label := range labels.Iterator().C {
			sourceURL = sourceURL + "&label=" + label.(string)
		}
	}

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
