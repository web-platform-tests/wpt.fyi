// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"
)

// interopHandler handles the view of test results broken down by the
// number of browsers for which the test passes.
func interopHandler(w http.ResponseWriter, r *http.Request) {
	uiFilters, err := parseTestRunUIFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := struct {
		Metadata string
		Filter   testRunUIFilter
		Query    string
	}{
		Filter: uiFilters,
		Query:  r.URL.Query().Get("q"),
	}

	if err := templates.ExecuteTemplate(w, "interoperability.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
