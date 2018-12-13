// +build small

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"strings"
	"testing"

	"github.com/deckarep/golang-set"
	"github.com/stretchr/testify/assert"
)

const mockTestPath = "/mock/path.html"

func allDifferences() DiffFilterParam {
	return DiffFilterParam{
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
	rBefore := ResultsSummary{"/foo.html": []int{1, 1}}
	rAfter := ResultsSummary{"/bar.html": []int{1, 1}}
	assert.Equal(
		t,
		map[string]TestDiff{"/bar.html": {0, 0, 0}},
		GetResultsDiff(rBefore, rAfter, allDifferences(), nil, renames))
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
	changedFilter := DiffFilterParam{Changed: true}
	addedFilter := DiffFilterParam{Added: true}
	deletedFilter := DiffFilterParam{Deleted: true}
	const removedPath = "/mock/removed.html"
	const changedPath = "/mock/changed.html"
	const addedPath = "/mock/added.html"

	before := ResultsSummary{
		removedPath: {1, 2},
		changedPath: {2, 5},
	}
	after := ResultsSummary{
		changedPath: {3, 5},
		addedPath:   {1, 3},
	}
	assert.Equal(t, map[string]TestDiff{changedPath: {1, 0, 0}}, GetResultsDiff(before, after, changedFilter, nil, nil))
	assert.Equal(t, map[string]TestDiff{addedPath: {1, 2, 3}}, GetResultsDiff(before, after, addedFilter, nil, nil))
	assert.Equal(t, map[string]TestDiff{removedPath: {0, 0, -2}}, GetResultsDiff(before, after, deletedFilter, nil, nil))

	// Test filtering by each /, /mock/, and /mock/path.html
	pieces := strings.SplitAfter(mockTestPath, "/")
	for i := 1; i < len(pieces); i++ {
		paths := mapset.NewSet(strings.Join(pieces[:i], ""))
		filter := DiffFilterParam{Changed: true}
		assertDeltaWithFilter(t, []int{1, 3}, []int{2, 4}, []int{1, 0, 1}, filter, paths)
	}

	// Filter where none match
	rBefore, rAfter := getDeltaResultsMaps([]int{0, 5}, []int{5, 5})
	filter := DiffFilterParam{Changed: true}
	paths := mapset.NewSet("/different/path/")
	assert.Empty(t, GetResultsDiff(rBefore, rAfter, filter, paths, nil))

	// Filter where one matches
	mockPath1, mockPath2 := "/mock/path-1.html", "/mock/path-2.html"
	rBefore = ResultsSummary{
		mockPath1: {0, 1},
		mockPath2: {0, 1},
	}
	rAfter = ResultsSummary{
		mockPath1: {2, 2},
		mockPath2: {2, 2},
	}
	delta := GetResultsDiff(rBefore, rAfter, filter, mapset.NewSet(mockPath1), nil)
	assert.NotContains(t, delta, mockPath2)
	assert.Contains(t, delta, mockPath1)
	assert.Equal(t, TestDiff{2, 0, 1}, delta[mockPath1])
}

func assertNoDeltaDifferences(t *testing.T, before []int, after []int) {
	assertNoDeltaDifferencesWithFilter(t, before, after, DiffFilterParam{Added: true, Deleted: true, Changed: true})
}

func assertNoDeltaDifferencesWithFilter(t *testing.T, before []int, after []int, filter DiffFilterParam) {
	rBefore, rAfter := getDeltaResultsMaps(before, after)
	assert.Equal(t, map[string]TestDiff{}, GetResultsDiff(rBefore, rAfter, filter, nil, nil))
}

func assertDelta(t *testing.T, before []int, after []int, delta []int) {
	assertDeltaWithFilter(t, before, after, delta, DiffFilterParam{Added: true, Deleted: true, Changed: true}, nil)
}

func assertDeltaWithFilter(t *testing.T, before []int, after []int, delta []int, filter DiffFilterParam, paths mapset.Set) {
	rBefore, rAfter := getDeltaResultsMaps(before, after)
	assert.Equal(
		t,
		map[string]TestDiff{mockTestPath: delta},
		GetResultsDiff(rBefore, rAfter, filter, paths, nil))
}

func getDeltaResultsMaps(before []int, after []int) (ResultsSummary, ResultsSummary) {
	return ResultsSummary{mockTestPath: before},
		ResultsSummary{mockTestPath: after}
}

func TestRegressions(t *testing.T) {
	// Note: TestDiff items are {passing, regressed, total-delta}.
	regressed := TestDiff{0, 1, 0}
	assert.Equal(t, 1, regressed.Regressions())
	diff := ResultsDiff{"/foo.html": regressed}
	regressions := diff.Regressions()
	assert.Equal(t, 1, regressions.Cardinality())
	assert.True(t, regressions.Contains("/foo.html"))

	newlyPassed := TestDiff{1, 0, 1}
	assert.Equal(t, 0, newlyPassed.Regressions())
	diff = ResultsDiff{"/bar.html": newlyPassed}
	regressions = diff.Regressions()
	assert.Equal(t, 0, regressions.Cardinality())
	assert.False(t, regressions.Contains("/bar.html"))

	// A reduction in test-count is treated as though that test regressed,
	// in spite of there being zero newly-failing tests.
	droppedTests := TestDiff{0, 0, -2}
	assert.Equal(t, 0, droppedTests.Regressions())
	diff = ResultsDiff{"/baz.html": droppedTests}
	regressions = diff.Regressions()
	assert.Equal(t, 1, regressions.Cardinality())
	assert.True(t, regressions.Contains("/baz.html"))
}
