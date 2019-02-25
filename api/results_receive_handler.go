// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/web-platform-tests/wpt.fyi/api/checks"
	"github.com/web-platform-tests/wpt.fyi/api/receiver"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func apiResultsUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	a := receiver.NewAppEngineAPI(ctx)
	receiver.HandleResultsUpload(a, w, r)
}

func apiResultsCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	a := receiver.NewAppEngineAPI(ctx)
	s := checks.NewAPI(ctx)
	receiver.HandleResultsCreate(a, s, w, r)
}

func apiResultsNotifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)
	a := receiver.NewAppEngineAPI(ctx)

	runIDStr := r.PostFormValue("run_id")
	runID, err := strconv.ParseInt(runIDStr, 0, 0)
	if err != nil {
		log.Errorf("Invalid run_id %s", runIDStr)
		http.Error(w, fmt.Sprintf("Invalid run_id %s", runIDStr), http.StatusBadRequest)
		return
	}

	store := shared.NewAppEngineDatastore(ctx, true)
	run := new(shared.TestRun)
	if err = store.Get(store.NewIDKey("TestRun", runID), run); err != nil {
		log.Errorf("run_id %s not found", runIDStr)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	n := shared.NewNotificationsAPI(a)
	err = receiver.SendResultsAvailableNotifications(n, run)
	if err != nil {
		http.Error(w, "Error sending notifications: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
