// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// interopHandler handles the view of test results broken down by the
// number of browsers for which the test passes.
func interopHandler(w http.ResponseWriter, r *http.Request) {
	filter, err := parseTestResultsUIFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := struct {
		Metadata string
		Filter   testResultsUIFilter
		Query    string
	}{
		Filter: filter,
		Query:  filter.Search,
	}

	ctx := shared.NewAppEngineContext(r)
	if shared.IsFeatureEnabled(shared.NewAppEngineDatastore(ctx, false), "appRoute") {
		if err := templates.ExecuteTemplate(w, "index.html", filter); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := templates.ExecuteTemplate(w, "interoperability.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
