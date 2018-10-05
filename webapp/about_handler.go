// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"
	"strings"

	"google.golang.org/appengine"
)

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var version string
	if appengine.IsDevAppServer() {
		version = strings.Split(appengine.VersionID(ctx), ".")[0]
	} else {
		version = "local-dev"
	}
	data := struct {
		Version string
	}{
		Version: version,
	}
	if err := templates.ExecuteTemplate(w, "about.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
