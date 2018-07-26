// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// interopHandler handles the view of test results broken down by the
// number of browsers for which the test passes.
func interopHandler(w http.ResponseWriter, r *http.Request) {
	sourceURL, _ := url.Parse("/api/interop")
	f, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sourceURL.RawQuery = f.ToQuery(true).Encode()

	// We 'load by SHA' by fetching any interop result with all TestRunIDs for that SHA.
	if !filters.IsDefaultQuery() {
		// Load default browser runs for SHA.
		// Ignore any max-count; makes no sense for a interop run.
		one := 1
		runs, err := shared.LoadTestRuns(
			ctx, filters.GetProductsOrDefault(), filters.Labels, []string{filters.SHA}, filters.From, filters.To, &one)
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

	uiFilters, err := parseTestRunUIFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadataBytes, err := json.Marshal(*metadata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Metadata string
		Filter   testRunUIFilter
	}{
		Metadata: string(metadataBytes),
		Filter:   uiFilters,
	}

	if err := templates.ExecuteTemplate(w, "interoperability.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
