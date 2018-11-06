// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// apiTestRunsHandler is responsible for emitting test-run JSON for all the runs at a given SHA.
//
// URL Params:
//     sha: SHA[0:10] of the repo when the tests were executed (or 'latest')
func apiTestRunsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	ids, err := shared.ParseRunIDsParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var testRuns shared.TestRuns
	if len(ids) > 0 {
		testRuns, err = ids.LoadTestRuns(ctx)
		if multiError, ok := err.(appengine.MultiError); ok {
			all404s := true
			for _, err := range multiError {
				if err != datastore.ErrNoSuchEntity {
					all404s = false
				}
			}
			if all404s {
				w.WriteHeader(http.StatusNotFound)
				err = nil
			}
		}
	} else {
		var filters shared.TestRunFilter
		filters, err = shared.ParseTestRunFilterParams(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		testRuns, err = LoadTestRunsForFilters(ctx, filters)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(testRuns) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("[]"))
		return
	}

	testRunsBytes, err := json.Marshal(testRuns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(testRunsBytes)
}

// LoadTestRunKeysForFilters deciphers the filters and executes a corresponding
// query to load the TestRun keys.
func LoadTestRunKeysForFilters(ctx context.Context, filters shared.TestRunFilter) (result []*datastore.Key, err error) {
	limit := filters.MaxCount
	from := filters.From
	if limit == nil && from == nil {
		// Default to a single, latest run when from & max-count both empty.
		one := 1
		limit = &one
	}
	products := filters.GetProductsOrDefault()

	// When ?aligned=true, make sure to show results for the same aligned run (executed for all browsers).
	if shared.IsLatest(filters.SHA) && filters.Aligned != nil && *filters.Aligned {
		shas, shaKeys, err := shared.GetAlignedRunSHAs(ctx, products, filters.Labels, from, filters.To, limit)
		if err != nil {
			return result, err
		}
		if len(shas) < 1 {
			// Bail out early - can't find any complete runs.
			return result, nil
		}
		keys := []*datastore.Key{}
		for _, sha := range shas {
			keys = append(keys, shaKeys[sha]...)
		}
		return keys, err
	}
	return shared.LoadTestRunKeys(ctx, products, filters.Labels, filters.SHA, from, filters.To, limit)
}

// LoadTestRunsForFilters deciphers the filters and executes a corresponding query to load
// the TestRuns.
func LoadTestRunsForFilters(ctx context.Context, filters shared.TestRunFilter) (result []shared.TestRun, err error) {
	var keys []*datastore.Key
	if keys, err = LoadTestRunKeysForFilters(ctx, filters); err != nil {
		return nil, err
	}
	return shared.LoadTestRunsByKeys(ctx, keys)
}
