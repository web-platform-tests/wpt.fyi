// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/api/receiver"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func adminUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	a := receiver.NewAppEngineAPI(ctx)
	showAdminUploadForm(a, w, r)
}

func showAdminUploadForm(a receiver.AppEngineAPI, w http.ResponseWriter, r *http.Request) {
	if !a.IsLoggedIn() {
		loginURL, err := a.LoginURL("/admin/results/upload")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}
	if !a.IsAdmin() {
		http.Error(w, "Admin only", http.StatusUnauthorized)
		return
	}

	if err := templates.ExecuteTemplate(w, "admin_upload.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
