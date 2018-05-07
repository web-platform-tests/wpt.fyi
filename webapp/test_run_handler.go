// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/api"
)

// testRunsHandler handles GET/POST requests to /test-runs
func testRunsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// TODO(#251): Move consumers of old endpoint.
		// /test-runs is the legacy POST endpoint, migrated to /api/run, and left to avoid breakages
		api.TestRunPostHandler(w, r)
	} else if r.Method == "GET" {
		handleTestRunGet(w, r)
	} else {
		http.Error(w, "This endpoint only supports GET and POST.", http.StatusMethodNotAllowed)
	}
}

func handleTestRunGet(w http.ResponseWriter, r *http.Request) {
	maxCount, err := api.ParseMaxCountParamWithDefault(r, 100)
	if err != nil {
		http.Error(w, "Invalid max-count param: "+err.Error(), http.StatusBadRequest)
		return
	}
	sourceURL := fmt.Sprintf(`/api/runs?max-count=%d`, maxCount)

	labels := api.ParseLabelsParam(r)
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
