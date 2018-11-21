// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
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

	filters, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var interop *metrics.PassRateMetadataLegacy
	if interop, err = loadMostRecentInteropRun(ctx, filters); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if interop == nil {
		if interop, err = loadFallbackInteropRun(ctx, filters); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if interop == nil {
		http.Error(w, "Interop data not found", http.StatusNotFound)
		return
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

func loadMostRecentInteropRun(ctx context.Context, filters shared.TestRunFilter) (result *metrics.PassRateMetadataLegacy, err error) {
	// Load default browser runs for SHA.
	// Force any max-count to one; more than one of each product makes no sense for a interop run.
	shaFilters := filters
	limit := 1
	shaFilters.MaxCount = &limit
	keys, err := LoadTestRunKeysForFilters(ctx, shaFilters)
	if err != nil {
		return nil, err
	} else if len(keys) < 1 {
		return nil, nil
	}
	passRateType := metrics.GetDatastoreKindName(metrics.PassRateMetadata{})
	query := datastore.NewQuery(passRateType).Order("-StartTime").Limit(1)
	for _, key := range keys.AllKeys() {
		query = query.Filter("TestRunIDs =", key.IntID())
	}
	var results []metrics.PassRateMetadataLegacy
	if _, err = query.GetAll(ctx, &results); err != nil {
		return nil, err
	} else if len(results) < 1 {
		return nil, nil
	}
	return &results[0], nil
}

func loadFallbackInteropRun(ctx context.Context, filters shared.TestRunFilter) (result *metrics.PassRateMetadataLegacy, err error) {
	passRateType := metrics.GetDatastoreKindName(metrics.PassRateMetadata{})
	query := datastore.NewQuery(passRateType).Order("-StartTime").Limit(100)

	// We load non-default queries by fetching any interop result with all their
	// TestRunIDs present in the TestRuns matching the query.
	var keysChecker func(shared.TestRunIDs) bool
	if !filters.IsDefaultQuery() {
		products := filters.GetProductsOrDefault()
		// When ?aligned=true, make sure to show results for the same aligned run.
		// We don't want to mismatch an interop which has runs from several different SHAs
		// (but, each SHA being from an aligned run), so we need to keep the keys grouped.
		if shared.IsLatest(filters.SHA) && filters.Aligned != nil && *filters.Aligned {
			ten := 10
			_, shaKeys, err := shared.GetAlignedRunSHAs(ctx, products, filters.Labels, filters.From, filters.To, &ten)
			if err != nil {
				return nil, err
			} else if len(shaKeys) < 1 {
				return nil, nil
			}
			keysChecker = checkKeysAreAligned(shaKeys)
		} else {
			// We arbitrarily take at most 16 * N runs, which should (typically) be ~16 sets.
			shaFilters := filters
			limit := 16 * len(products)
			shaFilters.MaxCount = &limit
			keys, err := LoadTestRunKeysForFilters(ctx, shaFilters)

			if err != nil {
				return nil, err
			} else if len(keys) < 1 {
				return nil, nil
			}
			keysChecker = checkKeysMatchQuery(keys)
		}
	}

	// Iterate until we find interop data where its TestRunIDs match the query.
	var interop metrics.PassRateMetadataLegacy
	it := query.Run(ctx)
	found := false
	for {
		_, err := it.Next(&interop)
		if err == datastore.Done {
			return nil, nil
		} else if err != nil {
			return nil, err
		}
		if keysChecker != nil && !keysChecker(interop.TestRunIDs) {
			continue
		}
		found = true
		break
	}
	if !found {
		return nil, nil
	}
	return &interop, nil
}

func checkKeysAreAligned(shaKeys map[string]shared.KeysByProduct) func(shared.TestRunIDs) bool {
	keySets := make(map[string]mapset.Set)
	for sha, keys := range shaKeys {
		keySet := mapset.NewSet()
		for _, key := range keys.AllKeys() {
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

func checkKeysMatchQuery(keys shared.KeysByProduct) func(shared.TestRunIDs) bool {
	keysFilter := mapset.NewSet()
	for _, key := range keys.AllKeys() {
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
