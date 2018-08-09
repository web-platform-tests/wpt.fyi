// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
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

// summary is the golang type for the JSON format in pass/total summary files.
type summary map[string][]int

type byName []SearchResult

func (r byName) Len() int           { return len(r) }
func (r byName) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byName) Less(i, j int) bool { return r[i].Name < r[j].Name }

type readable interface {
	Get(string) ([]byte, error)
}

type readWritable interface {
	readable
	Put(string, []byte) error
}

type httpReadable struct {
	ctx context.Context
}

func (hr httpReadable) Get(url string) ([]byte, error) {
	client := urlfetch.Client(hr.ctx)
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code from %s: %d", url, r.StatusCode)
	}

	var data []byte
	if data, err = ioutil.ReadAll(r.Body); err != nil {
		return nil, err
	}

	return data, nil
}

type memcacheReadWritable struct {
	ctx context.Context
}

func (mc memcacheReadWritable) Get(key string) ([]byte, error) {
	item, err := memcache.Get(mc.ctx, key)
	if item == nil {
		return nil, err
	}

	return item.Value, err
}

func (mc memcacheReadWritable) Put(key string, value []byte) error {
	return memcache.Add(mc.ctx, &memcache.Item{
		Key:   key,
		Value: value,
	})
}

type sharedImpl interface {
	ParseSearchFilterParams(*http.Request) (shared.SearchFilter, error)
	LoadTestRuns([]shared.ProductSpec, mapset.Set, []string, *time.Time, *time.Time, *int) ([]shared.TestRun, error)
	LoadTestRun(int64) (*shared.TestRun, error)
}

type defaultSharedImpl struct {
	ctx context.Context
}

func (defaultSharedImpl) ParseSearchFilterParams(r *http.Request) (shared.SearchFilter, error) {
	return shared.ParseSearchFilterParams(r)
}

func (simpl defaultSharedImpl) LoadTestRuns(ps []shared.ProductSpec, ls mapset.Set, shas []string, from *time.Time, to *time.Time, limit *int) ([]shared.TestRun, error) {
	return shared.LoadTestRuns(simpl.ctx, ps, ls, shas, from, to, limit)
}

func (simpl defaultSharedImpl) LoadTestRun(id int64) (*shared.TestRun, error) {
	return shared.LoadTestRun(simpl.ctx, id)
}

type searchHandler struct {
	simpl sharedImpl
	cache readWritable
	store readable
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	// Parse query params.
	ctx := appengine.NewContext(r)
	sh := searchHandler{
		simpl: defaultSharedImpl{ctx},
		cache: memcacheReadWritable{ctx},
		store: httpReadable{},
	}
	sh.ServeHTTP(w, r)
}

func (sh searchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters, err := sh.simpl.ParseSearchFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	testRuns, filters, err := sh.getRunsAndFilters(filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	summaries, err := sh.loadSummaries(testRuns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := prepareResponse(filters, testRuns, summaries)

	// Send response.
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(data)
}

func (sh searchHandler) getRunsAndFilters(in shared.SearchFilter) ([]shared.TestRun, shared.SearchFilter, error) {
	filters := in
	var testRuns []shared.TestRun

	if filters.RunIDs == nil || len(filters.RunIDs) == 0 {
		var runFilters shared.TestRunFilter
		var shas []string
		var err error
		limit := 1
		products := runFilters.GetProductsOrDefault()
		testRuns, err = sh.simpl.LoadTestRuns(products, runFilters.Labels, shas, runFilters.From, runFilters.To, &limit)
		if err != nil {
			return testRuns, filters, err
		}

		filters.RunIDs = make([]int64, 0, len(testRuns))
		for _, testRun := range testRuns {
			filters.RunIDs = append(filters.RunIDs, testRun.ID)
		}
	} else {
		var err error
		var wg sync.WaitGroup
		testRuns = make([]shared.TestRun, len(filters.RunIDs))
		for i, id := range filters.RunIDs {
			wg.Add(1)
			go func(i int, id int64) {
				defer wg.Done()

				var testRun *shared.TestRun
				testRun, err = sh.simpl.LoadTestRun(id)
				if err == nil {
					testRuns[i] = *testRun
				}
			}(i, id)
		}
		wg.Wait()

		if err != nil {
			return testRuns, filters, err
		}
	}

	return testRuns, filters, nil
}

func (sh searchHandler) loadSummaries(testRuns []shared.TestRun) ([]summary, error) {
	var err error
	summaries := make([]summary, len(testRuns))

	var wg sync.WaitGroup
	for i, testRun := range testRuns {
		wg.Add(1)

		go func(i int, testRun shared.TestRun) {
			defer wg.Done()

			var data []byte
			s := make(summary)
			data, loadErr := sh.loadSummary(testRun)
			if err == nil && loadErr != nil {
				err = loadErr
				return
			}
			marshalErr := json.Unmarshal(data, &s)
			if err == nil && marshalErr != nil {
				err = marshalErr
				return
			}
			summaries[i] = s
		}(i, testRun)
	}
	wg.Wait()

	return summaries, err
}

func (sh searchHandler) loadSummary(testRun shared.TestRun) ([]byte, error) {
	mkey := getMemcacheKey(testRun)
	cached, err := sh.cache.Get(mkey)
	if cached != nil && err == nil {
		return cached, nil
	}

	if err != nil {
		log.Printf("WARNING: Error fetching cache key %s: %v", mkey, err)
		err = nil
	}

	url := getResultsURL(testRun, "")
	data, err := sh.store.Get(url)
	if err != nil {
		return nil, err
	}

	// Cache summary.
	go func() {
		if err := sh.cache.Put(mkey, data); err != nil {
			log.Printf("WARNING: Failed to write TestRun summary to memcache key %s", mkey)
		}
	}()

	return data, nil
}

func prepareResponse(filters shared.SearchFilter, testRuns []shared.TestRun, summaries []summary) SearchResponse {
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

func getMemcacheKey(testRun shared.TestRun) string {
	return "RESULTS_SUMMARY-" + getResultsURL(testRun, "")
}
