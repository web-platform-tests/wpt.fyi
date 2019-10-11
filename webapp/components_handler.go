// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate packr2

package webapp

import (
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
)

const packageRegexReplacement = "$1 $2/node_modules/$3"

var (
	packageRegex = regexp.MustCompile(`(import .* from|import) (['"])(@[^/]*/)`)
	box          *packr.Box
)

func init() {
	box = packr.New("node modules", "./node_modules/")
}

// componentsHandler loads a /node_modules/ path, and replaces any
// npm package loads in the js file with paths on the host.
func componentsHandler(w http.ResponseWriter, r *http.Request) {
	filePath := mux.Vars(r)["path"]
	body, err := box.FindString(filePath)
	if err != nil || body == "" {
		http.Error(w, fmt.Sprintf("Component %s not found", filePath), http.StatusNotFound)
		return
	}
	body = packageRegex.ReplaceAllString(body, packageRegexReplacement)
	// Cache up to a day (same as the default expiration in app.yaml).
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(filePath)))
	w.Write([]byte(body))
}
