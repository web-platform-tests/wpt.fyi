// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestGetMemcacheKey(t *testing.T) {
	assert.Equal(t, "RESULTS_SUMMARY-1", getMemcacheKey(shared.TestRun{
		ID: 1,
	}))
}

func TestStructuredQuery_empty(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_missingRunIDs(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"query": {
			"pattern": "/2dcontext/"
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_missingQuery(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2]
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_emptyRunIDs(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [],
		"query": {
			"pattern": "/2dcontext/"
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_emptyBrowserName(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"browser_name": "",
			"status": "PASS"
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_missingStatus(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"browser_name": "chrome"
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_badStatus(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"browser_name": "chrome",
			"status": "NOT_A_REAL_STATUS"
		}
	}`), &rq)
	assert.NotNil(t, err)
}
func TestStructuredQuery_unknownStatus(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"browser_name": "chrome",
			"status": "UNKNOWN"
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: testStatusConstraint{browserName: "chrome", status: shared.TestStatusValueFromString("UNKNOWN")}}, rq)
}

func TestStructuredQuery_pattern(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"pattern": "/2dcontext/"
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: testNamePattern{pattern: "/2dcontext/"}}, rq)
}

func TestStructuredQuery_status(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"browser_name": "FiReFoX",
			"status": "PaSs"
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: testStatusConstraint{browserName: "firefox", status: shared.TestStatusValueFromString("PASS")}}, rq)
}

func TestStructuredQuery_not(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"not": {
				"pattern": "cssom"
			}
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: not{testNamePattern{pattern: "cssom"}}}, rq)
}

func TestStructuredQuery_or(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"or": [
				{"pattern": "cssom"},
				{"pattern": "html"}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: or{or: []query{testNamePattern{pattern: "cssom"}, testNamePattern{pattern: "html"}}}}, rq)
}

func TestStructuredQuery_and(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"and": [
				{"pattern": "cssom"},
				{"pattern": "html"}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: and{and: []query{testNamePattern{pattern: "cssom"}, testNamePattern{pattern: "html"}}}}, rq)
}

func TestStructuredQuery_nested(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"or": [
				{
					"and": [
						{"not": {"pattern": "cssom"}},
						{"pattern": "html"}
					]
				},
				{
					"browser_name": "eDgE",
					"status": "tImEoUt"
				}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{
		runIDs: []int64{0, 1, 2},
		query: or{
			or: []query{
				and{
					and: []query{
						not{not: testNamePattern{pattern: "cssom"}},
						testNamePattern{pattern: "html"},
					},
				}, testStatusConstraint{
					browserName: "edge",
					status:      shared.TestStatusValueFromString("TIMEOUT"),
				},
			},
		},
	}, rq)
}

func TestLoadSummaries_success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	urls := []string{
		"https://example.com/1-summary.json.gz",
		"https://example.com/2-summary.json.gz",
	}
	testRuns := []shared.TestRun{
		shared.TestRun{
			ID:         1,
			ResultsURL: urls[0],
		},
		shared.TestRun{
			ID:         2,
			ResultsURL: urls[1],
		},
	}
	keys := []string{
		getMemcacheKey(testRuns[0]),
		getMemcacheKey(testRuns[1]),
	}

	cachedStore := shared.NewMockCachedStore(mockCtrl)
	sh := unstructuredSearchHandler{queryHandler{dataSource: cachedStore}}
	summaryBytes := [][]byte{
		[]byte(`{"/a/b/c":[1,2]}`),
		[]byte(`{"/x/y/z":[3,4]}`),
	}
	summaries := []summary{
		map[string][]int{"/a/b/c": []int{1, 2}},
		map[string][]int{"/x/y/z": []int{3, 4}},
	}

	bindCopySlice := func(i int) func(_, _, _ interface{}) {
		return func(cid, sid, iv interface{}) {
			ptr := iv.(*[]byte)
			*ptr = summaryBytes[i]
		}
	}
	for i, key := range keys {
		cachedStore.EXPECT().Get(key, urls[i], gomock.Any()).Do(bindCopySlice(i)).Return(nil)
	}

	ss, err := sh.loadSummaries(testRuns)
	assert.Nil(t, err)
	assert.Equal(t, summaries[0], ss[0])
	assert.Equal(t, summaries[1], ss[1])
}

func TestLoadSummaries_fail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	urls := []string{
		"https://example.com/1-summary.json.gz",
		"https://example.com/2-summary.json.gz",
	}
	testRuns := []shared.TestRun{
		shared.TestRun{
			ID:         1,
			ResultsURL: urls[0],
		},
		shared.TestRun{
			ID:         2,
			ResultsURL: urls[1],
		},
	}
	keys := []string{
		getMemcacheKey(testRuns[0]),
		getMemcacheKey(testRuns[1]),
	}

	cachedStore := shared.NewMockCachedStore(mockCtrl)
	sh := unstructuredSearchHandler{queryHandler{dataSource: cachedStore}}
	summaryBytes := [][]byte{
		[]byte(`{"/a/b/c":[1,2]}`),
	}

	storeMiss := errors.New("No such summary file")
	cachedStore.EXPECT().Get(keys[0], urls[0], gomock.Any()).Do(func(cid, sid, iv interface{}) {
		ptr := iv.(*[]byte)
		*ptr = summaryBytes[0]
	}).Return(nil)
	cachedStore.EXPECT().Get(keys[1], urls[1], gomock.Any()).Return(storeMiss)

	_, err := sh.loadSummaries(testRuns)
	assert.Equal(t, storeMiss, err)
}

func TestGetRunsAndFilters_default(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	si := NewMocksharedInterface(mockCtrl)
	sh := unstructuredSearchHandler{queryHandler{
		sharedImpl: si,
	}}

	runIDs := []int64{1, 2}
	urls := []string{
		"https://example.com/1-summary.json.gz",
		"https://example.com/2-summary.json.gz",
	}
	testRuns := []shared.TestRun{
		shared.TestRun{
			ID:         runIDs[0],
			ResultsURL: urls[0],
		},
		shared.TestRun{
			ID:         runIDs[1],
			ResultsURL: urls[1],
		},
	}
	filters := shared.QueryFilter{}

	si.EXPECT().LoadTestRuns(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(testRuns, nil)

	trs, fs, err := sh.getRunsAndFilters(filters)
	assert.Nil(t, err)
	assert.Equal(t, testRuns, trs)
	assert.Equal(t, shared.QueryFilter{
		RunIDs: runIDs,
	}, fs)
}

func TestGetRunsAndFilters_specificRunIDs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	si := NewMocksharedInterface(mockCtrl)
	sh := unstructuredSearchHandler{queryHandler{
		sharedImpl: si,
	}}

	runIDs := []int64{1, 2}
	urls := []string{
		"https://example.com/1-summary.json.gz",
		"https://example.com/2-summary.json.gz",
	}
	testRuns := []shared.TestRun{
		shared.TestRun{
			ID:         runIDs[0],
			ResultsURL: urls[0],
		},
		shared.TestRun{
			ID:         runIDs[1],
			ResultsURL: urls[1],
		},
	}
	filters := shared.QueryFilter{
		RunIDs: runIDs,
	}

	si.EXPECT().LoadTestRunsByIDs(shared.TestRunIDs([]int64{testRuns[0].ID, testRuns[1].ID})).Return([]shared.TestRun{testRuns[0], testRuns[1]}, nil)

	trs, fs, err := sh.getRunsAndFilters(filters)
	assert.Nil(t, err)
	assert.Equal(t, testRuns, trs)
	assert.Equal(t, filters, fs)
}
