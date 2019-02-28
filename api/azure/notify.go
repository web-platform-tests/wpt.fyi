// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package azure

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func notifyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var buildID int64
	var err error
	if buildID, err = strconv.ParseInt(id, 0, 0); err != nil {
		http.Error(w, fmt.Sprintf("Invalid build id: %s", id), http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	azureAPI := NewAPI(ctx)
	log := shared.GetLogger(ctx)

	processed, err := processAzureBuild(
		aeAPI,
		azureAPI,
		"web-platform-tests",
		"wpt",
		"",                      // No sender info.
		r.FormValue("artifact"), // artifact=foo will only process foo.
		buildID)

	if err != nil {
		log.Errorf("%v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if processed {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Azure build artifacts retrieved successfully")
	} else {
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintln(w, "Notification of build artifacts was ignored")
	}
	return
}
