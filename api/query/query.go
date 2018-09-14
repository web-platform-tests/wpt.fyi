// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/api"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// summary is the golang type for the JSON format in pass/total summary files.
type summary map[string][]int

type sharedInterface interface {
	ParseQueryParamInt(r *http.Request, key string) (int, error)
	ParseQueryFilterParams(*http.Request) (shared.QueryFilter, error)
	LoadTestRuns([]shared.ProductSpec, mapset.Set, []string, *time.Time, *time.Time, *int) ([]shared.TestRun, error)
	LoadTestRunsByIDs(ids shared.TestRunIDs) (result []shared.TestRun, err error)
	LoadTestRun(int64) (*shared.TestRun, error)
}

type defaultShared struct {
	ctx context.Context
}

func (defaultShared) ParseQueryParamInt(r *http.Request, key string) (int, error) {
	return shared.ParseQueryParamInt(r, key)
}

func (defaultShared) ParseQueryFilterParams(r *http.Request) (shared.QueryFilter, error) {
	return shared.ParseQueryFilterParams(r)
}

func (sharedImpl defaultShared) LoadTestRuns(ps []shared.ProductSpec, ls mapset.Set, shas []string, from *time.Time, to *time.Time, limit *int) ([]shared.TestRun, error) {
	return shared.LoadTestRuns(sharedImpl.ctx, ps, ls, shas, from, to, limit)
}

func (sharedImpl defaultShared) LoadTestRunsByIDs(ids shared.TestRunIDs) (result []shared.TestRun, err error) {
	return ids.LoadTestRuns(sharedImpl.ctx)
}

func (sharedImpl defaultShared) LoadTestRun(id int64) (*shared.TestRun, error) {
	return shared.LoadTestRun(sharedImpl.ctx, id)
}

type queryHandler struct {
	sharedImpl sharedInterface
	dataSource shared.CachedStore
}

func (qh queryHandler) processInput(w http.ResponseWriter, r *http.Request) (*shared.QueryFilter, []shared.TestRun, []summary, error) {
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

func (qh queryHandler) getRunsAndFilters(in shared.QueryFilter) ([]shared.TestRun, shared.QueryFilter, error) {
	filters := in
	var testRuns []shared.TestRun
	var err error

	if filters.RunIDs == nil || len(filters.RunIDs) == 0 {
		var runFilters shared.TestRunFilter
		var shas []string
		limit := 1
		products := runFilters.GetProductsOrDefault()
		testRuns, err = qh.sharedImpl.LoadTestRuns(products, runFilters.Labels, shas, runFilters.From, runFilters.To, &limit)
		if err != nil {
			return testRuns, filters, err
		}

		filters.RunIDs = make([]int64, 0, len(testRuns))
		for _, testRun := range testRuns {
			filters.RunIDs = append(filters.RunIDs, testRun.ID)
		}
	} else {
		testRuns, err = qh.sharedImpl.LoadTestRunsByIDs(shared.TestRunIDs(filters.RunIDs))
		if err != nil {
			return testRuns, filters, err
		}
	}

	return testRuns, filters, nil
}

func (qh queryHandler) loadSummaries(testRuns []shared.TestRun) ([]summary, error) {
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

func (qh queryHandler) loadSummary(testRun shared.TestRun) ([]byte, error) {
	mkey := getMemcacheKey(testRun)
	url := api.GetResultsURL(testRun, "")
	var data []byte
	err := qh.dataSource.Get(mkey, url, &data)
	return data, err
}

func getMemcacheKey(testRun shared.TestRun) string {
	return "RESULTS_SUMMARY-" + strconv.FormatInt(testRun.ID, 10)
}
