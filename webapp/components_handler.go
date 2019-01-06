// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/gorilla/mux"
)

var packageRegex = regexp.MustCompile(`(import|from) (['"])(@[^/]*/)`)

const packageRegexReplacement = "$1 $2/node_modules/$3"

// componentsHandler loads a /node_modules/ path, and replaces any
// npm package loads in the js file with paths on the host.
func componentsHandler(w http.ResponseWriter, r *http.Request) {
	filePath := mux.Vars(r)["path"]
	var bytes []byte
	var err error
	if filePath != "" {
		bytes, err = ioutil.ReadFile(fmt.Sprintf("./node_modules/%s", filePath))
	}
	if err != nil || bytes == nil {
		http.Error(w, fmt.Sprintf("Component %s not found", filePath), http.StatusNotFound)
		return
	}
	bytes = packageRegex.ReplaceAll(bytes, []byte(packageRegexReplacement))
	w.Header().Add("Content-Type", mime.TypeByExtension(filepath.Ext(filePath)))
	w.Write(bytes)
}
