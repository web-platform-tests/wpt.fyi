// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/memcache"
)

type MockWriteCloser struct {
	b      bytes.Buffer
	closed bool
	t      *testing.T
	c      chan bool
}

func (mcw *MockWriteCloser) Write(p []byte) (n int, err error) {
	assert.False(mcw.t, mcw.closed)
	return mcw.b.Write(p)
}

func (mcw *MockWriteCloser) Close() error {
	mcw.closed = true
	if mcw.c != nil {
		mcw.c <- true
	}
	return nil
}

func NewMockWriteCloser(t *testing.T, c chan bool) *MockWriteCloser {
	return &MockWriteCloser{
		b:      bytes.Buffer{},
		closed: false,
		t:      t,
		c:      c,
	}
}

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
	sh := searchHandler{
		dataSource: cachedStore{
			cache: cache,
			store: store,
		},
	}
	smry := []byte("{}")

	// Use channel to synchronize with expected async cache.Put().
	c := make(chan bool)
	w := NewMockWriteCloser(t, c)
	cache.EXPECT().NewReader(key).Return(nil, memcache.ErrCacheMiss)
	store.EXPECT().NewReader(url).Return(bytes.NewReader(smry), nil)
	cache.EXPECT().NewWriteCloser(key).Return(w, nil)

	s, err := sh.loadSummary(shared.TestRun{
		ID:         1,
		ResultsURL: url,
	})
	assert.Nil(t, err)
	assert.Equal(t, smry, s)

	b := <-c
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
	sh := searchHandler{
		dataSource: cachedStore{cache: cache},
	}
	smry := []byte("{}")

	cache.EXPECT().NewReader(key).Return(bytes.NewReader(smry), nil)

	s, err := sh.loadSummary(shared.TestRun{
		ID:         1,
		ResultsURL: url,
	})
	assert.Nil(t, err)
	assert.Equal(t, smry, s)
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
	sh := searchHandler{
		dataSource: cachedStore{
			cache: cache,
			store: store,
		},
	}
	storeMiss := errors.New("No such summary file")

	cache.EXPECT().NewReader(key).Return(nil, memcache.ErrCacheMiss)
	store.EXPECT().NewReader(url).Return(nil, storeMiss)

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
	sh := searchHandler{
		dataSource: cachedStore{cache: cache},
	}
	summaryBytes := [][]byte{
		[]byte(`{"/a/b/c":[1,2]}`),
		[]byte(`{"/x/y/z":[3,4]}`),
	}
	summaries := []summary{
		map[string][]int{"/a/b/c": []int{1, 2}},
		map[string][]int{"/x/y/z": []int{3, 4}},
	}

	cache.EXPECT().NewReader(keys[0]).Return(bytes.NewReader(summaryBytes[0]), nil)
	cache.EXPECT().NewReader(keys[1]).Return(bytes.NewReader(summaryBytes[1]), nil)

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

	cache := NewMockreadWritable(mockCtrl)
	store := NewMockreadable(mockCtrl)
	sh := searchHandler{
		dataSource: cachedStore{
			cache: cache,
			store: store,
		},
	}
	summaryBytes := [][]byte{
		[]byte(`{"/a/b/c":[1,2]}`),
	}
	storeMiss := errors.New("No such summary file")

	cache.EXPECT().NewReader(keys[0]).Return(bytes.NewReader(summaryBytes[0]), nil)
	cache.EXPECT().NewReader(keys[1]).Return(nil, memcache.ErrCacheMiss)
	store.EXPECT().NewReader(urls[1]).Return(nil, storeMiss)

	_, err := sh.loadSummaries(testRuns)
	assert.Equal(t, storeMiss, err)
}

func TestGetRunsAndFilters_default(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	si := NewMocksharedInterface(mockCtrl)
	sh := searchHandler{
		sharedImpl: si,
	}

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
	filters := shared.SearchFilter{}

	si.EXPECT().LoadTestRuns(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(testRuns, nil)

	trs, fs, err := sh.getRunsAndFilters(filters)
	assert.Nil(t, err)
	assert.Equal(t, testRuns, trs)
	assert.Equal(t, shared.SearchFilter{
		RunIDs: runIDs,
	}, fs)
}

func TestGetRunsAndFilters_specificRunIDs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	si := NewMocksharedInterface(mockCtrl)
	sh := searchHandler{
		sharedImpl: si,
	}

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
	filters := shared.SearchFilter{
		RunIDs: runIDs,
	}

	si.EXPECT().LoadTestRun(testRuns[0].ID).Return(&testRuns[0], nil)
	si.EXPECT().LoadTestRun(testRuns[1].ID).Return(&testRuns[1], nil)

	trs, fs, err := sh.getRunsAndFilters(filters)
	assert.Nil(t, err)
	assert.Equal(t, testRuns, trs)
	assert.Equal(t, filters, fs)
}

func TestPrepareResponse(t *testing.T) {
	runIDs := []int64{1, 2}
	testRuns := []shared.TestRun{
		shared.TestRun{
			ID:         runIDs[0],
			ResultsURL: "https://example.com/1-summary.json.gz",
		},
		shared.TestRun{
			ID:         runIDs[1],
			ResultsURL: "https://example.com/2-summary.json.gz",
		},
	}
	filters := shared.SearchFilter{
		RunIDs: runIDs,
		Q:      "/b/",
	}
	summaries := []summary{
		map[string][]int{
			"/a/b/c": []int{1, 2},
			"/b/c":   []int{9, 9},
		},
		map[string][]int{
			"/z/b/c": []int{0, 8},
			"/x/y/z": []int{3, 4},
			"/b/c":   []int{5, 9},
		},
	}

	resp := prepareResponse(filters, testRuns, summaries)
	assert.Equal(t, testRuns, resp.Runs)
	assert.Equal(t, []SearchResult{
		SearchResult{
			Name: "/a/b/c",
			Status: []SearchRunResult{
				SearchRunResult{
					Passes: 1,
					Total:  2,
				},
				SearchRunResult{},
			},
		},
		SearchResult{
			Name: "/b/c",
			Status: []SearchRunResult{
				SearchRunResult{
					Passes: 9,
					Total:  9,
				},
				SearchRunResult{
					Passes: 5,
					Total:  9,
				},
			},
		},
		SearchResult{
			Name: "/z/b/c",
			Status: []SearchRunResult{
				SearchRunResult{},
				SearchRunResult{
					Passes: 0,
					Total:  8,
				},
			},
		},
	}, resp.Results)
}
