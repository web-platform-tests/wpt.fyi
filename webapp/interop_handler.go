// Copyright 2022 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

type interopData struct {
	Embedded bool
	Year     string
}

// interopHandler handles GET requests to /interop
func interopHandler(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	year, err := strconv.Atoi(path)
	// TODO(danielrsmith): Change this redirect for next year's interop.
	if err != nil {
		http.Redirect(w, r, "2022", http.StatusTemporaryRedirect)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Only GET is supported.", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	embedded, err := shared.ParseBooleanParam(q, "embedded")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := interopData{
		Embedded: embedded != nil && *embedded,
		Year:     strconv.Itoa(year),
	}
	RenderTemplate(w, r, "interop.html", data)
}
