// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/query/test"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/memcache"
)

func TestGetMemcacheKey(t *testing.T) {
	assert.Equal(t, "RESULTS_SUMMARY-1", getMemcacheKey(shared.TestRun{
		ID: 1,
	}))
}

func TestLoadSummary_cacheMiss(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	url := "https://example.com/1-summary.json.gz"
	testRun := shared.TestRun{
		ID:         1,
		ResultsURL: url,
	}
	key := getMemcacheKey(testRun)

	cache := NewMockreadWritable(mockCtrl)
	store := NewMockreadable(mockCtrl)
	sh := searchHandler{queryHandler{
		dataSource: cachedStore{
			cache: cache,
			store: store,
		},
	}}
	smry := []byte("{}")

	// Use channel to synchronize with expected async cache.Put().
	c := make(chan bool)
	w := test.NewMockWriteCloser(t, c)
	r := test.NewMockReadCloser(t, smry)
	cache.EXPECT().NewReadCloser(key).Return(nil, memcache.ErrCacheMiss)
	store.EXPECT().NewReadCloser(url).Return(r, nil)
	cache.EXPECT().NewWriteCloser(key).Return(w, nil)

	s, err := sh.loadSummary(shared.TestRun{
		ID:         1,
		ResultsURL: url,
	})
	assert.Nil(t, err)
	assert.Equal(t, smry, s)

	b := <-c
	assert.True(t, r.IsClosed())
	assert.Equal(t, true, b)
}

func TestLoadSummary_cacheHit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	url := "https://example.com/1-summary.json.gz"
	testRun := shared.TestRun{
		ID:         1,
		ResultsURL: url,
	}
	key := getMemcacheKey(testRun)

	cache := NewMockreadWritable(mockCtrl)
	sh := searchHandler{queryHandler{
		dataSource: cachedStore{cache: cache},
	}}
	smry := []byte("{}")
	r := test.NewMockReadCloser(t, smry)

	cache.EXPECT().NewReadCloser(key).Return(r, nil)

	s, err := sh.loadSummary(shared.TestRun{
		ID:         1,
		ResultsURL: url,
	})
	assert.Nil(t, err)
	assert.Equal(t, smry, s)
	assert.True(t, r.IsClosed())
}

func TestLoadSummary_missing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	url := "https://example.com/1-summary.json.gz"
	testRun := shared.TestRun{
		ID:         1,
		ResultsURL: url,
	}
	key := getMemcacheKey(testRun)

	cache := NewMockreadWritable(mockCtrl)
	store := NewMockreadable(mockCtrl)
	sh := searchHandler{queryHandler{
		dataSource: cachedStore{
			cache: cache,
			store: store,
		},
	}}
	storeMiss := errors.New("No such summary file")

	cache.EXPECT().NewReadCloser(key).Return(nil, memcache.ErrCacheMiss)
	store.EXPECT().NewReadCloser(url).Return(nil, storeMiss)

	s, err := sh.loadSummary(shared.TestRun{
		ID:         1,
		ResultsURL: url,
	})
	assert.Equal(t, storeMiss, err)
	assert.Nil(t, s)
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

	cache := NewMockreadWritable(mockCtrl)
	sh := searchHandler{queryHandler{
		dataSource: cachedStore{cache: cache},
	}}
	summaryBytes := [][]byte{
		[]byte(`{"/a/b/c":[1,2]}`),
		[]byte(`{"/x/y/z":[3,4]}`),
	}
	summaries := []summary{
		map[string][]int{"/a/b/c": []int{1, 2}},
		map[string][]int{"/x/y/z": []int{3, 4}},
	}
	rs := []*test.MockReadCloser{
		test.NewMockReadCloser(t, summaryBytes[0]),
		test.NewMockReadCloser(t, summaryBytes[1]),
	}

	cache.EXPECT().NewReadCloser(keys[0]).Return(rs[0], nil)
	cache.EXPECT().NewReadCloser(keys[1]).Return(rs[1], nil)

	ss, err := sh.loadSummaries(testRuns)
	assert.Nil(t, err)
	assert.Equal(t, summaries[0], ss[0])
	assert.Equal(t, summaries[1], ss[1])
	assert.True(t, rs[0].IsClosed())
	assert.True(t, rs[1].IsClosed())
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

	cache := NewMockreadWritable(mockCtrl)
	store := NewMockreadable(mockCtrl)
	sh := searchHandler{queryHandler{
		dataSource: cachedStore{
			cache: cache,
			store: store,
		},
	}}
	summaryBytes := [][]byte{
		[]byte(`{"/a/b/c":[1,2]}`),
	}
	storeMiss := errors.New("No such summary file")
	r := test.NewMockReadCloser(t, summaryBytes[0])

	cache.EXPECT().NewReadCloser(keys[0]).Return(r, nil)
	cache.EXPECT().NewReadCloser(keys[1]).Return(nil, memcache.ErrCacheMiss)
	store.EXPECT().NewReadCloser(urls[1]).Return(nil, storeMiss)

	_, err := sh.loadSummaries(testRuns)
	assert.Equal(t, storeMiss, err)
	assert.True(t, r.IsClosed())
}

func TestGetRunsAndFilters_default(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	si := NewMocksharedInterface(mockCtrl)
	sh := searchHandler{queryHandler{
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
	sh := searchHandler{queryHandler{
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

	si.EXPECT().LoadTestRun(testRuns[0].ID).Return(&testRuns[0], nil)
	si.EXPECT().LoadTestRun(testRuns[1].ID).Return(&testRuns[1], nil)

	trs, fs, err := sh.getRunsAndFilters(filters)
	assert.Nil(t, err)
	assert.Equal(t, testRuns, trs)
	assert.Equal(t, filters, fs)
}
