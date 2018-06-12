// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/deckarep/golang-set"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// apiSHAsHandler is responsible for emitting just the revision SHAs for test runs.
//
// URL Params:
//     sha: SHA[0:10] of the repo when the tests were executed (or 'latest')
func apiSHAsHandler(w http.ResponseWriter, r *http.Request) {
	var products []shared.Product
	var err error
	if products, err = shared.GetProductsForRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	limit, err := shared.ParseMaxCountParam(r)
	if err != nil {
		http.Error(w, "Invalid 'max-count' param: "+err.Error(), http.StatusBadRequest)
		return
	}
	var from *time.Time
	if from, err = shared.ParseFromParam(r); err != nil {
		http.Error(w, fmt.Sprintf("Invalid 'from' param: %s", err.Error()), http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)

	var shas []string
	if complete, err := strconv.ParseBool(r.URL.Query().Get("complete")); err == nil && complete {
		if shas, err = getCompleteRunSHAs(ctx, from, limit); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		labels := shared.ParseLabelsParam(r)
		testRuns, err := shared.LoadTestRuns(ctx, products, labels, shas, from, limit)
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
	shasBytes, err := json.Marshal(shas)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(shasBytes)
}

// getCompleteRunSHAs returns an array of the SHA[0:10] for runs that
// exists for all initially-loaded browser names (see GetBrowserNames),
// ordered by most-recent.
func getCompleteRunSHAs(ctx context.Context, from *time.Time, limit *int) (shas []string, err error) {
	query := datastore.
		NewQuery("TestRun").
		Order("-CreatedAt").
		Project("Revision", "BrowserName")

	var browserNames []string
	if browserNames, err = shared.GetBrowserNames(); err != nil {
		return nil, err
	}

	if from != nil {
		query = query.Filter("CreatedAt >=", *from)
	}

	bySHA := make(map[string]mapset.Set)
	done := mapset.NewSet()
	it := query.Run(ctx)
	for {
		var testRun shared.TestRun
		_, err := it.Next(&testRun)
		if err == datastore.Done {
			break
		} else if err != nil {
			return nil, err
		} else if !shared.IsBrowserName(testRun.BrowserName) {
			continue
		}
		set, ok := bySHA[testRun.Revision]
		if !ok {
			bySHA[testRun.Revision] = mapset.NewSetWith(testRun.BrowserName)
		} else {
			set.Add(testRun.BrowserName)
			if set.Cardinality() == len(browserNames) && !done.Contains(testRun.Revision) {
				done.Add(testRun.Revision)
				shas = append(shas, testRun.Revision)
				if limit != nil && len(shas) >= *limit {
					return shas, nil
				}
			}
		}
	}
	return shas, err
}
