// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// apiBSFHandler fetches browser-specific failure data.
func apiBSFHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := shared.GetLogger(ctx)
	path, _ := os.Getwd()
	rootPath := filepath.Dir(filepath.Dir(path))
	fmt.Println(rootPath)
	file, err := os.Open(filepath.Join(rootPath, "api", "stable-browser-specific-failures.csv"))
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	marshalled, err := json.Marshal(lines)
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(marshalled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
