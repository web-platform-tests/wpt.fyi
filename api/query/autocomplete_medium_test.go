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
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func TestAutocompleteHandler(t *testing.T) {
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
	rs := []*MockReadCloser{
		NewMockReadCloser(t, summaryBytes[0]),
		NewMockReadCloser(t, summaryBytes[1]),
	}

	store.EXPECT().NewReadCloser(urls[0]).Return(rs[0], nil)
	store.EXPECT().NewReadCloser(urls[1]).Return(rs[1], nil)

	// Same params as TestPrepareAutocompleteResponse_several.
	q := "/b/"
	url := fmt.Sprintf(
		"/api/autocomplete?run_ids=%s&q=%s",
		url.QueryEscape(fmt.Sprintf("%d,%d", testRuns[0].ID, testRuns[1].ID)),
		url.QueryEscape(q))
	r, err := i.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	ctx := appengine.NewContext(r)
	w := httptest.NewRecorder()

	sh := autocompleteHandler{queryHandler{
		sharedImpl: defaultShared{ctx},
		dataSource: cachedStore{
			cache: memcacheReadWritable{ctx},
			store: store,
		},
	}}

	sh.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, err := ioutil.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	var data AutocompleteResponse
	err = json.Unmarshal(bytes, &data)
	assert.Nil(t, err)

	// Same result as TestPrepareAutocompleteResponse_several.
	assert.Equal(t, AutocompleteResponse{
		Suggestions: []AutocompleteResult{
			AutocompleteResult{"/b/c"},
			AutocompleteResult{"/a/b/c"},
			AutocompleteResult{"/z/b/c"},
		},
	}, data)
	assert.True(t, rs[0].closed)
	assert.True(t, rs[1].closed)
}
