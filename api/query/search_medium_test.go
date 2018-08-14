// +build medium

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"bytes"
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
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine"
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
		ctx := appengine.NewContext(req)

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

	store := NewMockreadable(mockCtrl)

	store.EXPECT().NewReader(urls[0]).Return(bytes.NewReader(summaryBytes[0]), nil)
	store.EXPECT().NewReader(urls[01]).Return(bytes.NewReader(summaryBytes[1]), nil)

	// Same params as TestGetRunsAndFilters_specificRunIDs.
	q := "/b/"
	url := fmt.Sprintf(
		"/api/search?run_ids=%s&q=%s",
		url.QueryEscape(fmt.Sprintf("%d,%d", testRuns[0].ID, testRuns[1].ID)),
		url.QueryEscape(q))
	r, err := i.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	ctx := appengine.NewContext(r)
	w := httptest.NewRecorder()

	sh := searchHandler{
		sharedImpl: defaultShared{ctx},
		cache:      memcacheReadWritable{ctx},
		store:      store,
	}

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
		},
	}, data)
}
