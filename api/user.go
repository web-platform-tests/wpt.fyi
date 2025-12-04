// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file

package api //nolint:revive

import (
	"encoding/json"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

type loginSuccessResponse struct {
	User *shared.User `json:"user"`
}

type loginFailureResponse struct {
	Error string `json:"error"`
}

func apiUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	aeAPI := shared.NewAppEngineAPI(ctx)
	if !aeAPI.IsFeatureEnabled("githubLogin") {
		http.Error(w, "Feature not enabled", http.StatusNotImplemented)

		return
	}

	ds := shared.NewAppEngineDatastore(ctx, false)
	user, _ := shared.GetUserFromCookie(ctx, ds, r)
	if user == nil {
		response := loginFailureResponse{Error: "Unable to retrieve login information, please log in again"}
		marshalled, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(marshalled)
		if err != nil {
			logger := shared.GetLogger(ctx)
			logger.Warningf("Failed to write data in api/user handler: %s", err.Error())
		}

		return
	}

	response := loginSuccessResponse{User: user}
	marshalled, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	_, err = w.Write(marshalled)
	if err != nil {
		logger := shared.GetLogger(ctx)
		logger.Warningf("Failed to write data in api/user handler: %s", err.Error())
	}
}
