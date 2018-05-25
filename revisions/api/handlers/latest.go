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
		w.WriteHeader(503)
		w.Write(a.ErrorJSON("Announcer not yet initialized"))
		return
	}

	epochs := a.GetEpochs()
	if len(epochs) == 0 {
		w.WriteHeader(500)
		w.Write(a.ErrorJSON("No epochs"))
		return
	}

	now := time.Now()
	revs, err := ancr.GetRevisions(a.GetLatestGetRevisionsInput(), announcer.Limits{
		Now:   now,
		Start: now.Add(-2 * epochs[0].GetData().MaxDuration),
	})
	if err != nil {
		w.WriteHeader(500)
		w.Write(a.ErrorJSON(err.Error()))
		return
	}

	response, err := api.LatestFromEpochs(revs)
	if err != nil {
		w.WriteHeader(500)
		w.Write(a.ErrorJSON(err.Error()))
		return
	}

	bytes, err := a.Marshal(response)
	if err != nil {
		w.WriteHeader(500)
		w.Write(a.ErrorJSON("Failed to marshal latest epochal revisions JSON"))
		return
	}

	w.Write(bytes)
}
