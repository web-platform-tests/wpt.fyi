// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// apiPendingTestRunsHandler is responsible for emitting JSON for
// all the pending test runs.
func apiPendingTestRunsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	store := shared.NewAppEngineDatastore(ctx, false)

	q := store.NewQuery("PendingTestRun").Order("-Updated")
	var runs []shared.PendingTestRun
	keys, err := store.GetAll(q, &runs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i, key := range keys {
		runs[i].ID = key.IntID()
	}

	testRunsBytes, err := json.Marshal(runs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(testRunsBytes)
}
