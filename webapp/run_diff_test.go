// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"strings"
	"testing"

	"github.com/deckarep/golang-set"
	"github.com/stretchr/testify/assert"
)

const mockTestPath = "/mock/path.html"

func TestDiffResults_NoDifference(t *testing.T) {
	assertNoDeltaDifferences(t, []int{0, 1}, []int{0, 1})
	assertNoDeltaDifferences(t, []int{3, 4}, []int{3, 4})
}

func TestDiffResults_Difference(t *testing.T) {
	// One test now passing
	assertDelta(t, []int{0, 1}, []int{1, 1}, []int{1, 1})

	// One test now failing
	assertDelta(t, []int{1, 1}, []int{0, 1}, []int{1, 1})

	// Two tests, one now failing
	assertDelta(t, []int{2, 2}, []int{1, 2}, []int{1, 2})

	// Three tests, two now passing
	assertDelta(t, []int{1, 3}, []int{3, 3}, []int{2, 3})
}

func TestDiffResults_Added(t *testing.T) {
	// One new test, all passing
	assertDelta(t, []int{1, 1}, []int{2, 2}, []int{1, 2})

	// One new test, all failing
	assertDelta(t, []int{0, 1}, []int{0, 2}, []int{1, 2})

	// One new test, new test passing
	assertDelta(t, []int{0, 1}, []int{1, 2}, []int{1, 2})

	// One new test, new test failing
	assertDelta(t, []int{1, 1}, []int{1, 2}, []int{1, 2})
}

func TestDiffResults_Removed(t *testing.T) {
	// One removed test, all passing
	assertDelta(t, []int{2, 2}, []int{1, 1}, []int{1, 2})

	// One removed test, all failing
	assertDelta(t, []int{0, 2}, []int{0, 1}, []int{1, 2})

	// One removed test, deleted test passing
	assertDelta(t, []int{1, 2}, []int{0, 1}, []int{1, 2})

	// One removed test, deleted test failing
	assertDelta(t, []int{1, 2}, []int{1, 1}, []int{1, 2})
}

func TestDiffResults_Filtered(t *testing.T) {
	changedFilter := DiffFilterParam{Changed: true}
	addedFilter := DiffFilterParam{Added: true}
	deletedFilter := DiffFilterParam{Deleted: true}
	const removedPath = "/mock/removed.html"
	const changedPath = "/mock/changed.html"
	const addedPath = "/mock/added.html"

	before := map[string][]int{
		removedPath: {1, 2},
		changedPath: {2, 5},
	}
	after := map[string][]int{
		changedPath: {3, 5},
		addedPath:   {1, 3},
	}
	assert.Equal(t, map[string][]int{changedPath: {1, 5}}, getResultsDiff(before, after, changedFilter))
	assert.Equal(t, map[string][]int{addedPath: {3, 3}}, getResultsDiff(before, after, addedFilter))
	assert.Equal(t, map[string][]int{removedPath: {2, 2}}, getResultsDiff(before, after, deletedFilter))

	// Test filtering by each /, /mock/, and /mock/path.html
	pieces := strings.SplitAfter(mockTestPath, "/")
	for i := 1; i < len(pieces); i++ {
		paths := mapset.NewSet(strings.Join(pieces[:i], ""))
		filter := DiffFilterParam{Changed: true, Paths: paths}
		assertDeltaWithFilter(t, []int{0, 5}, []int{5, 5}, []int{5, 5}, filter)
	}

	// Filter where none match
	rBefore, rAfter := getDeltaResultsMaps([]int{0, 5}, []int{5, 5})
	filter := DiffFilterParam{Changed: true, Paths: mapset.NewSet("/different/path/")}
	assert.Empty(t, getResultsDiff(rBefore, rAfter, filter))

	// Filter where one matches
	mockPath1, mockPath2 := "/mock/path-1.html", "/mock/path-2.html"
	rBefore = map[string][]int{
		mockPath1: {0, 1},
		mockPath2: {0, 1},
	}
	rAfter = map[string][]int{
		mockPath1: {2, 2},
		mockPath2: {2, 2},
	}
	filter.Paths = mapset.NewSet(mockPath1)
	delta := getResultsDiff(rBefore, rAfter, filter)
	assert.NotContains(t, delta, mockPath2)
	assert.Contains(t, delta, mockPath1)
	assert.Equal(t, []int{2, 2}, delta[mockPath1])
}

func assertNoDeltaDifferences(t *testing.T, before []int, after []int) {
	assertNoDeltaDifferencesWithFilter(t, before, after, DiffFilterParam{Added: true, Deleted: true, Changed: true})
}

func assertNoDeltaDifferencesWithFilter(t *testing.T, before []int, after []int, filter DiffFilterParam) {
	rBefore, rAfter := getDeltaResultsMaps(before, after)
	assert.Equal(t, map[string][]int{}, getResultsDiff(rBefore, rAfter, filter))
}

func assertDelta(t *testing.T, before []int, after []int, delta []int) {
	assertDeltaWithFilter(t, before, after, delta, DiffFilterParam{Added: true, Deleted: true, Changed: true})
}

func assertDeltaWithFilter(t *testing.T, before []int, after []int, delta []int, filter DiffFilterParam) {
	rBefore, rAfter := getDeltaResultsMaps(before, after)
	assert.Equal(
		t,
		map[string][]int{mockTestPath: delta},
		getResultsDiff(rBefore, rAfter, filter))
}

func getDeltaResultsMaps(before []int, after []int) (map[string][]int, map[string][]int) {
	return map[string][]int{mockTestPath: before},
		map[string][]int{mockTestPath: after}
}
