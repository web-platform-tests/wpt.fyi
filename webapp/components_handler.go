// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/gorilla/mux"
)

var packageRegex = regexp.MustCompile(`(import .* from|import) (['"])(@[^/]*/)`)

const packageRegexReplacement = "$1 $2/node_modules/$3"

var rootDir string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	rootDir = filepath.Dir(filename)
}

// componentsHandler loads a /node_modules/ path, and replaces any
// npm package loads in the js file with paths on the host.
func componentsHandler(w http.ResponseWriter, r *http.Request) {
	filePath := mux.Vars(r)["path"]
	var bytes []byte
	var err error
	if filePath != "" {
		bytes, err = ioutil.ReadFile(fmt.Sprintf(path.Join(rootDir, "node_modules", filePath)))
	}
	if err != nil || bytes == nil {
		http.Error(w, fmt.Sprintf("Component %s not found", filePath), http.StatusNotFound)
		return
	}
	bytes = packageRegex.ReplaceAll(bytes, []byte(packageRegexReplacement))
	// Cache up to a day (same as the default expiration in app.yaml).
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(filePath)))
	w.Write(bytes)
}
