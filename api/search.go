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
	Get(ctx context.Context, id string) ([]byte, error)
}

type readWritable interface {
	readable
	Put(context.Context, string, []byte) error
}

type httpReadable struct{}

func (httpReadable) Get(ctx context.Context, url string) ([]byte, error) {
	r, err := http.Get(url)
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

type memcacheReadWritable struct{}

func (mc memcacheReadWritable) Get(ctx context.Context, key string) ([]byte, error) {
	item, err := memcache.Get(ctx, key)
	if item == nil {
		return nil, err
	}

	return item.Value, err
}

func (mc memcacheReadWritable) Put(ctx context.Context, key string, value []byte) error {
	return memcache.Add(ctx, &memcache.Item{
		Key:   key,
		Value: value,
	})
}

type sharedImpl interface {
	ParseSearchFilterParams(*http.Request) (shared.SearchFilter, error)
	LoadTestRuns(context.Context, []shared.ProductSpec, mapset.Set, []string, *time.Time, *time.Time, *int) ([]shared.TestRun, error)
	LoadTestRun(context.Context, int64) (*shared.TestRun, error)
}

type defaultSharedImpl struct{}

func (defaultSharedImpl) ParseSearchFilterParams(r *http.Request) (shared.SearchFilter, error) {
	return shared.ParseSearchFilterParams(r)
}

func (defaultSharedImpl) LoadTestRuns(ctx context.Context, ps []shared.ProductSpec, ls mapset.Set, shas []string, from *time.Time, to *time.Time, limit *int) ([]shared.TestRun, error) {
	return shared.LoadTestRuns(ctx, ps, ls, shas, from, to, limit)
}

func (defaultSharedImpl) LoadTestRun(ctx context.Context, id int64) (*shared.TestRun, error) {
	return shared.LoadTestRun(ctx, id)
}

type searchHandler struct {
	simpl sharedImpl
	cache readWritable
	store readable
}

func (sh searchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse query params.
	ctx := appengine.NewContext(r)
	filters, err := sh.simpl.ParseSearchFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	testRuns, filters, err := sh.getRunsAndFilters(ctx, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	summaries, err := sh.loadSummaries(ctx, testRuns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := sh.prepareResponse(filters, testRuns, summaries)

	// Send response.
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(data)
}

func (sh searchHandler) getRunsAndFilters(ctx context.Context, in shared.SearchFilter) ([]shared.TestRun, shared.SearchFilter, error) {
	filters := in
	var testRuns []shared.TestRun

	if filters.RunIDs == nil || len(filters.RunIDs) == 0 {
		var runFilters shared.TestRunFilter
		var shas []string
		var err error
		limit := 1
		products := runFilters.GetProductsOrDefault()
		testRuns, err = sh.simpl.LoadTestRuns(ctx, products, runFilters.Labels, shas, runFilters.From, runFilters.To, &limit)
		if err != nil {
			return testRuns, filters, err
		}

		filters.RunIDs = make([]int64, len(testRuns))
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
				testRun, err = sh.simpl.LoadTestRun(ctx, id)
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

func (sh searchHandler) loadSummaries(ctx context.Context, testRuns []shared.TestRun) ([]summary, error) {
	var err error
	summaries := make([]summary, len(testRuns))

	var wg sync.WaitGroup
	for i, testRun := range testRuns {
		wg.Add(1)

		go func(i int, testRun shared.TestRun) {
			var data []byte
			s := make(summary)
			data, err = sh.loadSummary(ctx, testRun)
			if err != nil {
				return
			}
			err = json.Unmarshal(data, &s)
			if err != nil {
				return
			}
			summaries[i] = s
		}(i, testRun)
	}
	wg.Wait()

	return summaries, err
}

func (sh searchHandler) loadSummary(ctx context.Context, testRun shared.TestRun) ([]byte, error) {
	mkey := getMemcacheKey(testRun)
	cached, err := sh.cache.Get(ctx, mkey)
	if cached != nil && err == nil {
		return cached, nil
	}

	if err != nil {
		log.Printf("WARNING: Error fetching cache key %s: %v", mkey, err)
	}

	url := getResultsURL(testRun, "")
	data, err := sh.store.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	// Cache summary.
	go func() {
		if err := sh.cache.Put(ctx, mkey, data); err != nil {
			log.Printf("WARNING: Failed to write TestRun summary to memcache key %s", mkey)
		}
	}()

	return data, nil
}

func (sh searchHandler) prepareResponse(filters shared.SearchFilter, testRuns []shared.TestRun, summaries []summary) SearchResponse {
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
