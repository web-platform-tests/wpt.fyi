// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// apiSHAsHandler is responsible for emitting just the revision SHAs for test runs.
func apiSHAsHandler(w http.ResponseWriter, r *http.Request) {
	filters, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)

	var shas []string
	products := filters.GetProductsOrDefault()
	if filters.Aligned != nil && *filters.Aligned {
		if shas, _, err = shared.GetAlignedRunSHAs(ctx, products, filters.Labels, filters.From, filters.To, filters.MaxCount); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		testRuns, err := shared.LoadTestRuns(ctx, products, filters.Labels, shared.LatestSHA, filters.From, filters.To, filters.MaxCount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		seen := mapset.NewSet()
		for _, testRun := range testRuns {
			if !seen.Contains(testRun.Revision) {
				shas = append(shas, testRun.Revision)
				seen.Add(testRun.Revision)
			}
		}
	}
	if len(shas) < 1 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("[]"))
		return
	}

	shasBytes, err := json.Marshal(shas)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(shasBytes)
}
