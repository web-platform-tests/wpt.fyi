// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file

package api

import (
	"encoding/json"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

type loginResponse struct {
	User  *shared.User `json:"user,omitempty"`
	Error string       `json:"error,omitempty"`
}

func apiUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	if !aeAPI.IsFeatureEnabled("githubLogin") {
		http.Error(w, "Feature not enabled", http.StatusNotImplemented)
		return
	}

	ds := shared.NewAppEngineDatastore(ctx, false)
	user, token := shared.GetUserFromCookie(ctx, ds, r)
	if user == nil || token == nil {
		response := loginResponse{Error: "Unable to retrieve log-in information, please log in again"}
		marshalled, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		w.Write(marshalled)
		return
	}

	response := loginResponse{User: user}
	marshalled, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(marshalled)
}
