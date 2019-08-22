// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine/memcache"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

func adminUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	a := shared.NewAppEngineAPI(ctx)
	showAdminUploadForm(a, w, r)
}

func showAdminUploadForm(a shared.AppEngineAPI, w http.ResponseWriter, r *http.Request) {
	data := struct {
		CallbackURL string
	}{
		CallbackURL: fmt.Sprintf("https://%s/api/results/create", a.GetVersionedHostname()),
	}
	assertAdminAndRenderTemplate(a, w, r, "/admin/results/upload", "admin_upload.html", data)
}

func adminFlagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	a := shared.NewAppEngineAPI(ctx)
	ds := shared.NewAppEngineDatastore(ctx, false)

	data := struct {
		Host string
	}{
		Host: a.GetHostname(),
	}
	if r.Method == "GET" {
		assertAdminAndRenderTemplate(a, w, r, "/admin/flags", "admin_flags.html", data)
	} else if r.Method == "POST" {
		if !a.IsAdmin() {
			http.Error(w, "Admin only", http.StatusUnauthorized)
			return
		}
		var flag shared.Flag
		if bytes, err := ioutil.ReadAll(r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if err = json.Unmarshal(bytes, &flag); err != nil {
			http.Error(w, fmt.Sprintf("Failed to unmarshal flag: %s", err.Error()), http.StatusBadRequest)
			return
		} else if err = shared.SetFeature(ds, flag); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save feature %s: %s", flag.Name, err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

func assertAdminAndRenderTemplate(
	a shared.AppEngineAPI,
	w http.ResponseWriter,
	r *http.Request,
	redirectPath,
	template string,
	data interface{}) {
	if !a.IsAdmin() {
		http.Error(w, "Admin only", http.StatusUnauthorized)
		return
	}

	if err := templates.ExecuteTemplate(w, template, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func adminCacheFlushHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	a := shared.NewAppEngineAPI(ctx)

	if !a.IsAdmin() {
		http.Error(w, "Admin only", http.StatusUnauthorized)
		return
	}
	if err := memcache.Flush(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Write([]byte("Successfully flushed cache"))
	}
}
