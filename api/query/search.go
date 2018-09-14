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
)

// LegacySearchRunResult is the results data from legacy test summarys.  These
// summaries contain a "pass count" and a "total count", where the test itself
// counts as 1, and each subtest counts as 1. The "pass count" contains any
// status values that are "PASS" or "OK".
type LegacySearchRunResult struct {
	// Passes is the number of test results in a PASS/OK state.
	Passes int `json:"passes"`
	// Total is the total number of test results for this run/file pair.
	Total int `json:"total"`
}

// SearchResult contains data regarding a particular test file over a collection
// of runs. The runs are identified externally in a parallel slice (see
// SearchResponse).
type SearchResult struct {
	// Test is the name of a test; this often corresponds to a test file path in
	// the WPT source reposiory.
	Test string `json:"test"`
	// LegacyStatus is the results data from legacy test summaries. These
	// summaries contain a "pass count" and a "total count", where the test itself
	// counts as 1, and each subtest counts as 1. The "pass count" contains any
	// status values that are "PASS" or "OK".
	LegacyStatus []LegacySearchRunResult `json:"legacy_status"`
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
func (r byName) Less(i, j int) bool { return r[i].Test < r[j].Test }

type searchHandler struct {
	queryHandler
}

func apiSearchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query params.
	ctx := shared.NewAppEngineContext(r)
	sh := searchHandler{queryHandler{
		sharedImpl: defaultShared{ctx},
		dataSource: shared.NewByteCachedStore(ctx, shared.NewGZReadWritable(shared.NewMemcacheReadWritable(ctx)), shared.NewHTTPReadable(ctx)),
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
	q := canonicalizeStr(filters.Q)
	// Dedup visited file names via a map of results.
	resMap := make(map[string]SearchResult)
	for i, s := range summaries {
		for filename, passAndTotal := range s {
			// Exclude filenames that do not match query.
			if !strings.Contains(canonicalizeStr(filename), q) {
				continue
			}

			if _, ok := resMap[filename]; !ok {
				resMap[filename] = SearchResult{
					Test:         filename,
					LegacyStatus: make([]LegacySearchRunResult, len(testRuns)),
				}
			}
			resMap[filename].LegacyStatus[i] = LegacySearchRunResult{
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
