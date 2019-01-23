// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"encoding/json"
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

func planAndExecute(t *testing.T, runs []shared.TestRun, idx Index, q query.AbstractQuery) []query.SearchResult {
	plan, err := idx.Bind(runs, q.BindToRuns(runs...))
	assert.Nil(t, err)

	res := plan.Execute(runs)
	srs, ok := res.([]query.SearchResult)
	assert.True(t, ok)

	return srs
}

func resultSet(t *testing.T, srs []query.SearchResult) mapset.Set {
	s := mapset.NewSet()
	for _, sr := range srs {
		// TODO: The json package should be unnecessary, but for some reason a
		// {Test: <string>, Results: []{Passes: <int>, Total: <int>}} is not
		// hashable.
		data, err := json.Marshal(sr)
		assert.Nil(t, err)
		str := string(data)
		assert.False(t, s.Contains(str))
		s.Add(str)
	}
	return s
}

func TestBindFail_NoRuns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	_, err = idx.Bind(nil, query.TestNamePattern{Pattern: "/"}.BindToRuns())
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

	runs := []shared.TestRun{shared.TestRun{ID: 1}}
	_, err = idx.Bind(runs, query.TestNamePattern{Pattern: "/"}.BindToRuns(runs...))
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
	srs := planAndExecute(t, runs, idx, q)

	assert.Equal(t, 1, len(srs))
	expectedResult := query.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []query.LegacySearchRunResult{
			query.LegacySearchRunResult{
				// Only matching test passes.
				Passes: 1,
				Total:  1,
			},
		},
	}
	assert.Equal(t, expectedResult, srs[0])
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
	data := []testRunData{
		//
		// [0]: Chrome test run.
		//
		testRunData{
			shared.TestRun{ID: 1},
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
		// [1] Safari test run: Several result values differ or are missing. One
		//     test does not appear in Chrome, but does appear here.
		//
		testRunData{
			shared.TestRun{ID: 2},
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
	}

	// Set BrowserName imperatively to avoid multi-layer type embedding.
	data[0].run.BrowserName = "Chrome"
	data[1].run.BrowserName = "Safari"

	runs := mockTestRuns(loader, idx, data)

	q := query.TestStatusEq{
		BrowserName: "Chrome",
		Status:      shared.TestStatusFail,
	}
	srs := planAndExecute(t, runs, idx, q)

	assert.Equal(t, 2, len(srs))
	assert.Equal(t, resultSet(t, []query.SearchResult{
		query.SearchResult{
			Test: match1Name,
			LegacyStatus: []query.LegacySearchRunResult{
				// Run [0]: Chrome: match1Name status is FAIL: 0 / 1.
				query.LegacySearchRunResult{
					Passes: 0,
					Total:  1,
				},
				// Run [1]: Safari: match1Name status is PASS: 1 / 1.
				query.LegacySearchRunResult{
					Passes: 1,
					Total:  1,
				},
			},
		},
		query.SearchResult{
			Test: match2Name,
			// Run [0]: Chrome: match1Name.match2Sub status is FAIL,
			//                  and no other subtests match: 0 / 1.
			LegacyStatus: []query.LegacySearchRunResult{
				query.LegacySearchRunResult{
					Passes: 0,
					Total:  1,
				},
				// Run [1]: Safari: match1Name.match2Sub is missing;
				//                  by logic used in legacy test summaries, result
				//                  should be: 0 / 0.
				query.LegacySearchRunResult{
					Passes: 0,
					Total:  0,
				},
			},
		},
	}), resultSet(t, srs))
}
