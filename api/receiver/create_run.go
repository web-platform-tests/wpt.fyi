// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// InternalUsername is a special uploader whose password is kept secret and can
// only be accessed by services in this AppEngine project via Datastore.
const InternalUsername = "_processor"

// HandleResultsCreate handles the POST requests for creating test runs.
func HandleResultsCreate(a AppEngineAPI, w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || username != InternalUsername || !a.AuthenticateUploader(username, password) {
		http.Error(w, "Authentication error", http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var testRun shared.TestRun
	if err := json.Unmarshal(body, &testRun); err != nil {
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

	key, err := a.AddTestRun(&testRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy int64 representation of key into TestRun.ID so that clients can
	// inspect/use key value.
	testRun.ID = key.IntID()

	jsonOutput, err := json.Marshal(testRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonOutput)
}
