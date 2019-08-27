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
	"github.com/web-platform-tests/wpt.fyi/api/checks"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func apiResultsUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	a := NewAPI(ctx)
	HandleResultsUpload(a, w, r)
}

func apiResultsCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	a := NewAPI(ctx)
	s := checks.NewAPI(ctx)
	HandleResultsCreate(a, s, w, r)
}

func apiPendingTestRunUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Only PATCH is supported", http.StatusMethodNotAllowed)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	a := NewAPI(ctx)
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
