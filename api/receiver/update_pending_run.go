// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// HandleUpdatePendingTestRun handles the PATCH request for updating pending test runs.
func HandleUpdatePendingTestRun(a API, w http.ResponseWriter, r *http.Request) {
	if AuthenticateUploader(a, r) != InternalUsername {
		http.Error(w, "This is a private API.", http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var run shared.PendingTestRun
	if err := json.Unmarshal(body, &run); err != nil {
		http.Error(w, "Failed to parse JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := strconv.ParseInt(idParam, 10, 0)
	if err != nil {
		http.Error(w, "Invalid ID: "+idParam, http.StatusBadRequest)
		return
	}
	if id != run.ID {
		http.Error(w, fmt.Sprintf("Inconsistent ID: %d != %d", id, run.ID), http.StatusBadRequest)
		return
	}

	if err := a.UpdatePendingTestRun(run); err != nil {
		http.Error(w, "Failed to update run: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
