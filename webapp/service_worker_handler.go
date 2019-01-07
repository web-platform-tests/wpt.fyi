// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"fmt"
	"io/ioutil"
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

var (
	swTemplate = template.Must(template.ParseFiles("templates/service-worker.js"))
	// NOTE(lukebjerring): If tweaking service worker locally, change to
	// sevenCharSHA, _ = regexp.Compile("^[0-9a-f]{7}|None$")
	sevenCharSHA, _ = regexp.Compile("^[0-9a-f]{7}$")
)

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

	var files []string
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	for _, folder := range []string{"components", "static"} {
		appendFiles(path.Join(dir, folder), &files)
	}

	data := struct {
		Version string
		Files   []string
	}{
		Version: version,
		Files:   files,
	}
	swTemplate.Execute(w, data)
}

func appendFiles(dirPath string, files *[]string) error {
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, f := range dir {
		name := path.Join(dirPath, f.Name())
		if f.IsDir() {
			appendFiles(name, files)
		} else {
			*files = append(*files, name)
		}
	}
	return nil
}
