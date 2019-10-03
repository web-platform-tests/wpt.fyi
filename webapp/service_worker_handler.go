// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
)

var swTemplate *template.Template

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	file := path.Join(dir, "templates/service-worker.js")
	swTemplate = template.Must(template.ParseFiles(file))
}

// NOTE(lukebjerring): If tweaking service worker locally, change to
// sevenCharSHA, _ = regexp.Compile("^[0-9a-f]{7}|None$")
var sevenCharSHA, _ = regexp.Compile("^[0-9a-f]{7}$")

func serviceWorkerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	if !aeAPI.IsFeatureEnabled("serviceWorker") {
		http.NotFound(w, r)
		return
	}

	w.Header().Add("Content-Type", "application/javascript")
	version := strings.Split(appengine.VersionID(ctx), ".")[0]
	if !sevenCharSHA.MatchString(version) {
		http.Error(w, fmt.Sprintf("Service worker not implemented for version '%s'", version), http.StatusNotImplemented)
		return
	}

	data := struct {
		Version string
	}{
		Version: version,
	}
	swTemplate.Execute(w, data)
}
