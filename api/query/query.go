// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/api"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

var browsers = []string{
	"chrome",
	"edge",
	"firefox",
	"safari",
}

// TODO: Establish query execution plans over internal data schema.
type plan interface{}

type query interface {
	toPlan([]int64) plan
}

type runQuery struct {
	runIDs []int64
	query
}

func (rq runQuery) toPlan() plan {
	return rq.query.toPlan(rq.runIDs)
}

type testNamePattern struct {
	pattern string
}

func (tnp testNamePattern) toPlan(runIDs []int64) plan {
	return nil
}

type testStatusConstraint struct {
	browserName string
	status      int64
}

func (tsc testStatusConstraint) toPlan(runIDs []int64) plan {
	return nil
}

type not struct {
	not query
}

func (n not) toPlan(runIDs []int64) plan {
	return nil
}

type or struct {
	or []query
}

func (o or) toPlan(runIDs []int64) plan {
	return nil
}

type and struct {
	and []query
}

func (a and) toPlan(runIDs []int64) plan {
	return nil
}

func (rq *runQuery) UnmarshalJSON(b []byte) error {
	var data struct {
		RunIDs []int64         `json:"run_ids"`
		Query  json.RawMessage `json:"query"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if len(data.RunIDs) == 0 {
		return errors.New(`Missing run query property: "run_ids"`)
	}
	if len(data.Query) == 0 {
		return errors.New(`Missing run query property: "query"`)
	}

	q, err := unmarshalQ(data.Query)
	if err != nil {
		return err
	}

	rq.runIDs = data.RunIDs
	rq.query = q
	return nil
}

func (tnp *testNamePattern) UnmarshalJSON(b []byte) error {
	var data struct {
		Pattern string `json:"pattern"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if len(data.Pattern) == 0 {
		return errors.New(`Missing testn mae pattern property: "pattern"`)
	}

	tnp.pattern = data.Pattern
	return nil
}

func (tsc *testStatusConstraint) UnmarshalJSON(b []byte) error {
	var data struct {
		BrowserName string `json:"browser_name"`
		Status      string `json:"status"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if len(data.BrowserName) == 0 {
		return errors.New(`Missing test status constraint property: "browser_name"`)
	}
	if len(data.Status) == 0 {
		return errors.New(`Missing test status constraint property: "status"`)
	}

	browserName := strings.ToLower(data.BrowserName)
	browserNameOK := false
	for _, name := range browsers {
		browserNameOK = browserNameOK || browserName == name
	}
	if !browserNameOK {
		return fmt.Errorf(`Invalid browser name: "%s"`, data.BrowserName)
	}

	statusStr := strings.ToUpper(data.Status)
	status := shared.TestStatusValueFromString(statusStr)
	statusStr2 := shared.TestStatusStringFromValue(status)
	if statusStr != statusStr2 {
		return fmt.Errorf(`Invalid test status: "%s"`, data.Status)
	}

	tsc.browserName = browserName
	tsc.status = status
	return nil
}

func (n *not) UnmarshalJSON(b []byte) error {
	var data struct {
		Not json.RawMessage `json:"not"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if len(data.Not) == 0 {
		return errors.New(`Missing negation property: "not"`)
	}

	q, err := unmarshalQ(data.Not)
	n.not = q
	return err
}

func (o *or) UnmarshalJSON(b []byte) error {
	var data struct {
		Or []json.RawMessage `json:"or"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if len(data.Or) == 0 {
		return errors.New(`Missing disjunction property: "or"`)
	}

	qs := make([]query, 0, len(data.Or))
	for _, msg := range data.Or {
		q, err := unmarshalQ(msg)
		if err != nil {
			return err
		}
		qs = append(qs, q)
	}
	o.or = qs
	return nil
}

func (a *and) UnmarshalJSON(b []byte) error {
	var data struct {
		And []json.RawMessage `json:"and"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if len(data.And) == 0 {
		return errors.New(`Missing conjunction property: "and"`)
	}

	qs := make([]query, 0, len(data.And))
	for _, msg := range data.And {
		q, err := unmarshalQ(msg)
		if err != nil {
			return err
		}
		qs = append(qs, q)
	}
	a.and = qs
	return nil
}

func unmarshalQ(b []byte) (query, error) {
	var tnp testNamePattern
	err := json.Unmarshal(b, &tnp)
	if err == nil {
		return tnp, nil
	}
	var tsc testStatusConstraint
	err = json.Unmarshal(b, &tsc)
	if err == nil {
		return tsc, nil
	}
	var n not
	err = json.Unmarshal(b, &n)
	if err == nil {
		return n, nil
	}
	var o or
	err = json.Unmarshal(b, &o)
	if err == nil {
		return o, nil
	}
	var a and
	err = json.Unmarshal(b, &a)
	if err == nil {
		return a, nil
	}

	return nil, errors.New(`Failed to parse query fragment as test name pattern, test status constraint, negation, disjunction, or conjunction`)
}

// summary is the golang type for the JSON format in pass/total summary files.
type summary map[string][]int

type sharedInterface interface {
	ParseQueryParamInt(r *http.Request, key string) (*int, error)
	ParseQueryFilterParams(*http.Request) (shared.QueryFilter, error)
	LoadTestRuns([]shared.ProductSpec, mapset.Set, string, *time.Time, *time.Time, *int) ([]shared.TestRun, error)
	LoadTestRunsByIDs(ids shared.TestRunIDs) (result []shared.TestRun, err error)
	LoadTestRun(int64) (*shared.TestRun, error)
}

type defaultShared struct {
	ctx context.Context
}

func (defaultShared) ParseQueryParamInt(r *http.Request, key string) (*int, error) {
	return shared.ParseQueryParamInt(r, key)
}

func (defaultShared) ParseQueryFilterParams(r *http.Request) (shared.QueryFilter, error) {
	return shared.ParseQueryFilterParams(r)
}

func (sharedImpl defaultShared) LoadTestRuns(ps []shared.ProductSpec, ls mapset.Set, sha string, from *time.Time, to *time.Time, limit *int) ([]shared.TestRun, error) {
	return shared.LoadTestRuns(sharedImpl.ctx, ps, ls, sha, from, to, limit)
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
		var sha string
		var err error
		limit := 1
		products := runFilters.GetProductsOrDefault()
		testRuns, err = qh.sharedImpl.LoadTestRuns(products, runFilters.Labels, sha, runFilters.From, runFilters.To, &limit)
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

func isRequestCacheable(r *http.Request) bool {
	ids, err := shared.ParseRunIDsParam(r)
	return err == nil && len(ids) > 0
}
