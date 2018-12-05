// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"testing"

	mapset "github.com/deckarep/golang-set"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	metrics "github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

const testNumShards = 16

type testRunData struct {
	run     shared.TestRun
	results *metrics.TestResultsReport
}

func mockTestRuns(loader *MockReportLoader, idx Index, data []testRunData) []shared.TestRun {
	runs := make([]shared.TestRun, len(data))
	for i, datum := range data {
		loader.EXPECT().Load(datum.run).Return(datum.results, nil)
		idx.IngestRun(datum.run)
		runs[i] = datum.run
	}
	return runs
}

func planAndExecute(t *testing.T, runs []shared.TestRun, idx Index, q query.AbstractQuery) []TestID {
	plan, err := idx.Bind(runs, q)
	assert.Nil(t, err)

	res := plan.Execute()
	ts, ok := res.([]TestID)
	assert.True(t, ok)

	return ts
}

func testSet(t *testing.T, ts []TestID) mapset.Set {
	s := mapset.NewSet()
	for _, id := range ts {
		assert.False(t, s.Contains(id))
		s.Add(id)
	}
	return s
}

func TestBindFail_NoRuns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	_, err = idx.Bind(nil, query.TestNamePattern{Pattern: "/"})
	assert.NotNil(t, err)
}

func TestBindFail_NoQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	_, err = idx.Bind([]shared.TestRun{shared.TestRun{ID: 1}}, nil)
	assert.NotNil(t, err)
}

func TestBindFail_MissingRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	_, err = idx.Bind([]shared.TestRun{shared.TestRun{ID: 1}}, query.TestNamePattern{Pattern: "/"})
	assert.NotNil(t, err)
}

func TestBindExecute_TestNamePattern(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		testRunData{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					&metrics.TestResults{
						Test:   matchingTestName,
						Status: "PASS",
					},
					&metrics.TestResults{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
	})

	q := query.TestNamePattern{
		Pattern: "/a",
	}
	ts := planAndExecute(t, runs, idx, q)

	expectedTestID, err := computeTestID(matchingTestName, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ts))
	assert.Equal(t, expectedTestID, ts[0])
}

func TestBindExecute_TestStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	// Two matching (sub)tests Chrome(Status=FAIL):
	// <"/a/b/c", nil> and </"d/e/f", "sub">.
	match1Name := "/a/b/c"
	match2Name := "/d/e/f"
	match2Sub := "sub"
	runs := mockTestRuns(loader, idx, []testRunData{
		//
		// Chrome test run.
		//
		testRunData{
			shared.TestRun{
				ID: 1,
				ProductAtRevision: shared.ProductAtRevision{
					Product: shared.Product{
						BrowserName: "Chrome",
					},
				},
			},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					&metrics.TestResults{
						Test:   match1Name,
						Status: "FAIL",
					},
					&metrics.TestResults{
						Test:   match2Name,
						Status: "OK",
						Subtests: []metrics.SubTest{
							metrics.SubTest{
								Name:   match2Sub,
								Status: "FAIL",
							},
							metrics.SubTest{
								Name:   "other sub",
								Status: "PASS",
							},
						},
					},
					&metrics.TestResults{
						Test:   "m/n/o",
						Status: "TIMEOUT",
					},
					&metrics.TestResults{
						Test:   "x/y/z",
						Status: "OK",
						Subtests: []metrics.SubTest{
							metrics.SubTest{
								Name:   "last sub",
								Status: "PASS",
							},
						},
					},
				},
			},
		},
		//
		// Safari test run: Several result values differ or are missing. One test
		// does not appear in Chrome, but does appear here.
		//
		testRunData{
			shared.TestRun{
				ID: 2,
				ProductAtRevision: shared.ProductAtRevision{
					Product: shared.Product{
						BrowserName: "Safari",
					},
				},
			},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					&metrics.TestResults{
						Test:   match1Name,
						Status: "PASS",
					},
					&metrics.TestResults{
						Test:   match2Name,
						Status: "OK",
						Subtests: []metrics.SubTest{
							metrics.SubTest{
								Name:   "other sub",
								Status: "FAIL",
							},
						},
					},
					&metrics.TestResults{
						Test:   "x/y/z",
						Status: "OK",
						Subtests: []metrics.SubTest{
							metrics.SubTest{
								Name:   "last sub",
								Status: "TIMEOUT",
							},
						},
					},
					&metrics.TestResults{
						Test:   "/safari/only",
						Status: "FAIL",
					},
				},
			},
		},
	})

	q := query.TestStatusConstraint{
		BrowserName: "Chrome",
		Status:      shared.TestStatusFail,
	}
	ts := planAndExecute(t, runs, idx, q)

	id1, err := computeTestID(match1Name, nil)
	assert.Nil(t, err)
	id2, err := computeTestID(match2Name, &match2Sub)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(ts))
	assert.Equal(t, testSet(t, []TestID{id1, id2}), testSet(t, ts))
}
