// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"
	"net/url"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// interopHandler handles the view of test results broken down by the
// number of browsers for which the test passes.
func interopHandler(w http.ResponseWriter, r *http.Request) {
	sourceURL, _ := url.Parse("/api/interop")
	f, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sourceURL.RawQuery = f.ToQuery(true).Encode()

	data := struct {
		Metadata        string
		MetadataSources string
		SHA             string
	}{
		MetadataSources: sourceURL.String(),
		SHA:             f.SHA,
	}

	if err := templates.ExecuteTemplate(w, "interoperability.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
