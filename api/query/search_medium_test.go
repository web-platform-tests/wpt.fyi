// +build medium

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	time "time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine/datastore"
)

type shouldCache struct {
	Called bool

	t        *testing.T
	expected bool
	delegate func(context.Context, int, []byte) bool
}

func (sc *shouldCache) ShouldCache(ctx context.Context, statusCode int, payload []byte) bool {
	sc.Called = true
	ret := sc.delegate(ctx, statusCode, payload)
	assert.Equal(sc.t, sc.expected, ret)
	return ret
}

func NewShouldCache(t *testing.T, expected bool, delegate func(context.Context, int, []byte) bool) *shouldCache {
	return &shouldCache{false, t, expected, delegate}
}

func TestUnstructuredSearchHandler(t *testing.T) {
	urls := []string{
		"https://example.com/1-summary.json.gz",
		"https://example.com/2-summary.json.gz",
	}
	testRuns := shared.TestRuns{
		shared.TestRun{
			ResultsURL: urls[0],
		},
		shared.TestRun{
			ResultsURL: urls[1],
		},
	}
	summaryBytes := [][]byte{
		[]byte(`{"/a/b/c":[1,2],"/b/c":[9,9]}`),
		[]byte(`{"/z/b/c":[0,8],"/x/y/z":[3,4],"/b/c":[5,9]}`),
	}

	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()

	// Scope setup context.
	{
		req, err := i.NewRequest("GET", "/", nil)
		assert.Nil(t, err)
		ctx := shared.NewAppEngineContext(req)

		for idx := range testRuns {
			key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &testRuns[idx])
			assert.Nil(t, err)
			id := key.IntID()
			assert.NotEqual(t, 0, id)
			testRuns[idx].ID = id
		}
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// TODO(markdittmer): Should this be hitting GCS instead?
	cache := sharedtest.NewMockReadable(mockCtrl)
	rs := []*sharedtest.MockReadCloser{
		sharedtest.NewMockReadCloser(t, summaryBytes[0]),
		sharedtest.NewMockReadCloser(t, summaryBytes[1]),
	}

	cache.EXPECT().NewReadCloser(urls[0]).Return(rs[0], nil)
	cache.EXPECT().NewReadCloser(urls[1]).Return(rs[1], nil)

	// Same params as TestGetRunsAndFilters_specificRunIDs.
	q := "/b/"
	url := fmt.Sprintf(
		"/api/search?run_ids=%s&q=%s",
		url.QueryEscape(fmt.Sprintf("%d,%d", testRuns[0].ID, testRuns[1].ID)),
		url.QueryEscape(q))
	r, err := i.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	ctx := shared.NewAppEngineContext(r)
	w := httptest.NewRecorder()

	// TODO: This is parroting apiSearchHandler details. Perhaps they should be
	// abstracted and tested directly.
	mc := shared.NewGZReadWritable(shared.NewMemcacheReadWritable(ctx, 48*time.Hour))
	sh := unstructuredSearchHandler{queryHandler{
		store:      shared.NewAppEngineDatastore(ctx, false),
		sharedImpl: defaultShared{ctx},
		dataSource: shared.NewByteCachedStore(ctx, mc, cache),
	}}
	sc := NewShouldCache(t, true, shouldCacheSearchResponse)

	ch := shared.NewCachingHandler(ctx, sh, mc, isRequestCacheable, shared.URLAsCacheKey, sc.ShouldCache)
	ch.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, err := ioutil.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	var data shared.SearchResponse
	err = json.Unmarshal(bytes, &data)
	assert.Nil(t, err)

	// Same result as TestGetRunsAndFilters_specificRunIDs.
	assert.Equal(t, shared.SearchResponse{
		Runs: testRuns,
		Results: []shared.SearchResult{
			shared.SearchResult{
				Test: "/a/b/c",
				LegacyStatus: []shared.LegacySearchRunResult{
					shared.LegacySearchRunResult{
						Passes: 1,
						Total:  2,
					},
					shared.LegacySearchRunResult{},
				},
			},
			shared.SearchResult{
				Test: "/b/c",
				LegacyStatus: []shared.LegacySearchRunResult{
					shared.LegacySearchRunResult{
						Passes: 9,
						Total:  9,
					},
					shared.LegacySearchRunResult{
						Passes: 5,
						Total:  9,
					},
				},
			},
			shared.SearchResult{
				Test: "/z/b/c",
				LegacyStatus: []shared.LegacySearchRunResult{
					shared.LegacySearchRunResult{},
					shared.LegacySearchRunResult{
						Passes: 0,
						Total:  8,
					},
				},
			},
		},
	}, data)

	assert.True(t, rs[0].IsClosed())
	assert.True(t, rs[1].IsClosed())
	assert.True(t, sc.Called)
}

func TestStructuredSearchHandler_equivalentToUnstructured(t *testing.T) {
	urls := []string{
		"https://example.com/1-summary.json.gz",
		"https://example.com/2-summary.json.gz",
	}
	testRuns := []shared.TestRun{
		shared.TestRun{
			ResultsURL: urls[0],
		},
		shared.TestRun{
			ResultsURL: urls[1],
		},
	}
	summaryBytes := [][]byte{
		[]byte(`{"/a/b/c":[1,2],"/b/c":[9,9]}`),
		[]byte(`{"/z/b/c":[0,8],"/x/y/z":[3,4],"/b/c":[5,9]}`),
	}

	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()

	// Scope setup context.
	{
		req, err := i.NewRequest("GET", "/", nil)
		assert.Nil(t, err)
		ctx := shared.NewAppEngineContext(req)

		for idx, testRun := range testRuns {
			key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &testRun)
			assert.Nil(t, err)
			id := key.IntID()
			assert.NotEqual(t, 0, id)
			testRun.ID = id
			// Copy back testRun after mutating ID.
			testRuns[idx] = testRun
		}
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// TODO(markdittmer): Should this be hitting GCS instead?
	store := sharedtest.NewMockReadable(mockCtrl)
	rs := []*sharedtest.MockReadCloser{
		sharedtest.NewMockReadCloser(t, summaryBytes[0]),
		sharedtest.NewMockReadCloser(t, summaryBytes[1]),
	}

	store.EXPECT().NewReadCloser(urls[0]).Return(rs[0], nil)
	store.EXPECT().NewReadCloser(urls[1]).Return(rs[1], nil)

	// Same params as TestGetRunsAndFilters_specificRunIDs.
	q, err := json.Marshal("/b/")
	assert.Nil(t, err)
	url := "/api/search"
	r, err := i.NewRequest("POST", url, bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"run_ids": [%d, %d],
		"query": {"exists": [{"pattern": %s}] }
	}`, testRuns[0].ID, testRuns[1].ID, string(q)))))
	assert.Nil(t, err)
	ctx := shared.NewAppEngineContext(r)
	w := httptest.NewRecorder()

	// TODO: This is parroting apiSearchHandler details. Perhaps they should be
	// abstracted and tested directly.
	api := shared.NewAppEngineAPI(ctx)
	mc := shared.NewGZReadWritable(shared.NewMemcacheReadWritable(ctx, 48*time.Hour))
	sh := structuredSearchHandler{
		queryHandler{
			store:      shared.NewAppEngineDatastore(ctx, false),
			sharedImpl: defaultShared{ctx},
			dataSource: shared.NewByteCachedStore(ctx, mc, store),
		},
		api,
	}
	sc := NewShouldCache(t, true, shouldCacheSearchResponse)

	ch := shared.NewCachingHandler(ctx, sh, mc, isRequestCacheable, shared.URLAsCacheKey, sc.ShouldCache)
	ch.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, err := ioutil.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	var data shared.SearchResponse
	err = json.Unmarshal(bytes, &data)
	assert.Nil(t, err, "Error unmarshalling \"%s\"", string(bytes))

	// Same result as TestGetRunsAndFilters_specificRunIDs.
	assert.Equal(t, shared.SearchResponse{
		Runs: testRuns,
		Results: []shared.SearchResult{
			shared.SearchResult{
				Test: "/a/b/c",
				LegacyStatus: []shared.LegacySearchRunResult{
					shared.LegacySearchRunResult{
						Passes: 1,
						Total:  2,
					},
					shared.LegacySearchRunResult{},
				},
			},
			shared.SearchResult{
				Test: "/b/c",
				LegacyStatus: []shared.LegacySearchRunResult{
					shared.LegacySearchRunResult{
						Passes: 9,
						Total:  9,
					},
					shared.LegacySearchRunResult{
						Passes: 5,
						Total:  9,
					},
				},
			},
			shared.SearchResult{
				Test: "/z/b/c",
				LegacyStatus: []shared.LegacySearchRunResult{
					shared.LegacySearchRunResult{},
					shared.LegacySearchRunResult{
						Passes: 0,
						Total:  8,
					},
				},
			},
		},
	}, data)

	assert.True(t, rs[0].IsClosed())
	assert.True(t, rs[1].IsClosed())
	assert.True(t, sc.Called)
}

func TestUnstructuredSearchHandler_doNotCacheEmptyResult(t *testing.T) {
	urls := []string{
		"https://example.com/1-summary.json.gz",
		"https://example.com/2-summary.json.gz",
	}
	testRuns := shared.TestRuns{
		shared.TestRun{
			ResultsURL: urls[0],
		},
		shared.TestRun{
			ResultsURL: urls[1],
		},
	}
	summaryBytes := [][]byte{
		[]byte(`{}`),
		[]byte(`{}`),
	}

	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()

	// Scope setup context.
	{
		req, err := i.NewRequest("GET", "/", nil)
		assert.Nil(t, err)
		ctx := shared.NewAppEngineContext(req)

		for idx, testRun := range testRuns {
			key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &testRun)
			assert.Nil(t, err)
			id := key.IntID()
			assert.NotEqual(t, 0, id)
			testRun.ID = id
			// Copy back testRun after mutating ID.
			testRuns[idx] = testRun
		}
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// TODO(markdittmer): Should this be hitting GCS instead?
	store := sharedtest.NewMockReadable(mockCtrl)
	rs := []*sharedtest.MockReadCloser{
		sharedtest.NewMockReadCloser(t, summaryBytes[0]),
		sharedtest.NewMockReadCloser(t, summaryBytes[1]),
	}

	store.EXPECT().NewReadCloser(urls[0]).Return(rs[0], nil)
	store.EXPECT().NewReadCloser(urls[1]).Return(rs[1], nil)

	// Same params as TestGetRunsAndFilters_specificRunIDs.
	q := "/b/"
	url := fmt.Sprintf(
		"/api/search?run_ids=%s&q=%s",
		url.QueryEscape(fmt.Sprintf("%d,%d", testRuns[0].ID, testRuns[1].ID)),
		url.QueryEscape(q))
	r, err := i.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	ctx := shared.NewAppEngineContext(r)
	w := httptest.NewRecorder()

	// TODO: This is parroting apiSearchHandler details. Perhaps they should be
	// abstracted and tested directly.
	mc := shared.NewGZReadWritable(shared.NewMemcacheReadWritable(ctx, 48*time.Hour))
	sh := unstructuredSearchHandler{
		queryHandler{
			store:      shared.NewAppEngineDatastore(ctx, false),
			sharedImpl: defaultShared{ctx},
			dataSource: shared.NewByteCachedStore(ctx, mc, store),
		},
	}
	sc := NewShouldCache(t, false, shouldCacheSearchResponse)

	ch := shared.NewCachingHandler(ctx, sh, mc, isRequestCacheable, shared.URLAsCacheKey, sc.ShouldCache)
	ch.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, err := ioutil.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	var data shared.SearchResponse
	err = json.Unmarshal(bytes, &data)
	assert.Nil(t, err)

	// Same result as TestGetRunsAndFilters_specificRunIDs.
	assert.Equal(t, shared.SearchResponse{
		Runs:    testRuns,
		Results: []shared.SearchResult{},
	}, data)

	assert.True(t, rs[0].IsClosed())
	assert.True(t, rs[1].IsClosed())
	assert.True(t, sc.Called)
}

func TestStructuredSearchHandler_doNotCacheEmptyResult(t *testing.T) {
	urls := []string{
		"https://example.com/1-summary.json.gz",
		"https://example.com/2-summary.json.gz",
	}
	testRuns := []shared.TestRun{
		shared.TestRun{
			ResultsURL: urls[0],
		},
		shared.TestRun{
			ResultsURL: urls[1],
		},
	}
	summaryBytes := [][]byte{
		[]byte(`{}`),
		[]byte(`{}`),
	}

	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()

	// Scope setup context.
	{
		req, err := i.NewRequest("GET", "/", nil)
		assert.Nil(t, err)
		ctx := shared.NewAppEngineContext(req)

		for idx, testRun := range testRuns {
			key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &testRun)
			assert.Nil(t, err)
			id := key.IntID()
			assert.NotEqual(t, 0, id)
			testRun.ID = id
			// Copy back testRun after mutating ID.
			testRuns[idx] = testRun
		}
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// TODO(markdittmer): Should this be hitting GCS instead?
	store := sharedtest.NewMockReadable(mockCtrl)
	rs := []*sharedtest.MockReadCloser{
		sharedtest.NewMockReadCloser(t, summaryBytes[0]),
		sharedtest.NewMockReadCloser(t, summaryBytes[1]),
	}

	store.EXPECT().NewReadCloser(urls[0]).Return(rs[0], nil)
	store.EXPECT().NewReadCloser(urls[1]).Return(rs[1], nil)

	// Same params as TestGetRunsAndFilters_specificRunIDs.
	q, err := json.Marshal("/b/")
	assert.Nil(t, err)
	url := "/api/search"
	r, err := i.NewRequest("POST", url, bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"run_ids": [%d, %d],
		"query": {"exists": [{"pattern": %s}] }
	}`, testRuns[0].ID, testRuns[1].ID, string(q)))))
	assert.Nil(t, err)
	ctx := shared.NewAppEngineContext(r)
	w := httptest.NewRecorder()

	// TODO: This is parroting apiSearchHandler details. Perhaps they should be
	// abstracted and tested directly.
	api := shared.NewAppEngineAPI(ctx)
	mc := shared.NewGZReadWritable(shared.NewMemcacheReadWritable(ctx, 48*time.Hour))
	sh := structuredSearchHandler{
		queryHandler{
			store:      shared.NewAppEngineDatastore(ctx, false),
			sharedImpl: defaultShared{ctx},
			dataSource: shared.NewByteCachedStore(ctx, mc, store),
		},
		api,
	}
	sc := NewShouldCache(t, false, shouldCacheSearchResponse)

	ch := shared.NewCachingHandler(ctx, sh, mc, isRequestCacheable, shared.URLAsCacheKey, sc.ShouldCache)
	ch.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, err := ioutil.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	var data shared.SearchResponse
	err = json.Unmarshal(bytes, &data)
	assert.Nil(t, err)

	// Same result as TestGetRunsAndFilters_specificRunIDs.
	assert.Equal(t, shared.SearchResponse{
		Runs:    testRuns,
		Results: []shared.SearchResult{},
	}, data)

	assert.True(t, rs[0].IsClosed())
	assert.True(t, rs[1].IsClosed())
	assert.True(t, sc.Called)
}
