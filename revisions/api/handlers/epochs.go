// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handlers

import (
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/revisions/api"
)

// EpochsHandler handles HTTP requests for a listing of supported epochs.
func EpochsHandler(a api.API, w http.ResponseWriter, r *http.Request) {
	bytes, err := a.Marshal(a.GetAPIEpochs())
	if err != nil {
		w.WriteHeader(500)
		w.Write(a.ErrorJSON("Failed to marshal epochs JSON"))
		return
	}
	w.Write(bytes)
}
