// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"encoding/json"
	http "net/http"
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set"
	shared "github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
)

// AutocompleteResult contains a single autocomplete suggestion.
type AutocompleteResult struct {
	// String represents the most basic form of an autocomplete result. It is the
	// complete string being recommended. The client is responsibile for fitting
	// it to any substring(s) already appearing in a search UI.
	String string `json:"string"`
}

// AutocompleteResponse contains a response to autocmplete API calls.
type AutocompleteResponse struct {
	// Results is the collection of test results, grouped by test file name.
	Results []AutocompleteResult `json:"results"`
}

type byQueryIndex struct {
	q  string
	rs []AutocompleteResult
}

func (r byQueryIndex) Len() int      { return len(r.rs) }
func (r byQueryIndex) Swap(i, j int) { r.rs[i], r.rs[j] = r.rs[j], r.rs[i] }
func (r byQueryIndex) Less(i, j int) bool {
	return strings.Index(r.rs[i].String, r.q) < strings.Index(r.rs[j].String, r.q)
}

type autocompleteHandler struct {
	queryHandler
}

func apiAutocompleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	sh := autocompleteHandler{queryHandler{
		sharedImpl: defaultShared{ctx},
		dataSource: cachedStore{
			cache: gzipReadWritable{memcacheReadWritable{ctx}},
			store: httpReadable{ctx},
		},
	}}
	sh.ServeHTTP(w, r)
}

func (ah autocompleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters, testRuns, summaries, err := ah.processInput(w, r)
	// processInput handles writing any error to w.
	if err != nil {
		return
	}

	resp := prepareAutocompleteResponse(filters, testRuns, summaries)

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(data)
}

func prepareAutocompleteResponse(filters *shared.QueryFilter, testRuns []shared.TestRun, summaries []summary) AutocompleteResponse {
	fileSet := mapset.NewSet()
	for _, smry := range summaries {
		for file := range smry {
			fileSet.Add(file)
		}
	}

	files := make([]AutocompleteResult, 0, fileSet.Cardinality()/len(testRuns))
	for fileInterface := range fileSet.Iter() {
		file := fileInterface.(string)
		if strings.Contains(file, filters.Q) {
			files = append(files, AutocompleteResult{file})
		}
	}

	sortable := byQueryIndex{
		q:  filters.Q,
		rs: files,
	}
	sort.Sort(sortable)
	return AutocompleteResponse{sortable.rs}
}
