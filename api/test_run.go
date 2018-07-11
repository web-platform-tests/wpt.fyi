// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func apiTestRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		TestRunPostHandler(w, r)
	} else if r.Method == "GET" {
		apiTestRunGetHandler(w, r)
	} else {
		http.Error(w, "This endpoint only supports GET and POST.", http.StatusMethodNotAllowed)
	}
}

// apiTestRunGetHandler is responsible for emitting the test-run JSON for a specific run.
func apiTestRunGetHandler(w http.ResponseWriter, r *http.Request) {
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
		testRuns, err := shared.LoadTestRuns(ctx, filters.Products, filters.Labels, []string{filters.SHA}, nil, nil, &one)
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

// TestRunPostHandler is responsible for handling TestRun submissions (via HTTP POST requests).
// It asserts the presence of a required secret token, then saves the JSON blob to the Datastore.
// See shared.go for the JSON format expected.
// It is exported for re-use as the legacy endpoint '/test-runs' in the webapp.
func TestRunPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var err error

	// Fetch pre-uploaded shared.Token entity.
	suppliedSecret := r.URL.Query().Get("secret")
	tokenKey := datastore.NewKey(ctx, "Token", "upload-token", 0, nil)
	var token shared.Token
	if err = datastore.Get(ctx, tokenKey, &token); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if suppliedSecret != token.Secret {
		http.Error(w, fmt.Sprintf("Invalid token '%s'", suppliedSecret), http.StatusUnauthorized)
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var testRun shared.TestRun
	if err = json.Unmarshal(body, &testRun); err != nil {
		http.Error(w, "Failed to parse JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if testRun.TimeStart.IsZero() {
		testRun.TimeStart = time.Now()
	}
	if testRun.TimeEnd.IsZero() {
		testRun.TimeEnd = testRun.TimeStart
	}
	testRun.CreatedAt = time.Now()

	// Create a new shared.TestRun out of the JSON body of the request.
	key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
	if _, err := datastore.Put(ctx, key, &testRun); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var jsonOutput []byte
	if jsonOutput, err = json.Marshal(testRun); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonOutput)
}
