// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package spanner

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/web-platform-tests/wpt.fyi/api/auth"
)

// PushID is a unique identifier for a request to push a test run to
// Cloud Spanner.
type PushID struct {
	Time  time.Time `json:"time"`
	RunID int64     `json:"run_id"`
}

// HandlePushRun handles a request to push a test run to Cloud Spanner.
func HandlePushRun(a auth.AppEngineAPI, w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Only PUT is supported", http.StatusMethodNotAllowed)
		return
	}

	username, password, ok := r.BasicAuth()
	if !ok || username != auth.InternalUsername || !a.AuthenticateUploader(username, password) {
		http.Error(w, "Authentication error", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(r.URL.Query().Get("run_id"), 10, 0)
	if err != nil {
		http.Error(w, `Missing or invalid query param: "run_id"`, http.StatusBadRequest)
		return
	}

	t := time.Now().UTC()

	// TODO(mdittmer): Load run report URL from datastore, load report from GCS,
	// write results to Cloud Spanner.

	data, err := json.Marshal(PushID{t, id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(data)
}
