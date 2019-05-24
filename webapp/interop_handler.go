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
	filter, err := parseTestResultsUIFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := templates.ExecuteTemplate(w, "index.html", filter); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
