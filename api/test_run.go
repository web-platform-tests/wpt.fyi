// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// apiTestRunHandler is responsible for emitting the test-run JSON for a specific run.
func apiTestRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET is supported.", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	idParam := vars["id"]
	ctx := appengine.NewContext(r)
	var testRun shared.TestRun
	if idParam != "" {
		id, err := strconv.ParseInt(idParam, 10, 0)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid id '%s'", idParam), http.StatusBadRequest)
			return
		}
		run, err := shared.LoadTestRun(ctx, id)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		testRun = *run
	} else {
		filters, err := shared.ParseTestRunFilterParams(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if len(filters.Products) == 0 {
			http.Error(w, fmt.Sprintf("Missing required 'product' param"), http.StatusBadRequest)
			return
		} else if len(filters.Products) > 1 {
			http.Error(w, fmt.Sprintf("Only one 'product' param value is allowed"), http.StatusBadRequest)
			return
		}
		one := 1
		testRuns, err := shared.LoadTestRuns(ctx, filters.Products, filters.Labels, filters.SHA, nil, nil, &one)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(testRuns) == 0 {
			http.NotFound(w, r)
			return
		}
		testRun = testRuns[0]
	}

	testRunsBytes, err := json.Marshal(testRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(testRunsBytes)
}
