// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

func checkAdmin(acl shared.GitHubAccessControl, log shared.Logger, w http.ResponseWriter) bool {
	if acl == nil {
		http.Error(w, "Log in from the homepage first", http.StatusUnauthorized)
		return false
	}
	admin, err := acl.IsValidAdmin()
	if err != nil {
		log.Errorf("Error checking admin: %s", err.Error())
		http.Error(w, "Error checking admin", http.StatusInternalServerError)
		return false
	}
	if !admin {
		http.Error(w, "Admin only", http.StatusForbidden)
		return false
	}
	return true
}

func adminUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a := shared.NewAppEngineAPI(ctx)
	ds := shared.NewAppEngineDatastore(ctx, false)
	log := shared.GetLogger(ctx)
	acl, err := shared.NewGitHubAccessControlFromRequest(a, ds, r)
	if err != nil {
		log.Errorf("Error creating GitHubAccessControl: %s", err.Error())
		http.Error(w, "Error creating GitHubAccessControl", http.StatusInternalServerError)
		return
	}

	showAdminUploadForm(a, acl, log, w)
}

func showAdminUploadForm(a shared.AppEngineAPI, acl shared.GitHubAccessControl, log shared.Logger, w http.ResponseWriter) {
	if !checkAdmin(acl, log, w) {
		return
	}

	data := struct {
		CallbackURL string
	}{
		CallbackURL: fmt.Sprintf("https://%s/api/results/create", a.GetVersionedHostname()),
	}
	// We don't need user info in this template.
	RenderTemplate(w, nil, "admin_upload.html", data)
}

func adminFlagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a := shared.NewAppEngineAPI(ctx)
	ds := shared.NewAppEngineDatastore(ctx, false)
	log := shared.GetLogger(ctx)
	acl, err := shared.NewGitHubAccessControlFromRequest(a, ds, r)
	if err != nil {
		log.Errorf("Error creating GitHubAccessControl: %s", err.Error())
		http.Error(w, "Error creating GitHubAccessControl", http.StatusInternalServerError)
		return
	}

	handleAdminFlags(a, ds, acl, log, w, r)
}

func handleAdminFlags(a shared.AppEngineAPI, ds shared.Datastore, acl shared.GitHubAccessControl, log shared.Logger, w http.ResponseWriter, r *http.Request) {
	if !checkAdmin(acl, log, w) {
		return
	}

	if r.Method == http.MethodGet {
		data := struct {
			Host string
		}{
			Host: a.GetHostname(),
		}
		// We don't need user info in this template.
		RenderTemplate(w, nil, "admin_flags.html", data)
	} else if r.Method == http.MethodPost {
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

func adminCacheFlushHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a := shared.NewAppEngineAPI(ctx)
	ds := shared.NewAppEngineDatastore(ctx, false)
	log := shared.GetLogger(ctx)
	acl, err := shared.NewGitHubAccessControlFromRequest(a, ds, r)
	if err != nil {
		log.Errorf("Error creating GitHubAccessControl: %s", err.Error())
		http.Error(w, "Error creating GitHubAccessControl", http.StatusInternalServerError)
		return
	}
	if !checkAdmin(acl, log, w) {
		return
	}

	if err := shared.FlushCache(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Write([]byte("Successfully flushed cache"))
	}
}
