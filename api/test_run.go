// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api //nolint:revive

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// apiTestRunHandler is responsible for emitting the test-run JSON for a specific run.
func apiTestRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is supported.", http.StatusMethodNotAllowed)

		return
	}

	vars := mux.Vars(r)
	idParam := vars["id"]
	ctx := r.Context()
	store := shared.NewAppEngineDatastore(ctx, true)
	var testRun shared.TestRun
	// nolint:nestif // TODO: Fix nestif lint error
	if idParam != "" {
		id, err := strconv.ParseInt(idParam, 10, 0)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid id '%s'", idParam), http.StatusBadRequest)

			return
		}
		run := new(shared.TestRun)
		err = store.Get(store.NewIDKey("TestRun", id), run)
		if err != nil {
			if errors.Is(err, shared.ErrNoSuchEntity) {
				http.NotFound(w, r)

				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
		testRun = *run
	} else {
		filters, err := shared.ParseTestRunFilterParams(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		} else if len(filters.Products) == 0 {
			http.Error(w, "Missing required 'product' param", http.StatusBadRequest)

			return
		} else if len(filters.Products) > 1 {
			http.Error(w, "Only one 'product' param value is allowed", http.StatusBadRequest)

			return
		}
		one := 1
		testRuns, err := store.TestRunQuery().LoadTestRuns(
			filters.Products,
			filters.Labels,
			filters.SHAs,
			nil,
			nil,
			&one,
			nil,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		allRuns := testRuns.AllRuns()
		if len(allRuns) == 0 {
			http.NotFound(w, r)

			return
		}
		testRun = allRuns[0]
	}

	testRunsBytes, err := json.Marshal(testRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	_, err = w.Write(testRunsBytes)
	if err != nil {
		logger := shared.GetLogger(r.Context())
		logger.Warningf("Failed to write data in api/run handler: %s", err.Error())
	}
}
