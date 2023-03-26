// Copyright 2022 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

type interopData struct {
	Embedded bool
	Year     string
}

// Set of years that are valid for Interop 20XX.
var validYears = map[string]bool{"2021": true, "2022": true, "2023": true}

// Year that any invalid year will redirect to.
// TODO(danielrsmith): Change this redirect for next year's interop page.
const defaultRedirectYear = "2023"

// interopHandler handles GET requests to /interop-20XX and /compat20XX
func interopHandler(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	year := mux.Vars(r)["year"]

	// /compat20XX redirects to /interop-20XX 
    // If the year is not valid, render the 404 page instead.

	needsRedirect := name == "compat"
	if _, ok := validYears[year]; !ok {
        http.ServeFile(w, r, "components/interop-404.js")
        return
    }

	if needsRedirect {
		destination := *(r.URL)

		destination.Path = fmt.Sprintf("interop-%s", year)
		http.Redirect(w, r, destination.String(), http.StatusTemporaryRedirect)
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
		Year:     year,
	}

    

	RenderTemplate(w, r, "interop.html", data)
}
