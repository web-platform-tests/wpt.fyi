// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handlers

import (
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/revisions/api"
	"github.com/web-platform-tests/wpt.fyi/revisions/api/push"
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

	response, err := push.GetLatestRevisions(a, ancr, epochs)
	if err != nil {
		http.Error(w, a.ErrorJSON(err.Error()), http.StatusInternalServerError)
		return
	}

	bytes, err := a.Marshal(*response)
	if err != nil {
		http.Error(w, a.ErrorJSON("Failed to marshal latest epochal revisions JSON"), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
