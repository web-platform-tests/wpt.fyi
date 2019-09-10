// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// apiPendingTestRunsHandler is responsible for emitting JSON for
// all the non-completed PendingTestRun entities.
func apiPendingTestRunsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	store := shared.NewAppEngineDatastore(ctx, false)

	filter := strings.ToLower(mux.Vars(r)["filter"])
	q := store.NewQuery("PendingTestRun")
	switch filter {
	case "pending":
		q = q.Order("-Stage").Filter("Stage < ", shared.StageValid)
	case "invalid":
		q = q.Order("Stage").Filter("Stage > ", shared.StageValid)
	case "":
		// No-op
	default:
		http.Error(w, "Invalid filter: "+filter, http.StatusBadRequest)
		return
	}
	q = q.Order("-Updated")

	var runs []shared.PendingTestRun
	keys, err := store.GetAll(q, &runs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i, key := range keys {
		runs[i].ID = key.IntID()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// When we filter by status (pending) we need to re-sort.
	sort.Sort(sort.Reverse(shared.PendingTestRunByUpdated(runs)))
	emit(w, runs)
}

// emit to the given writer the JSON marshalled output of the given interface.
func emit(w http.ResponseWriter, i interface{}) {
	data, err := json.Marshal(i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
