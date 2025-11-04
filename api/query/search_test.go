//go:build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func doTestIC(t *testing.T, p, q string) {
	runIDs := []int64{1, 2}
	testRuns := []shared.TestRun{
		{
			ID:         runIDs[0],
			ResultsURL: "https://example.com/1-summary_v2.json.gz",
		},
		{
			ID:         runIDs[1],
			ResultsURL: "https://example.com/2-summary_v2.json.gz",
		},
	}
	filters := shared.QueryFilter{
		RunIDs: runIDs,
		Q:      q,
	}
	summaries := []summary{
		{
			"/a" + p + "c": {
				Status: "T", Counts: []int{0, 0},
			},
			p + "c": {
				Status: "O", Counts: []int{8, 8},
			},
		},
		{
			"/z" + p + "c": {
				Status: "F", Counts: []int{0, 7},
			},
			"/x/y/z": {
				Status: "O", Counts: []int{2, 3},
			},
			p + "c": {
				Status: "O", Counts: []int{4, 8},
			},
		},
	}

	resp := prepareSearchResponse(&filters, testRuns, summaries)
	assert.Equal(t, testRuns, resp.Runs)
	expectedResults := []shared.SearchResult{
		{
			Test: "/a" + p + "c",
			LegacyStatus: []shared.LegacySearchRunResult{
				{
					Passes:        0,
					Total:         0,
					Status:        "T",
					NewAggProcess: true,
				},
				{
					Passes:        0,
					Total:         0,
					Status:        "",
					NewAggProcess: false,
				},
			},
		},
		{
			Test: p + "c",
			LegacyStatus: []shared.LegacySearchRunResult{
				{
					Passes:        8,
					Total:         8,
					Status:        "O",
					NewAggProcess: true,
				},
				{
					Passes:        4,
					Total:         8,
					Status:        "O",
					NewAggProcess: true,
				},
			},
		},
		{
			Test: "/z" + p + "c",
			LegacyStatus: []shared.LegacySearchRunResult{
				{},
				{
					Passes:        0,
					Total:         7,
					Status:        "F",
					NewAggProcess: true,
				},
			},
		},
	}
	sort.Sort(byName(expectedResults))
	assert.Equal(t, expectedResults, resp.Results)
}

func testIC(t *testing.T, str string, upperQ bool) {
	var p, q string
	if upperQ {
		p = strings.ToLower(str)
		q = strings.ToUpper(str)
	} else {
		p = strings.ToUpper(str)
		q = strings.ToLower(str)
	}

	doTestIC(t, p, q)
}

func TestPrepareSearchResponse_qUC(t *testing.T) {
	testIC(t, "/b/", true)
}

func TestPrepareSearchResponse_pUC(t *testing.T) {
	testIC(t, "/b/", false)
}

func TestStructuredSearchHandler_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := sharedtest.NewMockDatastore(ctrl)

	respBytes := []byte(`{}`)

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/search/cache", r.URL.Path)
		w.Write(respBytes)
	}))

	serverURL, err := url.Parse(server.URL)
	assert.Nil(t, err)
	hostname := serverURL.Host

	runIDs := []int64{1, 2}
	urls := []string{
		"https://example.com/1-summary_v2.json.gz",
		"https://example.com/2-summary_v2.json.gz",
	}
	chrome, _ := shared.ParseProductSpec("chrome")
	edge, _ := shared.ParseProductSpec("edge")
	testRuns := shared.TestRunsByProduct{
		shared.ProductTestRuns{
			Product: chrome,
			TestRuns: shared.TestRuns{
				shared.TestRun{
					ID:         runIDs[0],
					ResultsURL: urls[0],
				},
			},
		},
		shared.ProductTestRuns{
			Product: edge,
			TestRuns: shared.TestRuns{
				shared.TestRun{
					ID:         runIDs[1],
					ResultsURL: urls[1],
				},
			},
		},
	}
	for _, id := range runIDs {
		mockStore.EXPECT().NewIDKey("TestRun", id).Return(sharedtest.MockKey{ID: id})
	}
	mockStore.EXPECT().GetMulti(sharedtest.SameKeys(runIDs), gomock.Any()).DoAndReturn(sharedtest.MultiRuns(testRuns.AllRuns()))

	api := sharedtest.NewMockAppEngineAPI(ctrl)
	r := httptest.NewRequest("POST", "https://example.com/api/query", bytes.NewBuffer([]byte(`{"run_ids":[1,2],"query":{"browser_name":"chrome","status":"PASS"}}`)))

	api.EXPECT().Context().Return(sharedtest.NewTestContext())
	api.EXPECT().GetServiceHostname("searchcache").Return(hostname)
	api.EXPECT().GetHTTPClientWithTimeout(gomock.Any()).Return(server.Client())
	w := httptest.NewRecorder()
	structuredSearchHandler{queryHandler{store: mockStore}, api}.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, respBytes, w.Body.Bytes())
}

func TestStructuredSearchHandler_failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := sharedtest.NewMockDatastore(ctrl)

	respBytes := []byte(`Unknown run ID: 42`)

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/search/cache", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respBytes)
	}))

	serverURL, err := url.Parse(server.URL)
	assert.Nil(t, err)
	hostname := serverURL.Host
	runIDs := []int64{1, 2}
	urls := []string{
		"https://example.com/1-summary_v2.json.gz",
		"https://example.com/2-summary_v2.json.gz",
	}
	chrome, _ := shared.ParseProductSpec("chrome")
	edge, _ := shared.ParseProductSpec("edge")
	testRuns := shared.TestRunsByProduct{
		shared.ProductTestRuns{
			Product: chrome,
			TestRuns: shared.TestRuns{
				shared.TestRun{
					ID:         runIDs[0],
					ResultsURL: urls[0],
				},
			},
		},
		shared.ProductTestRuns{
			Product: edge,
			TestRuns: shared.TestRuns{
				shared.TestRun{
					ID:         runIDs[1],
					ResultsURL: urls[1],
				},
			},
		},
	}
	var id int64 = 42
	mockStore.EXPECT().NewIDKey("TestRun", id).Return(sharedtest.MockKey{ID: id})
	mockStore.EXPECT().GetMulti(sharedtest.SameKeys([]int64{id}), gomock.Any()).DoAndReturn(sharedtest.MultiRuns(testRuns.AllRuns()))

	api := sharedtest.NewMockAppEngineAPI(ctrl)
	r := httptest.NewRequest("POST", "https://example.com/api/query", bytes.NewBuffer([]byte(`{"run_ids":[42],"query":{"browser_name":"chrome","status":"PASS"}}`)))

	api.EXPECT().Context().Return(sharedtest.NewTestContext())
	api.EXPECT().GetServiceHostname("searchcache").Return(hostname)
	api.EXPECT().GetHTTPClientWithTimeout(gomock.Any()).Return(server.Client())

	w := httptest.NewRecorder()
	structuredSearchHandler{queryHandler{store: mockStore}, api}.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, respBytes, w.Body.Bytes())
}
