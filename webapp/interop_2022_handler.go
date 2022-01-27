// Copyright 2022 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

type interop2022Data struct {
	Embedded bool
}

// interop2022Handler handles GET requests to /interop-2022
func interop2022Handler(w http.ResponseWriter, r *http.Request) {
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

	data := interop2022Data{
		Embedded: embedded != nil && *embedded,
	}
	RenderTemplate(w, r, "interop-2022.html", data)
}
