// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestGetMemcacheKey(t *testing.T) {
	assert.Equal(t, "RESULTS_SUMMARY-1", getMemcacheKey(shared.TestRun{
		ID: 1,
	}))
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

	cachedStore := sharedtest.NewMockCachedStore(mockCtrl)
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

	cachedStore := sharedtest.NewMockCachedStore(mockCtrl)
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
	assert.Contains(t, err.Error(), storeMiss.Error())
}

func TestGetRunsAndFilters_default(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockStore := sharedtest.NewMockDatastore(mockCtrl)
	sh := unstructuredSearchHandler{queryHandler{
		store: mockStore,
	}}

	runIDs := []int64{1, 2}
	urls := []string{
		"https://example.com/1-summary.json.gz",
		"https://example.com/2-summary.json.gz",
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
					TimeStart:  time.Now(),
				},
			},
		},
		shared.ProductTestRuns{
			Product: edge,
			TestRuns: shared.TestRuns{
				shared.TestRun{
					ID:         runIDs[1],
					ResultsURL: urls[1],
					TimeStart:  time.Now().AddDate(0, 0, -1),
				},
			},
		},
	}
	filters := shared.QueryFilter{}

	mockStore.EXPECT().LoadTestRuns(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(testRuns, nil)

	trs, fs, err := sh.getRunsAndFilters(filters)
	assert.Nil(t, err)
	assert.Equal(t, testRuns.AllRuns(), trs)
	assert.Equal(t, shared.QueryFilter{
		RunIDs: runIDs,
	}, fs)
}

func TestGetRunsAndFilters_specificRunIDs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockStore := sharedtest.NewMockDatastore(mockCtrl)
	sh := unstructuredSearchHandler{queryHandler{
		store: mockStore,
	}}

	runIDs := []int64{1, 2}
	urls := []string{
		"https://example.com/1-summary.json.gz",
		"https://example.com/2-summary.json.gz",
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
					TimeStart:  time.Now(),
				},
			},
		},
		shared.ProductTestRuns{
			Product: edge,
			TestRuns: shared.TestRuns{
				shared.TestRun{
					ID:         runIDs[1],
					ResultsURL: urls[1],
					TimeStart:  time.Now().AddDate(0, 0, -1),
				},
			},
		},
	}
	filters := shared.QueryFilter{
		RunIDs: runIDs,
	}

	for _, id := range runIDs {
		mockStore.EXPECT().NewKey("TestRun", id).Return(sharedtest.MockKey{id})
	}
	mockStore.EXPECT().GetMulti(sharedtest.SameKeys(runIDs), gomock.Any()).DoAndReturn(sharedtest.MultiRuns(testRuns.AllRuns()))

	trs, fs, err := sh.getRunsAndFilters(filters)
	assert.Nil(t, err)
	assert.Equal(t, testRuns.AllRuns(), trs)
	assert.Equal(t, filters, fs)
}

func TestIsRequestCacheable_getNotCacheable(t *testing.T) {
	assert.False(t, isRequestCacheable(httptest.NewRequest("GET", "https://wpt.fyi/api/search", nil)))
}

func TestIsRequestCacheable_getCacheable(t *testing.T) {
	assert.True(t, isRequestCacheable(httptest.NewRequest("GET", "https://wpt.fyi/api/search?run_ids=1,2,-3", nil)))
}

func TestIsRequestCacheable_postNotCacheable(t *testing.T) {
	assert.False(t, isRequestCacheable(httptest.NewRequest("POST", "https://wpt.fyi/api/search", bytes.NewBuffer([]byte("{}")))))
}

func TestIsRequestCacheable_postCacheable(t *testing.T) {
	assert.True(t, isRequestCacheable(httptest.NewRequest("POST", "https://wpt.fyi/api/search", bytes.NewBuffer([]byte(`{"run_ids":[1,2,-3]}`)))))
}
