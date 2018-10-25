// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/datastore"
)

// apiVersionsHandler is responsible for emitting just the browser versions for the test runs.
func apiVersionsHandler(w http.ResponseWriter, r *http.Request) {
	product, err := shared.ParseProductParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if product == nil {
		http.Error(w, fmt.Sprintf("Invalid product param: %s", r.URL.Query().Get("product")), http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	query := datastore.NewQuery("TestRun").Filter("BrowserName =", product.BrowserName)
	var queries []*datastore.Query
	if product.BrowserVersion == "" {
		queries = []*datastore.Query{query}
	} else {
		queries = []*datastore.Query{
			query.Filter("BrowserVersion =", product.BrowserVersion).Limit(1),
			shared.VersionPrefix(query, "BrowserVersion", product.BrowserVersion, false /*desc*/).
				Project("BrowserVersion").
				Distinct(),
		}
	}

	var runs shared.TestRuns
	for _, query := range queries {
		var someRuns shared.TestRuns
		if _, err := query.GetAll(ctx, &someRuns); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		runs = append(runs, someRuns...)
	}

	if len(runs) < 1 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("[]"))
		return
	}

	versions := make([]string, len(runs))
	for i := range runs {
		versions[i] = runs[i].BrowserVersion
	}
	// TODO(lukebjerring): Fix this, it will put 100 before 11..., etc.
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))

	versionsBytes, err := json.Marshal(versions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(versionsBytes)
}
