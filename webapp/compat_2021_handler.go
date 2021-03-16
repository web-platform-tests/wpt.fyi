// Copyright 2021 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

type compat2021Data struct {
	Embedded bool
}

// compat2021Handler handles GET requests to /compat2021
func compat2021Handler(w http.ResponseWriter, r *http.Request) {
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

	data := compat2021Data{
		Embedded: embedded != nil && *embedded,
	}
	RenderTemplate(w, r, "compat-2021.html", data)
}
