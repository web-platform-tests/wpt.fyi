// +build small

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	mapset "github.com/deckarep/golang-set"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

const mockTestPath = "/mock/path.html"

func allDifferences() shared.DiffFilterParam {
	return shared.DiffFilterParam{
		Added:     true,
		Deleted:   true,
		Changed:   true,
		Unchanged: true,
	}
}

func TestDiffResults_NoDifference(t *testing.T) {
	assertNoDeltaDifferences(t, []int{0, 1}, []int{0, 1})
	assertNoDeltaDifferences(t, []int{3, 4}, []int{3, 4})
}

func TestDiffResults_Difference(t *testing.T) {
	// One test now passing
	assertDelta(t, []int{0, 1}, []int{1, 1}, []int{1, 0, 0})

	// One test now failing
	assertDelta(t, []int{1, 1}, []int{0, 1}, []int{0, 1, 0})

	// Two tests, one now failing
	assertDelta(t, []int{2, 2}, []int{1, 2}, []int{0, 1, 0})

	// Three tests, two now passing
	assertDelta(t, []int{1, 3}, []int{3, 3}, []int{2, 0, 0})
}

func TestDiffResults_Added(t *testing.T) {
	// One new test, all passing
	assertDelta(t, []int{1, 1}, []int{2, 2}, []int{1, 0, 1})

	// One new test, all failing
	assertDelta(t, []int{0, 1}, []int{0, 2}, []int{0, 1, 1})

	// One new test, new test passing
	assertDelta(t, []int{0, 1}, []int{1, 2}, []int{1, 0, 1})

	// One new test, new test failing
	assertDelta(t, []int{1, 1}, []int{1, 2}, []int{0, 1, 1})

	// Added, but it was a rename.
	renames := map[string]string{"/foo.html": "/bar.html"}
	rBefore := shared.ResultsSummary{"/foo.html": []int{1, 1}}
	rAfter := shared.ResultsSummary{"/bar.html": []int{1, 1}}
	assert.Equal(
		t,
		map[string]shared.TestDiff{"/bar.html": {0, 0, 0}},
		shared.GetResultsDiff(rBefore, rAfter, allDifferences(), nil, renames))
}

func TestDiffResults_Removed(t *testing.T) {
	// One removed test, all passing
	assertDelta(t, []int{2, 2}, []int{1, 1}, []int{0, 0, -1})

	// One removed test, all failing
	assertDelta(t, []int{0, 2}, []int{0, 1}, []int{0, 0, -1})

	// One removed test, deleted test passing
	assertDelta(t, []int{1, 2}, []int{0, 1}, []int{0, 0, -1})

	// One removed test, deleted test failing
	assertDelta(t, []int{1, 2}, []int{1, 1}, []int{0, 0, -1})
}

func TestDiffResults_Filtered(t *testing.T) {
	changedFilter := shared.DiffFilterParam{Changed: true}
	addedFilter := shared.DiffFilterParam{Added: true}
	deletedFilter := shared.DiffFilterParam{Deleted: true}
	const removedPath = "/mock/removed.html"
	const changedPath = "/mock/changed.html"
	const addedPath = "/mock/added.html"

	before := shared.ResultsSummary{
		removedPath: {1, 2},
		changedPath: {2, 5},
	}
	after := shared.ResultsSummary{
		changedPath: {3, 5},
		addedPath:   {1, 3},
	}
	assert.Equal(t, map[string]shared.TestDiff{changedPath: {1, 0, 0}}, shared.GetResultsDiff(before, after, changedFilter, nil, nil))
	assert.Equal(t, map[string]shared.TestDiff{addedPath: {1, 2, 3}}, shared.GetResultsDiff(before, after, addedFilter, nil, nil))
	assert.Equal(t, map[string]shared.TestDiff{removedPath: {0, 0, -2}}, shared.GetResultsDiff(before, after, deletedFilter, nil, nil))

	// Test filtering by each /, /mock/, and /mock/path.html
	pieces := strings.SplitAfter(mockTestPath, "/")
	for i := 1; i < len(pieces); i++ {
		paths := mapset.NewSet(strings.Join(pieces[:i], ""))
		filter := shared.DiffFilterParam{Changed: true}
		assertDeltaWithFilter(t, []int{1, 3}, []int{2, 4}, []int{1, 0, 1}, filter, paths)
	}

	// Filter where none match
	rBefore, rAfter := getDeltaResultsMaps([]int{0, 5}, []int{5, 5})
	filter := shared.DiffFilterParam{Changed: true}
	paths := mapset.NewSet("/different/path/")
	assert.Empty(t, shared.GetResultsDiff(rBefore, rAfter, filter, paths, nil))

	// Filter where one matches
	mockPath1, mockPath2 := "/mock/path-1.html", "/mock/path-2.html"
	rBefore = shared.ResultsSummary{
		mockPath1: {0, 1},
		mockPath2: {0, 1},
	}
	rAfter = shared.ResultsSummary{
		mockPath1: {2, 2},
		mockPath2: {2, 2},
	}
	delta := shared.GetResultsDiff(rBefore, rAfter, filter, mapset.NewSet(mockPath1), nil)
	assert.NotContains(t, delta, mockPath2)
	assert.Contains(t, delta, mockPath1)
	assert.Equal(t, shared.TestDiff{2, 0, 1}, delta[mockPath1])
}

func assertNoDeltaDifferences(t *testing.T, before []int, after []int) {
	assertNoDeltaDifferencesWithFilter(t, before, after, shared.DiffFilterParam{Added: true, Deleted: true, Changed: true})
}

func assertNoDeltaDifferencesWithFilter(t *testing.T, before []int, after []int, filter shared.DiffFilterParam) {
	rBefore, rAfter := getDeltaResultsMaps(before, after)
	assert.Equal(t, map[string]shared.TestDiff{}, shared.GetResultsDiff(rBefore, rAfter, filter, nil, nil))
}

func assertDelta(t *testing.T, before []int, after []int, delta []int) {
	assertDeltaWithFilter(t, before, after, delta, shared.DiffFilterParam{Added: true, Deleted: true, Changed: true}, nil)
}

func assertDeltaWithFilter(t *testing.T, before []int, after []int, delta []int, filter shared.DiffFilterParam, paths mapset.Set) {
	rBefore, rAfter := getDeltaResultsMaps(before, after)
	assert.Equal(
		t,
		map[string]shared.TestDiff{mockTestPath: delta},
		shared.GetResultsDiff(rBefore, rAfter, filter, paths, nil))
}

func getDeltaResultsMaps(before []int, after []int) (shared.ResultsSummary, shared.ResultsSummary) {
	return shared.ResultsSummary{mockTestPath: before},
		shared.ResultsSummary{mockTestPath: after}
}

func TestRegressions(t *testing.T) {
	// Note: shared.TestDiff items are {passing, regressed, total-delta}.
	regressed := shared.TestDiff{0, 1, 0}
	assert.Equal(t, 1, regressed.Regressions())
	diff := shared.ResultsDiff{"/foo.html": regressed}
	regressions := diff.Regressions()
	assert.Equal(t, 1, regressions.Cardinality())
	assert.True(t, regressions.Contains("/foo.html"))

	newlyPassed := shared.TestDiff{1, 0, 1}
	assert.Equal(t, 0, newlyPassed.Regressions())
	diff = shared.ResultsDiff{"/bar.html": newlyPassed}
	regressions = diff.Regressions()
	assert.Equal(t, 0, regressions.Cardinality())
	assert.False(t, regressions.Contains("/bar.html"))

	// A reduction in test-count is treated as though that test regressed,
	// in spite of there being zero newly-failing tests.
	droppedTests := shared.TestDiff{0, 0, -2}
	assert.Equal(t, 0, droppedTests.Regressions())
	diff = shared.ResultsDiff{"/baz.html": droppedTests}
	regressions = diff.Regressions()
	assert.Equal(t, 1, regressions.Cardinality())
	assert.True(t, regressions.Contains("/baz.html"))
}

func TestRunDiffFromSearchResponse(t *testing.T) {
	body := []byte(`{
  "runs": [{
    "id": 203760020,
    "browser_name": "safari",
    "browser_version": "82 preview",
    "os_name": "mac",
    "os_version": "10.13",
    "revision": "b5d4599280",
    "full_revision_hash": "b5d4599280363dc4e4e6a87f3706f0edce5bbdb6",
    "results_url": "https://storage.googleapis.com/wptd-staging/b5d4599280363dc4e4e6a87f3706f0edce5bbdb6/safari-82_preview-mac-10.13-be2f6871ef-summary.json.gz",
    "created_at": "2019-06-18T17:30:23.755776Z",
    "time_start": "2019-06-18T17:27:54.716Z",
    "time_end": "2019-06-18T17:29:39.042Z",
    "raw_results_url": "https://storage.googleapis.com/wptd-results-staging/b5d4599280363dc4e4e6a87f3706f0edce5bbdb6/safari-82_preview-mac-10.13-be2f6871ef/report.json",
    "labels": ["azure", "experimental", "pr_base", "preview", "safari"]
  }, {
    "id": 207750011,
    "browser_name": "safari",
    "browser_version": "82 preview",
    "os_name": "mac",
    "os_version": "10.13",
    "revision": "b5d4599280",
    "full_revision_hash": "b5d4599280363dc4e4e6a87f3706f0edce5bbdb6",
    "results_url": "https://storage.googleapis.com/wptd-staging/b5d4599280363dc4e4e6a87f3706f0edce5bbdb6/safari-82_preview-mac-10.13-b58260d2de-summary.json.gz",
    "created_at": "2019-06-18T17:33:37.68543Z",
    "time_start": "2019-06-18T17:30:57.578Z",
    "time_end": "2019-06-18T17:32:50.741Z",
    "raw_results_url": "https://storage.googleapis.com/wptd-results-staging/b5d4599280363dc4e4e6a87f3706f0edce5bbdb6/safari-82_preview-mac-10.13-b58260d2de/report.json",
    "labels": ["azure", "experimental", "pr_head", "preview", "safari"]
  }],
	"results": [
		{
			"test":"/media-source/idlharness.any.worker.html",
			"legacy_status":[{"passes":1,"total":2},{"passes":1,"total":2}]
		},
		{
			"test": "/pointerevents/idlharness.window.html",
			"legacy_status": [{ "passes": 4, "total": 5 }, { "passes": 76, "total": 84 }],
			"diff": [72, 8, 0]
		}
	]
}`)
	var scDiff shared.SearchResponse
	json.Unmarshal(body, &scDiff)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
	aeAPI.EXPECT().IsFeatureEnabled("diffRenames").Return(false)

	diff, err := shared.RunDiffFromSearchResponse(aeAPI, shared.TestRun{}, shared.TestRun{}, scDiff)
	assert.Nil(t, err)
	assert.Equal(t, 1, diff.Differences.Regressions().Cardinality())
	assert.Equal(t, 4, diff.BeforeSummary["/pointerevents/idlharness.window.html"][0])
	assert.Equal(t, 5, diff.BeforeSummary["/pointerevents/idlharness.window.html"][1])
	assert.Equal(t, 76, diff.AfterSummary["/pointerevents/idlharness.window.html"][0])
	assert.Equal(t, 84, diff.AfterSummary["/pointerevents/idlharness.window.html"][1])
}
