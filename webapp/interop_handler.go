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

// interopHandler handles the view of test results broken down by the
// number of browsers for which the test passes.
func interopHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	query := datastore.
		NewQuery(metrics.GetDatastoreKindName(metrics.PassRateMetadata{})).
		Order("-StartTime").
		Limit(1)
	var metadataSlice []metrics.PassRateMetadata

	if _, err := query.GetAll(ctx, &metadataSlice); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(metadataSlice) != 1 {
		http.Error(w, "No metrics runs found", http.StatusInternalServerError)
		return
	}

	metadataBytes, err := json.Marshal(metadataSlice[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Metadata string
	}{
		string(metadataBytes),
	}

	if err := templates.ExecuteTemplate(w, "interoperability.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
