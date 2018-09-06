// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// interopHandler handles the view of test results broken down by the
// number of browsers for which the test passes.
func apiInteropHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	passRateType := metrics.GetDatastoreKindName(metrics.PassRateMetadata{})
	query := datastore.NewQuery(passRateType).Order("-StartTime")

	filters, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// We load non-default queries by fetching any interop result with all their
	// TestRunIDs present in the TestRuns matching the query.
	var keysFilter mapset.Set
	if !filters.IsDefaultQuery() {
		// Load default browser runs for SHA.
		// Force any max-count to one; more than one of each product makes no sense for a interop run.
		shaFilters := filters
		limit := 128
		shaFilters.MaxCount = &limit
		keys, err := LoadTestRunKeysForFilters(ctx, shaFilters)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if len(keys) < 1 {
			http.Error(w, "No metrics runs found", http.StatusNotFound)
			return
		}
		keysFilter = mapset.NewSet()
		for _, key := range keys {
			keysFilter.Add(key.IntID())
		}
	}

	// Iterate until we find a run where all test runs matched the query.
	var interop metrics.PassRateMetadata
	it := query.Run(ctx)
	for {
		_, err := it.Next(&interop)
		if err == datastore.Done {
			http.NotFound(w, r)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if keysFilter != nil {
			all := true
			for _, id := range interop.TestRunIDs {
				if !keysFilter.Contains(id) {
					all = false
					break
				}
			}
			if !all {
				continue
			}
		}
		break
	}

	if err := interop.LoadTestRuns(ctx); err != nil {
		http.Error(w, fmt.Sprintf("Failed to load interop's test runs: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	interopBytes, err := json.Marshal(interop)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(interopBytes)
}
