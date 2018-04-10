// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"github.com/w3c/wptdashboard/metrics"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"net/http"
)

type AnomalyData struct {
	Metadata string
	Browser  string
}

// anomalyHandler handles the view of test results showing which tests pass in
// some, but not all, browsers.
func anomalyHandler(w http.ResponseWriter, r *http.Request) {
	browser, err := ParseBrowserParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if browser != "" {
		browserAnomalyHandler(w, r, browser)
		return
	}

	ctx := appengine.NewContext(r)
	query := datastore.
		NewQuery(metrics.GetDatastoreKindName(
			metrics.PassRateMetadata{})).
		Order("-StartTime").Limit(1)
	var metadataSlice []metrics.PassRateMetadata

	if _, err := query.GetAll(ctx, &metadataSlice); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(metadataSlice) != 1 {
		http.Error(w, "No metrics runs found",
			http.StatusInternalServerError)
		return
	}

	metadataBytes, err := json.Marshal(metadataSlice[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := AnomalyData{
		Metadata: string(metadataBytes),
	}

	if err := templates.ExecuteTemplate(w, "anomalies.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// browserAnomalyHandler handles the view of test results showing which tests
// fail in a specific browser, but pass in at least one other browser.
func browserAnomalyHandler(w http.ResponseWriter, r *http.Request, browser string) {
	ctx := appengine.NewContext(r)
	query := datastore.
		NewQuery(metrics.GetDatastoreKindName(
			metrics.FailuresMetadata{})).
		Order("-StartTime").
		Filter("BrowserName =", browser).
		Limit(1)
	var metadataSlice []metrics.FailuresMetadata

	if _, err := query.GetAll(ctx, &metadataSlice); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(metadataSlice) != 1 {
		http.Error(w, "No metrics runs found",
			http.StatusInternalServerError)
		return
	}

	metadataBytes, err := json.Marshal(metadataSlice[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := AnomalyData{
		Metadata: string(metadataBytes),
		Browser:  browser,
	}

	if err := templates.ExecuteTemplate(w, "anomalies.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
