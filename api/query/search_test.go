// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestPrepareSearchResponse(t *testing.T) {
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

	resp := prepareSearchResponse(&filters, testRuns, summaries)
	assert.Equal(t, testRuns, resp.Runs)
	assert.Equal(t, []SearchResult{
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
	}, resp.Results)
}
