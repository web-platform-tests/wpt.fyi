// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"
)

// insightsHandler handles the view listing a range of useful queries for the
// wpt results.
func insightsHandler(w http.ResponseWriter, r *http.Request) {
	data := struct{}{}
	if err := templates.ExecuteTemplate(w, "insights.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
