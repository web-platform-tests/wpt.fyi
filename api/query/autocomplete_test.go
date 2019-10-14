// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestByQueryIndex_standard(t *testing.T) {
	bqi := byQueryIndex{
		q: "a",
		rs: []AutocompleteResult{
			AutocompleteResult{
				QueryString: "ab",
			},
			AutocompleteResult{
				QueryString: "ba",
			},
		},
	}
	assert.Equal(t, 2, bqi.Len())
	assert.True(t, bqi.Less(0, 1))
	assert.False(t, bqi.Less(1, 0))

	bqi.Swap(0, 1)
	assert.Equal(t, AutocompleteResult{
		QueryString: "ba",
	}, bqi.rs[0])
	assert.Equal(t, AutocompleteResult{
		QueryString: "ab",
	}, bqi.rs[1])
}

func TestByQueryIndex_lexFallback(t *testing.T) {
	bqi := byQueryIndex{
		q: "a",
		rs: []AutocompleteResult{
			AutocompleteResult{
				QueryString: "xab",
			},
			AutocompleteResult{
				QueryString: "yab",
			},
		},
	}
	assert.True(t, bqi.Less(0, 1))
	assert.False(t, bqi.Less(1, 0))
}

func TestByQueryIndex_same(t *testing.T) {
	bqi := byQueryIndex{
		rs: []AutocompleteResult{
			AutocompleteResult{},
			AutocompleteResult{},
		},
	}
	assert.False(t, bqi.Less(0, 1))
	assert.False(t, bqi.Less(1, 0))
}

func TestParseLimit_none(t *testing.T) {
	ah := autocompleteHandler{queryHandler: queryHandler{}}
	limit, err := ah.parseLimit(httptest.NewRequest("GET", "/api/autocomplete", nil))
	assert.Equal(t, autocompleteDefaultLimit, limit)
	assert.Nil(t, err)
}

func TestParseLimit_tooSmall(t *testing.T) {
	ah := autocompleteHandler{queryHandler: queryHandler{}}
	limit, err := ah.parseLimit(httptest.NewRequest("GET", fmt.Sprintf("/api/autocomplete?limit=%d", autocompleteMinLimit-1), nil))
	assert.Equal(t, autocompleteMinLimit, limit)
	assert.Nil(t, err)
}

func TestParseLimit_tooBig(t *testing.T) {
	ah := autocompleteHandler{queryHandler: queryHandler{}}
	limit, err := ah.parseLimit(httptest.NewRequest("GET", fmt.Sprintf("/api/autocomplete?limit=%d", autocompleteMaxLimit+1), nil))
	assert.Equal(t, autocompleteMaxLimit, limit)
	assert.Nil(t, err)
}

func TestParseLimit_bad(t *testing.T) {
	ah := autocompleteHandler{queryHandler: queryHandler{}}
	_, err := ah.parseLimit(httptest.NewRequest("GET", "/api/autocomplete?limit=notanumber", nil))
	assert.NotNil(t, err)
}

func TestParseLimit_ok(t *testing.T) {
	ah := autocompleteHandler{queryHandler: queryHandler{}}
	limit, err := ah.parseLimit(httptest.NewRequest("GET", fmt.Sprintf("/api/autocomplete?limit=%d", autocompleteMaxLimit-1), nil))
	assert.Equal(t, autocompleteMaxLimit-1, limit)
	assert.Nil(t, err)
}

func TestPrepareAutocompleteResponse_none(t *testing.T) {
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
	filters := shared.QueryFilter{
		RunIDs: runIDs,
		Q:      "/appears_nowhere/",
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

	resp := prepareAutocompleteResponse(50, &filters, testRuns, summaries)
	assert.Equal(t, []AutocompleteResult{}, resp.Suggestions)
}

func TestPrepareAutocompleteResponse_several(t *testing.T) {
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
	filters := shared.QueryFilter{
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

	resp := prepareAutocompleteResponse(50, &filters, testRuns, summaries)
	assert.Equal(t, []AutocompleteResult{
		AutocompleteResult{"/b/c"},
		AutocompleteResult{"/a/b/c"},
		AutocompleteResult{"/z/b/c"},
	}, resp.Suggestions)
}

func TestPrepareAutocompleteResponse_limited(t *testing.T) {
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
	filters := shared.QueryFilter{
		RunIDs: runIDs,
		Q:      "/B/",
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

	resp := prepareAutocompleteResponse(2, &filters, testRuns, summaries)
	assert.Equal(t, []AutocompleteResult{
		AutocompleteResult{"/b/c"},
		AutocompleteResult{"/a/b/c"},
	}, resp.Suggestions)
}
