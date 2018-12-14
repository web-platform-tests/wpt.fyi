// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// insightsHandler handles the view listing a range of useful queries for the
// wpt results.
func insightsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	if !aeAPI.IsFeatureEnabled("insightsTab") {
		http.NotFound(w, r)
		return
	}

	data := struct{}{}
	if err := templates.ExecuteTemplate(w, "insights.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
