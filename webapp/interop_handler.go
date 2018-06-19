// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"net/http"

	"github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// interopHandler handles the view of test results broken down by the
// number of browsers for which the test passes.
func interopHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	passRateType := metrics.GetDatastoreKindName(metrics.PassRateMetadata{})
	query := datastore.NewQuery(passRateType).Order("-StartTime").Limit(1)

	sha, err := shared.ParseSHAParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Try load by SHA, otherwise fall back to latest.
	if !shared.IsLatest(sha) {
		// Load default browser runs for SHA.
		runs, err := shared.LoadTestRuns(
			ctx, shared.GetDefaultProducts(), nil, []string{sha}, nil, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
	data := struct {
		Metadata string
		SHA      string
	}{
		string(metadataBytes),
		sha,
	}

	if err := templates.ExecuteTemplate(w, "interoperability.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
