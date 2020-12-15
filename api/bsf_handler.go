// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// apiBSFHandler fetches browser-specific failure data based on the URL params.
func apiBSFHandler(w http.ResponseWriter, r *http.Request) {
	handleBSF(w, r, shared.NewFetchBSF())
}

func handleBSF(w http.ResponseWriter, r *http.Request, fetcher shared.FetchBSF) {
	q := r.URL.Query()
	isExperimental := false
	val, _ := shared.ParseBooleanParam(q, "experimental")
	// If the experimental parameter is missing or present with no value, set
	// isExperimental to the default value, false.
	if val != nil {
		isExperimental = *val
	}

	lines, err := fetcher.Fetch(isExperimental)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var from *time.Time
	if from, err = shared.ParseDateTimeParam(q, "from"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var to *time.Time
	if to, err = shared.ParseDateTimeParam(q, "to"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bsfData := shared.FilterandExtractBSFData(lines, from, to)
	marshalled, err := json.Marshal(bsfData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(marshalled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
