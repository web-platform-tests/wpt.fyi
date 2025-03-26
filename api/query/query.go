// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// SummaryResult is the format of the data from summary files generated with the
// newest aggregation method.
type SummaryResult struct {
	// Status represents the 1-2 character abbreviation for the status of the test.
	Status string `json:"s"`
	// Counts represents the subtest counts (passes and total).
	Counts []int `json:"c"`
}

// summary is the golang type for the JSON format in pass/total summary files.
type summary map[string]SummaryResult

type queryHandler struct {
	store      shared.Datastore
	dataSource shared.CachedStore
}

// ErrBadSummaryVersion occurs when the summary file URL is not the correct version.
var ErrBadSummaryVersion = errors.New("invalid/unsupported summary version")

func (qh queryHandler) processInput(w http.ResponseWriter, r *http.Request) (
	*shared.QueryFilter,
	shared.TestRuns,
	[]summary,
	error,
) {
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

func (qh queryHandler) validateSummaryVersions(v url.Values, logger shared.Logger) error {
	filters, err := shared.ParseQueryFilterParams(v)
	if err != nil {
		return err
	}
	testRuns, _, err := qh.getRunsAndFilters(filters)
	if err != nil {
		return err
	}

	for _, testRun := range testRuns {
		summaryURL := shared.GetResultsURL(testRun, "")
		if !qh.summaryIsValid(summaryURL) {
			logger.Infof("summary URL has invalid suffix: %s", summaryURL)

			return fmt.Errorf("%w for URL %s", ErrBadSummaryVersion, summaryURL)
		}
	}

	return nil
}

func (qh queryHandler) summaryIsValid(summaryURL string) bool {
	// All new summary URLs end with "-summary_v2.json.gz". Any others are invalid.
	return strings.HasSuffix(summaryURL, "-summary_v2.json.gz")
}

func (qh queryHandler) getRunsAndFilters(in shared.QueryFilter) (shared.TestRuns, shared.QueryFilter, error) {
	filters := in
	var testRuns shared.TestRuns
	var err error

	if len(filters.RunIDs) == 0 {
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
			s := summary{}
			data, loadErr := qh.loadSummary(testRun)
			if err == nil && loadErr != nil {
				err = fmt.Errorf("failed to load test run %v: %s", testRun.ID, loadErr.Error())

				return
			}
			// Try to unmarshal the json using the new aggregation structure.
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
	mkey := getSummaryFileRedisKey(testRun)
	url := shared.GetResultsURL(testRun, "")
	var data []byte
	err := qh.dataSource.Get(mkey, url, &data)

	return data, err
}

func getSummaryFileRedisKey(testRun shared.TestRun) string {
	return "RESULTS_SUMMARY_v2-" + strconv.FormatInt(testRun.ID, 10)
}

func isRequestCacheable(r *http.Request) bool {
	if r.Method == http.MethodGet {
		ids, err := shared.ParseRunIDsParam(r.URL.Query())

		return err == nil && len(ids) > 0
	}

	if r.Method == http.MethodPost {
		ids, err := shared.ExtractRunIDsBodyParam(r, true)

		return err == nil && len(ids) > 0
	}

	return false
}
