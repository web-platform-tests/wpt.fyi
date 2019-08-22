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

	q := store.NewQuery("PendingTestRun")
	var runs []shared.PendingTestRun
	if _, err := store.GetAll(q, &runs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	testRunsBytes, err := json.Marshal(runs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(testRunsBytes)
}
