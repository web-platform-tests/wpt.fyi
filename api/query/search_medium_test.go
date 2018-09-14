// +build medium

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/iotest"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine/datastore"
)

func TestSearchHandler(t *testing.T) {
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
	store := shared.NewMockReadable(mockCtrl)
	rs := []*iotest.MockReadCloser{
		iotest.NewMockReadCloser(t, summaryBytes[0]),
		iotest.NewMockReadCloser(t, summaryBytes[1]),
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

	sh := searchHandler{queryHandler{
		sharedImpl: defaultShared{ctx},
		dataSource: shared.NewByteCachedStore(ctx, shared.NewMemcacheReadWritable(ctx), store),
	}}

	sh.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, err := ioutil.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	var data SearchResponse
	err = json.Unmarshal(bytes, &data)
	assert.Nil(t, err)

	// Same result as TestGetRunsAndFilters_specificRunIDs.
	assert.Equal(t, SearchResponse{
		Runs: testRuns,
		Results: []SearchResult{
			SearchResult{
				Test: "/a/b/c",
				LegacyStatus: []LegacySearchRunResult{
					LegacySearchRunResult{
						Passes: 1,
						Total:  2,
					},
					LegacySearchRunResult{},
				},
			},
			SearchResult{
				Test: "/b/c",
				LegacyStatus: []LegacySearchRunResult{
					LegacySearchRunResult{
						Passes: 9,
						Total:  9,
					},
					LegacySearchRunResult{
						Passes: 5,
						Total:  9,
					},
				},
			},
			SearchResult{
				Test: "/z/b/c",
				LegacyStatus: []LegacySearchRunResult{
					LegacySearchRunResult{},
					LegacySearchRunResult{
						Passes: 0,
						Total:  8,
					},
				},
			},
		},
	}, data)

	assert.True(t, rs[0].IsClosed())
	assert.True(t, rs[1].IsClosed())
}
