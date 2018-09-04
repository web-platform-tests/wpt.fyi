// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"

	"github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/datastore"
)

// interopHandler handles the view of test results broken down by the
// number of browsers for which the test passes.
func apiInteropHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	passRateType := metrics.GetDatastoreKindName(metrics.PassRateMetadata{})
	query := datastore.NewQuery(passRateType).Order("-StartTime").Limit(1)

	filters, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// We 'load by SHA' by fetching any interop result with all TestRunIDs for that SHA.
	if !filters.IsDefaultQuery() {
		// Load default browser runs for SHA.
		// Force any max-count to one; more than one of each product makes no sense for a interop run.
		one := 1
		filters.MaxCount = &one
		runs, err := LoadTestRunsForFilters(ctx, filters)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if len(runs) < 1 {
			http.Error(w, "No metrics runs found", http.StatusNotFound)
			return
		}
		for _, run := range runs {
			query = query.Filter("TestRunIDs =", run.ID)
		}
	}

	var metadataSlice []metrics.PassRateMetadataLegacy
	if _, err := query.GetAll(ctx, &metadataSlice); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(metadataSlice) != 1 {
		http.Error(w, "No metrics runs found", http.StatusNotFound)
		return
	}

	metadata := &metadataSlice[0]
	if err := metadata.LoadTestRuns(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	metadataBytes, err := json.Marshal(*metadata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(metadataBytes)
}
