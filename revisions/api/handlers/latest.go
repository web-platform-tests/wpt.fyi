// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handlers

import (
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/revisions/announcer"
	"github.com/web-platform-tests/wpt.fyi/revisions/api"
)

// LatestHandler handles HTTP requests for the latest epochal revisions.
func LatestHandler(a api.API, w http.ResponseWriter, r *http.Request) {
	ancr := a.GetAnnouncer()
	if ancr == nil {
		http.Error(w, a.ErrorJSON("Announcer not yet initialized"), http.StatusServiceUnavailable)
		return
	}

	epochs := a.GetEpochs()
	if len(epochs) == 0 {
		http.Error(w, a.ErrorJSON("No epochs"), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	revs, err := ancr.GetRevisions(a.GetLatestGetRevisionsInput(), announcer.Limits{
		At:    now,
		Start: now.Add(-2 * epochs[0].GetData().MaxDuration),
	})
	if err != nil {
		http.Error(w, a.ErrorJSON(err.Error()), http.StatusInternalServerError)
		return
	}

	response, err := api.LatestFromEpochs(revs)
	if err != nil {
		http.Error(w, string(a.ErrorJSON(err.Error())), http.StatusInternalServerError)
		return
	}

	bytes, err := a.Marshal(response)
	if err != nil {
		http.Error(w, a.ErrorJSON("Failed to marshal latest epochal revisions JSON"), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
