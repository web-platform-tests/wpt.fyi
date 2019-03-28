// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// summary is the golang type for the JSON format in pass/total summary files.
type summary map[string][]int

type sharedInterface interface {
	ParseQueryParamInt(r *http.Request, key string) (*int, error)
	ParseQueryFilterParams(*http.Request) (shared.QueryFilter, error)
}

type defaultShared struct {
	ctx context.Context
}

func (defaultShared) ParseQueryParamInt(r *http.Request, key string) (*int, error) {
	return shared.ParseQueryParamInt(r.URL.Query(), key)
}

func (defaultShared) ParseQueryFilterParams(r *http.Request) (shared.QueryFilter, error) {
	return shared.ParseQueryFilterParams(r.URL.Query())
}

type queryHandler struct {
	store      shared.Datastore
	sharedImpl sharedInterface
	dataSource shared.CachedStore
	client     *http.Client
}

func (qh queryHandler) processInput(w http.ResponseWriter, r *http.Request) (*shared.QueryFilter, shared.TestRuns, []summary, error) {
	filters, err := qh.sharedImpl.ParseQueryFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, nil, nil, err
	}

	testRuns, filters, err := qh.getRunsAndFilters(filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return nil, nil, nil, err
	}

	summaries, err := qh.loadSummaries(testRuns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, nil, nil, err
	}

	return &filters, testRuns, summaries, nil
}

func (qh queryHandler) getRunsAndFilters(in shared.QueryFilter) (shared.TestRuns, shared.QueryFilter, error) {
	filters := in
	var testRuns shared.TestRuns
	var err error

	if filters.RunIDs == nil || len(filters.RunIDs) == 0 {
		var runFilters shared.TestRunFilter
		var sha string
		var err error
		limit := 1
		products := runFilters.GetProductsOrDefault()
		runsByProduct, err := qh.store.TestRunQuery().LoadTestRuns(
			products, runFilters.Labels, []string{sha}, runFilters.From, runFilters.To, &limit, nil)
		if err != nil {
			return testRuns, filters, err
		}

		testRuns = runsByProduct.AllRuns()
		filters.RunIDs = make([]int64, 0, len(testRuns))
		for _, testRun := range testRuns {
			filters.RunIDs = append(filters.RunIDs, testRun.ID)
		}
	} else {
		ids := shared.TestRunIDs(filters.RunIDs)
		testRuns = make(shared.TestRuns, len(ids))
		err = qh.store.GetMulti(ids.GetKeys(qh.store), testRuns)
		if err != nil {
			return testRuns, filters, err
		}
		testRuns.SetTestRunIDs(ids)
	}

	return testRuns, filters, nil
}

func (qh queryHandler) loadSummaries(testRuns shared.TestRuns) ([]summary, error) {
	var err error
	summaries := make([]summary, len(testRuns))

	var wg sync.WaitGroup
	for i, testRun := range testRuns {
		wg.Add(1)

		go func(i int, testRun shared.TestRun) {
			defer wg.Done()

			var data []byte
			s := make(summary)
			data, loadErr := qh.loadSummary(testRun)
			if err == nil && loadErr != nil {
				err = fmt.Errorf("Failed to load test run %v: %s", testRun.ID, loadErr.Error())
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

func (qh queryHandler) loadSummary(testRun shared.TestRun) ([]byte, error) {
	mkey := getMemcacheKey(testRun)
	url := shared.GetResultsURL(testRun, "")
	var data []byte
	err := qh.dataSource.Get(mkey, url, &data)
	return data, err
}

func getMemcacheKey(testRun shared.TestRun) string {
	return "RESULTS_SUMMARY-" + strconv.FormatInt(testRun.ID, 10)
}

func isRequestCacheable(r *http.Request) bool {
	q := r.URL.Query()
	if _, showMetadata := q["metadataInfo"]; showMetadata {
		return false
	}

	if r.Method == "GET" {
		ids, err := shared.ParseRunIDsParam(r.URL.Query())
		return err == nil && len(ids) > 0
	}

	if r.Method == "POST" {
		ids, err := shared.ExtractRunIDsBodyParam(r, true)
		return err == nil && len(ids) > 0
	}

	return false
}
