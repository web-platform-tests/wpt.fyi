// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// apiScreenshotRedirectHandler is responsible for redirecting to the Google Cloud Storage API
// png blob for the given screenshot hash.
//
// URL format:
// /api/screenshot/{screenshot}
func apiScreenshotRedirectHandler(w http.ResponseWriter, r *http.Request) {
	shot := mux.Vars(r)["screenshot"]
	if shot == "" {
		http.Error(w, "Screenshot id missing", http.StatusBadRequest)
	}

	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	bucket := "wptd-screenshots-staging"
	if aeAPI.GetHostname() == "wpt.fyi" {
		bucket = "wptd-screenshots"
	}
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s.png", bucket, shot)
	http.Redirect(w, r, url, http.StatusPermanentRedirect)
}
