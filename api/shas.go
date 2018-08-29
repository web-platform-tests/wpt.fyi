// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// apiSHAsHandler is responsible for emitting just the revision SHAs for test runs.
func apiSHAsHandler(w http.ResponseWriter, r *http.Request) {
	filters, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)

	var shas []string
	if filters.Complete != nil && *filters.Complete {
		if shas, err = getCompleteRunSHAs(ctx, filters.From, filters.To, filters.MaxCount); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		products := filters.GetProductsOrDefault()
		testRuns, err := shared.LoadTestRuns(ctx, products, filters.Labels, nil, filters.From, filters.To, filters.MaxCount)
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

// getCompleteRunSHAs returns an array of the SHA[0:10] for runs that
// exists for all initially-loaded browser names (see GetDefaultBrowserNames),
// ordered by most-recent.
func getCompleteRunSHAs(ctx context.Context, from, to *time.Time, limit *int) (shas []string, err error) {
	query := datastore.
		NewQuery("TestRun").
		Order("-TimeStart").
		Project("Revision", "BrowserName")

	browserNames := shared.GetDefaultBrowserNames()

	if from != nil {
		query = query.Filter("TimeStart >=", *from)
	}
	if to != nil {
		query = query.Filter("TimeStart <", *to)
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
		} else if !shared.IsStableBrowserName(testRun.BrowserName) {
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
