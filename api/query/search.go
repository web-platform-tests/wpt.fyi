// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	time "time"

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
	LegacyStatus []LegacySearchRunResult `json:"legacy_status,omitempty"`

	// Interoperability scores. For N browsers, we have an array of
	// N+1 items, where the index X is the number of items passing in exactly
	// X of the N browsers. e.g. for 4 browsers, [0/4, 1/4, 2/4, 3/4, 4/4].
	Interop []int `json:"interop,omitempty"`

	// Subtests (names) which are included in the LegacyStatus summary.
	Subtests []string `json:"subtests,omitempty"`

	// Diff count of subtests which are included in the LegacyStatus summary.
	Diff shared.TestDiff `json:"diff,omitempty"`
}

// SearchResponse contains a response to search API calls, including specific
// runs whose results were searched and the search results themselves.
type SearchResponse struct {
	// Runs is the specific runs for which results were retrieved. Each run, in
	// order, corresponds to a Status entry in each SearchResult in Results.
	Runs []shared.TestRun `json:"runs"`
	// IgnoredRuns is any runs that the client requested to be included in the
	// query, but were not included. This optional field may be non-nil if, for
	// example, results are being served from an incompelte cache of runs and some
	// runs described in the query request are not resident in the cache.
	IgnoredRuns []shared.TestRun `json:"ignored_runs,omitempty"`
	// Results is the collection of test results, grouped by test file name.
	Results []SearchResult `json:"results"`
	// MetadataResponse is a response to a wpt-metadata query.
	MetadataResult shared.MetadataResponse `json:"metadata,omitempty"`
}

type byName []SearchResult

func (r byName) Len() int           { return len(r) }
func (r byName) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byName) Less(i, j int) bool { return r[i].Test < r[j].Test }

type searchHandler struct {
	api shared.AppEngineAPI
}

type unstructuredSearchHandler struct {
	queryHandler
}

type structuredSearchHandler struct {
	queryHandler

	api shared.AppEngineAPI
}

func apiSearchHandler(w http.ResponseWriter, r *http.Request) {
	api := shared.NewAppEngineAPI(shared.NewAppEngineContext(r))
	searchHandler{api}.ServeHTTP(w, r)
}

func (sh searchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Invalid HTTP method", http.StatusBadRequest)
		return
	}

	ctx := sh.api.Context()
	mc := shared.NewGZReadWritable(shared.NewMemcacheReadWritable(ctx, 48*time.Hour))
	qh := queryHandler{
		store:      shared.NewAppEngineDatastore(ctx, true),
		sharedImpl: defaultShared{ctx},
		dataSource: shared.NewByteCachedStore(ctx, mc, shared.NewHTTPReadable(ctx)),
		client:     sh.api.GetHTTPClient(),
	}
	var delegate http.Handler
	if r.Method == "GET" {
		delegate = unstructuredSearchHandler{queryHandler: qh}
	} else {
		delegate = structuredSearchHandler{queryHandler: qh, api: sh.api}
	}
	ch := shared.NewCachingHandler(ctx, delegate, mc, isRequestCacheable, cacheKey, shouldCacheSearchResponse)
	ch.ServeHTTP(w, r)
}

func (sh structuredSearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
	}
	err = r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to finish reading request body", http.StatusInternalServerError)
	}

	var rq RunQuery
	err = json.Unmarshal(data, &rq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the query is a simple (empty/just True, or test name only) query
	var simpleQ TestNamePattern
	var isSimpleQ bool
	{
		if _, isTrueQ := rq.AbstractQuery.(True); isTrueQ {
			isSimpleQ = true
		} else if exists, isExists := rq.AbstractQuery.(AbstractExists); isExists && len(exists.Args) == 1 {
			simpleQ, isSimpleQ = exists.Args[0].(TestNamePattern)
		}
		q := r.URL.Query()
		_, interop := q["interop"]
		_, subtests := q["subtests"]
		_, diff := q["diff"]
		isSimpleQ = isSimpleQ && !interop && !subtests && !diff
	}

	if !isSimpleQ {
		ctx := sh.api.Context()
		hostname := sh.api.GetServiceHostname("searchcache")
		// TODO: This will not work when hostname is localhost (http scheme needed).
		fwdURL, _ := url.Parse(fmt.Sprintf("https://%s/api/search/cache", hostname))
		fwdURL.RawQuery = r.URL.RawQuery

		logger := shared.GetLogger(ctx)
		logger.Infof("Forwarding structured search request to %s: %s", hostname, string(data))

		client := sh.api.GetHTTPClient()
		req, err := http.NewRequest("POST", fwdURL.String(), bytes.NewBuffer(data))
		if err != nil {
			logger.Errorf("Failed to create request to POST %s: %v", fwdURL.String(), err)
			http.Error(w, "Error connecting to search API cache", http.StatusInternalServerError)
			return
		}
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			logger.Errorf("Error connecting to search API cache: %v", err)
			http.Error(w, "Error connecting to search API cache", http.StatusInternalServerError)
			return
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			msg := fmt.Sprintf("Error from request: POST %s: STATUS %d", fwdURL.String(), resp.StatusCode)
			errBody, err2 := ioutil.ReadAll(resp.Body)
			if err2 == nil {
				msg = fmt.Sprintf("%s: %s", msg, string(errBody))
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(errBody))
			}
			logger.Errorf(msg)
		}

		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			logger.Errorf("Error forwarding response payload from search cache: %v", err)
		}
		return
	}

	// Structured query is equivalent to unstructured query.
	// Create an unstructured query request and delegate to unstructured query
	// handler.
	oldQ := r.URL.Query()
	r2 := *r
	r2url := *r.URL
	r2.URL = &r2url
	r2.Method = "GET"
	runIDStrs := make([]string, 0, len(rq.RunIDs))
	for _, id := range rq.RunIDs {
		runIDStrs = append(runIDStrs, strconv.FormatInt(id, 10))
	}
	runIDsStr := strings.Join(runIDStrs, ",")
	r2.URL.RawQuery = fmt.Sprintf("run_ids=%s&q=%s", url.QueryEscape(runIDsStr), url.QueryEscape(simpleQ.Pattern))

	if _, showMetadata := oldQ["metadataInfo"]; showMetadata {
		r2.URL.RawQuery = fmt.Sprintf("%s&metadataInfo=%s", r2.URL.RawQuery, oldQ["metadataInfo"])
	}

	unstructuredSearchHandler{queryHandler: sh.queryHandler}.ServeHTTP(w, &r2)
}

func (sh unstructuredSearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters, testRuns, summaries, err := sh.processInput(w, r)
	// processInput handles writing any error to w.
	if err != nil {
		return
	}

	resp := prepareSearchResponse(filters, testRuns, summaries)

	q := r.URL.Query()
	if _, showMetadata := q["metadataInfo"]; showMetadata {
		resp.MetadataResult = shared.GetMetadataResponse(testRuns, sh.queryHandler.client)
	}

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

var cacheKey = func(r *http.Request) interface{} {
	if r.Method == "GET" {
		return shared.URLAsCacheKey(r)
	}

	body := r.Body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Failed to read non-GET request body for generating cache key: %v", err)
		shared.GetLogger(shared.NewAppEngineContext(r)).Errorf(msg)
		panic(msg)
	}
	defer body.Close()

	// Ensure that r.Body can be read again by other request handling routines.
	r.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	return fmt.Sprintf("%s#%s", r.URL.String(), string(data))
}

// TODO: Sometimes an empty result set is being cached for a query over
// legitimate runs. For now, prevent serving empty result sets from cache.
// Eventually, a more durable fix to
// https://github.com/web-platform-tests/wpt.fyi/issues/759 should replace this
// approximation.
var shouldCacheSearchResponse = func(ctx context.Context, statusCode int, payload []byte) bool {
	if !shared.CacheStatusOK(ctx, statusCode, payload) {
		return false
	}

	var resp SearchResponse
	err := json.Unmarshal(payload, &resp)
	if err != nil {
		shared.GetLogger(ctx).Errorf("Malformed search response")
		return false
	}

	if len(resp.Results) == 0 {
		shared.GetLogger(ctx).Errorf("Query yielded no results; not caching")
		return false
	}

	return true
}
