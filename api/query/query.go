// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// SummaryResult is the format of the data from summary files generated with the newest aggregation method.
type SummaryResult struct {
	// Status represents the 1-2 character abbreviation for the status of the test.
	Status string `json:"s"`
	// Counts represents the subtest counts (passes and total).
	Counts []int `json:"c"`
}

// summary is the golang type for the JSON format in pass/total summary files.
// It has an old structure and a new structure - each which represent summary files
// that match the old or new summary format.
type summary struct {
	// oldFormat This holds summary information if the data is aggregated with the old method.
	// TODO (danielrsmith): This format should be removed once old summary files are invalidated.
	oldFormat map[string][]int
	// newFormat This holds summary information if the data is aggregated with the new method.
	newFormat map[string]SummaryResult
}
type queryHandler struct {
	store      shared.Datastore
	dataSource shared.CachedStore
	client     *http.Client
	logger     shared.Logger
}

func (qh queryHandler) processInput(w http.ResponseWriter, r *http.Request) (*shared.QueryFilter, shared.TestRuns, []summary, error) {
	filters, err := shared.ParseQueryFilterParams(r.URL.Query())
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
			s := summary{
				oldFormat: nil,
				newFormat: nil,
			}
			data, loadErr := qh.loadSummary(testRun)
			if err == nil && loadErr != nil {
				err = fmt.Errorf("Failed to load test run %v: %s", testRun.ID, loadErr.Error())
				return
			}
			// Try to unmarshal the json using the new aggregation structure.
			marshalErr := json.Unmarshal(data, &s.newFormat)
			if err == nil && marshalErr != nil {
				// If that failed, this is likely an old summary format.
				// Umarshal using the old structure.
				oldMarshalErr := json.Unmarshal(data, &s.oldFormat)
				if oldMarshalErr != nil {
					err = oldMarshalErr
					return
				}
			}
			summaries[i] = s
		}(i, testRun)
	}
	wg.Wait()

	return summaries, err
}

func (qh queryHandler) loadSummary(testRun shared.TestRun) ([]byte, error) {
	mkey := getRedisKey(testRun)
	url := shared.GetResultsURL(testRun, "")
	var data []byte
	err := qh.dataSource.Get(mkey, url, &data)
	return data, err
}

func getRedisKey(testRun shared.TestRun) string {
	return "RESULTS_SUMMARY-" + strconv.FormatInt(testRun.ID, 10)
}

func isRequestCacheable(r *http.Request) bool {
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
