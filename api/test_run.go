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
		key := datastore.NewKey(ctx, "TestRun", "", id, nil)
		if err = datastore.Get(ctx, key, &testRun); err != nil {
			if err == datastore.ErrNoSuchEntity {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		runSHA, err := shared.ParseSHAParam(r)
		if err != nil {
			http.Error(w, "Invalid query params", http.StatusBadRequest)
			return
		}

		var browser, product *shared.Product
		product, err = shared.ParseProductParam(r)
		if err != nil {
			http.Error(w, "Invalid 'product' param", http.StatusBadRequest)
			return
		}
		browser, err = shared.ParseBrowserParam(r)
		if err != nil {
			http.Error(w, "Invalid 'browser' param", http.StatusBadRequest)
			return
		}
		if product == nil && browser != nil {
			product = browser
		}
		if product == nil {
			http.Error(w, "Missing required 'product' param", http.StatusBadRequest)
			return
		}
		testRuns, err := shared.LoadTestRuns(ctx, []shared.Product{*product}, runSHA, nil, 1)
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

	// Use 'now' as created time, unless flagged as retroactive.
	if retro, err := strconv.ParseBool(r.URL.Query().Get("retroactive")); err != nil || !retro {
		testRun.CreatedAt = time.Now()
	}

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
	w.Write(jsonOutput)
	w.WriteHeader(http.StatusCreated)
}
