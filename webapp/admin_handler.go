// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"google.golang.org/appengine/memcache"

	"github.com/web-platform-tests/wpt.fyi/api/receiver"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func adminUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	a := shared.NewAppEngineAPI(ctx)
	showAdminUploadForm(a, w, r)
}

func showAdminUploadForm(a shared.AppEngineAPI, w http.ResponseWriter, r *http.Request) {
	assertAdminAndRenderTemplate(a, w, r, "/admin/results/upload", "admin_upload.html", nil)
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

func adminResultsNotifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)
	a := receiver.NewAPI(ctx)

	runIDStr := r.PostFormValue("run_id")
	runID, err := strconv.ParseInt(runIDStr, 0, 0)
	if err != nil {
		log.Errorf("Invalid run_id %s", runIDStr)
		http.Error(w, fmt.Sprintf("Invalid run_id %s", runIDStr), http.StatusBadRequest)
		return
	}

	store := shared.NewAppEngineDatastore(ctx, true)
	run := new(shared.TestRun)
	if err = store.Get(store.NewIDKey("TestRun", runID), run); err != nil {
		log.Errorf("run_id %s not found", runIDStr)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	n := shared.NewNotificationsAPI(a)

	spec := run.ProductSpec()
	title := fmt.Sprintf("New %s results available", spec.DisplayName())
	msg := fmt.Sprintf("Results are now available for %s", run.String())
	path := fmt.Sprintf("/results/?run_id=%v", run.ID)
	icon := spec.IconPath()
	if err = n.SendPushNotification(title, msg, path, &icon); err != nil {
		log.Errorf("Error sending notifications: %s", err.Error())
	}
	w.WriteHeader(http.StatusOK)
}
