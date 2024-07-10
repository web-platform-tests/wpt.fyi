// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"embed"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/gorilla/mux"
)

const packageRegexReplacement = "$1 $2/node_modules/$3"

var (
	packageRegex = regexp.MustCompile(`(import .* from|import) (['"])(@[^/]*/)`)
	//go:embed node_modules
	nodeModules embed.FS
)

// componentsHandler loads a /node_modules/ path, and replaces any
// npm package loads in the js file with paths on the host.
func componentsHandler(w http.ResponseWriter, r *http.Request) {
	filePath := mux.Vars(r)["path"]
	body, err := nodeModules.ReadFile(filePath)
	if err != nil || body == nil {
		http.Error(w, fmt.Sprintf("Component %s not found", filePath), http.StatusNotFound)
		return
	}
	body = packageRegex.ReplaceAll(body, []byte(packageRegexReplacement))
	// Cache up to a day (same as the default expiration in app.yaml).
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(filePath)))
	w.Write(body)
}
