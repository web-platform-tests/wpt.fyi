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
	"google.golang.org/appengine/datastore"
)

// interopHandler handles the view of test results broken down by the
// number of browsers for which the test passes.
func apiInteropHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	passRateType := metrics.GetDatastoreKindName(metrics.PassRateMetadata{})
	query := datastore.NewQuery(passRateType).Order("-StartTime")

	filters, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// We load non-default queries by fetching any interop result with all their
	// TestRunIDs present in the TestRuns matching the query.
	var keysChecker func(shared.TestRunIDs) bool
	if !filters.IsDefaultQuery() {
		// When ?complete=true, make sure to show results for the same complete run
		// (executed for all browsers). Because we don't want to mismatch an interop
		// which has SHAs from 2 separated-but-complete runs, we need to keep the
		// keys grouped.
		if shared.IsLatest(filters.SHA) && filters.Aligned != nil && *filters.Aligned {
			ten := 10
			_, shaKeys, err := shared.GetAlignedRunSHAs(ctx, filters.GetProductsOrDefault(), filters.Labels, filters.From, filters.To, &ten)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else if len(shaKeys) < 1 {
				// Bail out early - can't find any complete runs.
				http.Error(w, "No metrics runs found", http.StatusNotFound)
				return
			}
			keysChecker = checkKeysAreAligned(shaKeys)
		} else {
			// Load default browser runs for SHA.
			// Force any max-count to one; more than one of each product makes no sense for a interop run.
			shaFilters := filters
			limit := 64
			shaFilters.MaxCount = &limit
			keys, err := LoadTestRunKeysForFilters(ctx, shaFilters)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if len(keys) < 1 {
				http.Error(w, "No metrics runs found", http.StatusNotFound)
				return
			}
			keysChecker = checkKeysMatchQuery(keys)
		}
	}

	// Iterate until we find interop data where its TestRunIDs match the query.
	var interop metrics.PassRateMetadata
	it := query.Run(ctx)
	for {
		_, err := it.Next(&interop)
		if err == datastore.Done {
			http.Error(w, "No metrics runs found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if keysChecker != nil && !keysChecker(interop.TestRunIDs) {
			continue
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

func checkKeysAreAligned(shaKeys map[string][]*datastore.Key) func(shared.TestRunIDs) bool {
	keySets := make(map[string]mapset.Set)
	for sha, keys := range shaKeys {
		keySet := mapset.NewSet()
		for _, key := range keys {
			keySet.Add(key.IntID())
		}
		keySets[sha] = keySet
	}
	return func(ids shared.TestRunIDs) bool {
		for _, keys := range keySets {
			all := true
			for _, id := range ids {
				if !keys.Contains(id) {
					all = false
					break
				}
			}
			if all {
				return true
			}
		}
		return false
	}
}

func checkKeysMatchQuery(keys []*datastore.Key) func(shared.TestRunIDs) bool {
	keysFilter := mapset.NewSet()
	for _, key := range keys {
		keysFilter.Add(key.IntID())
	}
	return func(ids shared.TestRunIDs) bool {
		all := true
		for _, id := range ids {
			if !keysFilter.Contains(id) {
				all = false
				break
			}
		}
		return all
	}
}
