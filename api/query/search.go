// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
)

// SearchRunResult is the metadata associated with a particular
// (test run, test file) pair.
type SearchRunResult struct {
	// Passes is the number of test results in a PASS/OK state.
	Passes int `json:"passes"`
	// Total is the total number of test results for this run/file pair.
	Total int `json:"total"`
}

// SearchResult contains data regarding a particular test file over a collection
// of runs. The runs are identified externally in a parallel slice (see
// SearchResponse).
type SearchResult struct {
	// Name is the full path of the test file.
	Name string `json:"name"`
	// Status is the results data for this file for each relevant run.
	Status []SearchRunResult `json:"status"`
}

// SearchResponse contains a response to search API calls, including specific
// runs whose results were searched and the search results themselves.
type SearchResponse struct {
	// Runs is the specific runs for which results were retrieved. Each run, in
	// order, corresponds to a Status entry in each SearchResult in Results.
	Runs []shared.TestRun `json:"runs"`
	// Results is the collection of test results, grouped by test file name.
	Results []SearchResult `json:"results"`
}

type byName []SearchResult

func (r byName) Len() int           { return len(r) }
func (r byName) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byName) Less(i, j int) bool { return r[i].Name < r[j].Name }

type searchHandler struct {
	queryHandler
}

func apiSearchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query params.
	ctx := appengine.NewContext(r)
	sh := searchHandler{queryHandler{
		sharedImpl: defaultShared{ctx},
		dataSource: cachedStore{
			cache: gzipReadWritable{memcacheReadWritable{ctx}},
			store: httpReadable{ctx},
		},
	}}
	sh.ServeHTTP(w, r)
}

func (sh searchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters, testRuns, summaries, err := sh.processInput(w, r)
	// processInput handles writing any error to w.
	if err != nil {
		return
	}

	resp := prepareSearchResponse(filters, testRuns, summaries)

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(data)
}

func prepareSearchResponse(filters *shared.QueryFilter, testRuns []shared.TestRun, summaries []summary) SearchResponse {
	resp := SearchResponse{
		Runs: testRuns,
	}
	// Dedup visited file names via a map of results.
	resMap := make(map[string]SearchResult)
	for i, s := range summaries {
		for filename, passAndTotal := range s {
			// Exclude filenames that do not match query.
			if !strings.Contains(filename, filters.Q) {
				continue
			}

			if _, ok := resMap[filename]; !ok {
				resMap[filename] = SearchResult{
					Name:   filename,
					Status: make([]SearchRunResult, len(testRuns)),
				}
			}
			resMap[filename].Status[i] = SearchRunResult{
				Passes: passAndTotal[0],
				Total:  passAndTotal[1],
			}
		}
	}
	// Load map into slice and sort it.
	resp.Results = make([]SearchResult, 0, len(resMap))
	for _, r := range resMap {
		resp.Results = append(resp.Results, r)
	}
	sort.Sort(byName(resp.Results))

	return resp
}
