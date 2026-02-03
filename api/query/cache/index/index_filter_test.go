//go:build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"encoding/json"
	"testing"

	mapset "github.com/deckarep/golang-set"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
	metrics "github.com/web-platform-tests/wpt.fyi/shared/metrics"
	"go.uber.org/mock/gomock"
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

func planAndExecute(t *testing.T, runs []shared.TestRun, idx Index, q query.AbstractQuery) []shared.SearchResult {
	plan, err := idx.Bind(runs, q.BindToRuns(runs...))
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	return srs
}

func resultSet(t *testing.T, srs []shared.SearchResult) mapset.Set {
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

	_, err = idx.Bind([]shared.TestRun{{ID: 1}}, nil)
	assert.NotNil(t, err)
}

func TestBindFail_MissingRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	runs := []shared.TestRun{{ID: 1}}
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
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
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
	expectedResult := shared.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}
	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_TestNamePattern_CaseInsensitive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/custom-elements/Document-createElement-customized-builtins.html"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
				},
			},
		},
	})

	for _, pattern := range []string{"custom", "Custom", "CUSTOM", "createelement", "createElement", "CREATEELEMENT"} {
		t.Run("pattern: "+pattern, func(t *testing.T) {
			q := query.TestNamePattern{
				Pattern: pattern,
			}
			srs := planAndExecute(t, runs, idx, q)

			assert.Equal(t, 1, len(srs))
			assert.Equal(t, matchingTestName, srs[0].Test)
		})
	}
}

func TestBindExecute_SubtestNamePattern(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   "/a/b/c",
						Status: "OK",
						Subtests: []metrics.SubTest{
							{
								Name:   "a1",
								Status: "PASS",
							},
							{
								Name:   "a2",
								Status: "FAIL",
							},
						},
					},
					{
						Test:   "/d/e/f",
						Status: "TIMEOUT",
						Subtests: []metrics.SubTest{
							{
								Name:   "d1",
								Status: "PASS",
							},
							{
								Name:   "d2",
								Status: "FAIL",
							},
							{
								Name:   "d3",
								Status: "TIMEOUT",
							},
						},
					},
				},
			},
		},
	})

	for _, testCase := range []struct {
		Subtest string
		Passes  int
		Total   int
	}{
		{"a1", 1, 1},
		{"a", 1, 2},
	} {
		t.Run("subtest: "+testCase.Subtest, func(t *testing.T) {
			q := query.SubtestNamePattern{
				Subtest: testCase.Subtest,
			}
			srs := planAndExecute(t, runs, idx, q)

			assert.Equal(t, 1, len(srs))
			expectedResult := shared.SearchResult{
				Test: "/a/b/c",
				LegacyStatus: []shared.LegacySearchRunResult{
					{
						Passes:        testCase.Passes, // Only matches the subtest.
						Total:         testCase.Total,
						Status:        "",
						NewAggProcess: true,
					},
				},
			}
			assert.Equal(t, expectedResult, srs[0])
		})
	}
}
func TestBindExecute_SubtestNamePattern_CaseInsensitive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   "/test.html",
						Status: "OK",
						Subtests: []metrics.SubTest{
							{
								Name:   "TestCase1",
								Status: "PASS",
							},
							{
								Name:   "TestCase2",
								Status: "PASS",
							},
							{
								Name:   "OtherTest",
								Status: "FAIL",
							},
						},
					},
				},
			},
		},
	})

	for _, pattern := range []string{"testcase", "TestCase", "TESTCASE"} {
		t.Run("pattern: "+pattern, func(t *testing.T) {
			q := query.SubtestNamePattern{
				Subtest: pattern,
			}
			srs := planAndExecute(t, runs, idx, q)

			assert.Equal(t, 1, len(srs))
			expectedResult := shared.SearchResult{
				Test: "/test.html",
				LegacyStatus: []shared.LegacySearchRunResult{
					{
						Passes:        2, // Both TestCase1 and TestCase2 should match
						Total:         2,
						Status:        "",
						NewAggProcess: true,
					},
				},
			}
			assert.Equal(t, expectedResult, srs[0])
		})
	}

}

func TestBindExecute_TestPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingPath := "/dom/"
	unmatchingPath := "/html/dom/"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingPath,
						Status: "PASS",
					},
					{
						Test:   unmatchingPath,
						Status: "FAIL",
					},
				},
			},
		},
	})

	q := query.TestPath{
		Path: "/dom/",
	}
	srs := planAndExecute(t, runs, idx, q)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: matchingPath,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
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
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   match1Name,
						Status: "FAIL",
					},
					{
						Test:   match2Name,
						Status: "OK",
						Subtests: []metrics.SubTest{
							{
								Name:   match2Sub,
								Status: "FAIL",
							},
							{
								Name:   "other sub",
								Status: "PASS",
							},
						},
					},
					{
						Test:   "m/n/o",
						Status: "TIMEOUT",
					},
					{
						Test:   "x/y/z",
						Status: "OK",
						Subtests: []metrics.SubTest{
							{
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
		{
			shared.TestRun{ID: 2},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   match1Name,
						Status: "PASS",
					},
					{
						Test:   match2Name,
						Status: "OK",
						Subtests: []metrics.SubTest{
							{
								Name:   "other sub",
								Status: "FAIL",
							},
						},
					},
					{
						Test:   "x/y/z",
						Status: "OK",
						Subtests: []metrics.SubTest{
							{
								Name:   "last sub",
								Status: "TIMEOUT",
							},
						},
					},
					{
						Test:   "/safari/only",
						Status: "FAIL",
					},
				},
			},
		},
	}

	// Set BrowserName imperatively to avoid multi-layer type embedding.
	data[0].run.BrowserName = "chrome"
	data[1].run.BrowserName = "safari"

	runs := mockTestRuns(loader, idx, data)

	p := shared.ParseProductSpecUnsafe("Chrome")
	q := query.TestStatusEq{
		Product: &p,
		Status:  shared.TestStatusFail,
	}
	srs := planAndExecute(t, runs, idx, q)

	assert.Equal(t, 2, len(srs))
	assert.Equal(t, resultSet(t, []shared.SearchResult{
		{
			Test: match1Name,
			LegacyStatus: []shared.LegacySearchRunResult{
				// Run [0]: Chrome: match1Name status is FAIL: 0 / 1.
				{
					Passes:        0,
					Total:         1,
					Status:        "",
					NewAggProcess: true,
				},
				// Run [1]: Safari: match1Name status is PASS: 1 / 1.
				{
					Passes:        1,
					Total:         1,
					Status:        "",
					NewAggProcess: true,
				},
			},
		},
		{
			Test: match2Name,
			// Run [0]: Chrome: match1Name.match2Sub status is FAIL,
			//                  and no other subtests match: 0 / 1.
			LegacyStatus: []shared.LegacySearchRunResult{
				{
					Passes:        0,
					Total:         1,
					Status:        "",
					NewAggProcess: true,
				},
				// Run [1]: Safari: match1Name.match2Sub is missing;
				//                  by logic used in legacy test summaries, result
				//                  should be: 0 / 0.
				{
					Passes:        0,
					Total:         0,
					Status:        "",
					NewAggProcess: false,
				},
			},
		},
	}), resultSet(t, srs))
}

func TestBindExecute_Link(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
	})
	metadata := map[string][]string{"/foo/bar/b.html": {
		"https://bug.com/item", "https://bug.com/item", "https://bug.com/item"},
		matchingTestName: {"", "https://external.com/item", ""},
	}

	link := query.Link{Pattern: "external", Metadata: metadata}
	plan, err := idx.Bind(runs, link)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_LinkWithWildcards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
	})
	metadata := map[string][]string{
		"/foo/bar/b.html": {"https://bug.com/item", "https://bug.com/item", "https://bug.com/item"},
		"/a/*":            {"", "https://external.com/item", ""},
	}

	// Create an execute a plan for `link:external`. Inside the metadata
	// this matches the wildcard "/a/*". When mapped to test runs, that
	// means it should match "/a/b/c" due to wildcard expansion.
	link := query.Link{Pattern: "external", Metadata: metadata}
	plan, err := idx.Bind(runs, link)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}
	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_Triaged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
	})
	metadata := map[string][]string{"/foo/bar/b.html": {
		""},
		matchingTestName: {"https://bug.com/item"},
		"/d/e/f":         {""},
	}

	link := query.Or{Args: []query.ConcreteQuery{query.Triaged{Run: 1, Metadata: metadata}}}
	plan, err := idx.Bind(runs, link)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_TriagedWildcards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
	})
	metadata := map[string][]string{"/foo/bar/b.html": {
		""},
		"/a/*":   {"https://bug.com/item", "https://bug.com/item1"},
		"/d/e/f": {""},
	}

	link := query.Or{Args: []query.ConcreteQuery{query.Triaged{Run: 1, Metadata: metadata}}}
	plan, err := idx.Bind(runs, link)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_QueryAndTestLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
	})
	metadata := map[string][]string{
		"/foo/bar/b.html": {"random"},
		matchingTestName:  {"interop1", "INTEROP2"},
		"/d/e/f":          {""},
	}

	// It is equivalent to searching "label:interop1 & label:interop2".
	testlabel := query.And{[]query.ConcreteQuery{query.TestLabel{Label: "interop2", Metadata: metadata}, query.TestLabel{Label: "interop1", Metadata: metadata}}}
	plan, err := idx.Bind(runs, testlabel)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_QueryOrTestLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "PASS",
					},
				},
			},
		},
	})
	metadata := map[string][]string{
		"/foo/bar/b.html": {"random"},
		matchingTestName:  {"INTEROP2"},
		"/d/e/f":          {"interop1"},
	}

	// It is equivalent to searching "label:interop1 | label:interop2".
	testlabel := query.Or{[]query.ConcreteQuery{query.TestLabel{Label: "interop2", Metadata: metadata}, query.TestLabel{Label: "interop1", Metadata: metadata}}}
	plan, err := idx.Bind(runs, testlabel)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	expectedResult := shared.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	expectedResult1 := shared.SearchResult{
		Test: "/d/e/f",
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}
	assert.Equal(t, 2, len(srs))
	assert.Contains(t, srs, expectedResult)
	assert.Contains(t, srs, expectedResult1)
}

func TestBindExecute_TestLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
	})
	metadata := map[string][]string{
		"/foo/bar/b.html": {"random"},
		matchingTestName:  {"interop1", "INTEROP2"},
		"/d/e/f":          {""},
	}

	testlabel := query.TestLabel{Label: "interop2", Metadata: metadata}
	plan, err := idx.Bind(runs, testlabel)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_LabelWithWildcards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
	})
	metadata := map[string][]string{
		"/foo/bar/b.html": {"random"},
		"/a/*":            {"interop1", "INTEROP2"},
		"/d/e/f":          {""},
		matchingTestName:  {"foo"},
	}

	// Create an execute a plan for `label:interop1 & label:interop2`. Inside the metadata
	// this matches the wildcard "/a/*". When mapped to test runs, that
	// means it should match "/a/b/c" due to wildcard expansion.
	testlabel := query.And{[]query.ConcreteQuery{query.TestLabel{Label: "interop2", Metadata: metadata}, query.TestLabel{Label: "interop1", Metadata: metadata}}}
	plan, err := idx.Bind(runs, testlabel)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_TestWebFeature(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
	})
	data := shared.WebFeaturesData{
		"/foo/bar/b.html": {"random": nil},
		matchingTestName:  {"avif": nil, "grid": nil},
		"/d/e/f":          {"": nil},
	}

	testlabel := query.TestWebFeature{WebFeature: "grid", WebFeaturesData: data}
	plan, err := idx.Bind(runs, testlabel)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				// Only matching test passes.
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_TestWebFeature_PreservesCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/custom-elements/Document-createElement-customized-builtins.html"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/custom-elements/other-test.html",
						Status: "FAIL",
					},
				},
			},
		},
	})

	data := shared.WebFeaturesData{
		matchingTestName: {"customized-built-in-elements": nil},
		"/custom-elements/other-test.html": {"autonomous-custom-elements": nil},
	}

	testlabel := query.TestWebFeature{WebFeature: "customized-built-in-elements", WebFeaturesData: data}
	plan, err := idx.Bind(runs, testlabel)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: matchingTestName,
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_IsDifferent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   "/a/b/c",
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
		{
			shared.TestRun{ID: 2},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   "/a/b/c",
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "PASS",
					},
				},
			},
		},
	})

	quality := query.MetadataQualityDifferent
	plan, err := idx.Bind(runs, quality)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: "/d/e/f", // /a/b/c was the same, /d/e/f differed.
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				Passes:        0,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
			{
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_IsTentative(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   "/a/b/c.html",
						Status: "PASS",
					},
					{
						Test:   "/a/b/c.tentative.html",
						Status: "PASS",
					},
					{
						Test:   "/a/b/tentative/c.html",
						Status: "PASS",
					},
					{
						Test:   "/a/b/tentative/c.tentative.html",
						Status: "PASS",
					},
				},
			},
		},
	})

	quality := query.MetadataQualityTentative
	plan, err := idx.Bind(runs, quality)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 3, len(srs))
	assert.Equal(t, resultSet(t, []shared.SearchResult{
		{
			Test: "/a/b/c.tentative.html",
			LegacyStatus: []shared.LegacySearchRunResult{
				{
					Passes:        1,
					Total:         1,
					Status:        "",
					NewAggProcess: true,
				},
			},
		},
		{
			Test: "/a/b/tentative/c.html",
			LegacyStatus: []shared.LegacySearchRunResult{
				{
					Passes:        1,
					Total:         1,
					Status:        "",
					NewAggProcess: true,
				},
			},
		},
		{
			Test: "/a/b/tentative/c.tentative.html",
			LegacyStatus: []shared.LegacySearchRunResult{
				{
					Passes:        1,
					Total:         1,
					Status:        "",
					NewAggProcess: true,
				},
			},
		},
	}), resultSet(t, srs))
}

func TestBindExecute_IsOptional(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   "/a/b/c",
						Status: "PASS",
					},
					{
						Test:   "/a/b/c.optional.html",
						Status: "PASS",
					},
				},
			},
		},
	})

	quality := query.MetadataQualityOptional
	plan, err := idx.Bind(runs, quality)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: "/a/b/c.optional.html",
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_MoreThan(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   "/a/b/c",
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
		{
			shared.TestRun{ID: 2},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   "/a/b/c",
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "PASS",
					},
				},
			},
		},
	})

	moreThan := query.AbstractMoreThan{
		query.AbstractCount{
			Count: 1,
			Where: query.TestStatusEq{Status: shared.TestStatusPass},
		},
	}.BindToRuns(runs...)
	plan, err := idx.Bind(runs, moreThan)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: "/a/b/c", // /a/b/c has 2 passes.
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
			{
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_LessThan(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   "/a/b/c",
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
		{
			shared.TestRun{ID: 2},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   "/a/b/c",
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "PASS",
					},
				},
			},
		},
	})

	moreThan := query.AbstractLessThan{
		query.AbstractCount{
			Count: 2,
			Where: query.TestStatusEq{Status: shared.TestStatusPass},
		},
	}.BindToRuns(runs...)
	plan, err := idx.Bind(runs, moreThan)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: "/d/e/f", // /a/b/c has 1 passes.
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				Passes:        0,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
			{
				Passes:        1,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_LinkNoMatchingPattern(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	noMatchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   noMatchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
	})
	metadata := map[string][]string{
		"/foo/bar/b.html":  {"https://bug.com/item", "https://bug.com/item", "https://bug.com/item"},
		noMatchingTestName: {"", "https://external.com/item", ""},
	}

	link := query.Link{Pattern: "NoMatchingPattern", Metadata: metadata}
	plan, err := idx.Bind(runs, link)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 0, len(srs))
}

func TestBindExecute_NotLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "PASS",
					},
					{
						Test:   "/d/e/f",
						Status: "FAIL",
					},
				},
			},
		},
	})
	metadata := map[string][]string{
		"/foo/bar/b.html": {"https://bug.com/item", "https://bug.com/item", "https://bug.com/item"},
		matchingTestName:  {"", "https://external.com/item", ""},
	}

	notQuery := query.Not{Arg: query.Link{Pattern: "external", Metadata: metadata}}
	plan, err := idx.Bind(runs, notQuery)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: "/d/e/f",
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				Passes:        0,
				Total:         1,
				Status:        "",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}

func TestBindExecute_HandleHarness(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	idx, err := NewShardedWPTIndex(loader, testNumShards)
	assert.Nil(t, err)

	matchingTestName := "/a/b/c"
	runs := mockTestRuns(loader, idx, []testRunData{
		{
			shared.TestRun{ID: 1},
			&metrics.TestResultsReport{
				Results: []*metrics.TestResults{
					{
						Test:   matchingTestName,
						Status: "FAIL",
					},
					{
						Test:   "/d/e/f",
						Status: "OK",
					},
				},
			},
		},
	})
	metadata := map[string][]string{
		"/foo/bar/b.html": {"https://bug.com/item", "https://bug.com/item", "https://bug.com/item"},
		matchingTestName:  {"", "https://external.com/item", ""},
	}

	notQuery := query.Not{Arg: query.Link{Pattern: "external", Metadata: metadata}}
	plan, err := idx.Bind(runs, notQuery)
	assert.Nil(t, err)

	res := plan.Execute(runs, query.AggregationOpts{})
	srs, ok := res.([]shared.SearchResult)
	assert.True(t, ok)

	assert.Equal(t, 1, len(srs))
	expectedResult := shared.SearchResult{
		Test: "/d/e/f",
		LegacyStatus: []shared.LegacySearchRunResult{
			{
				Passes:        0,
				Total:         0,
				Status:        "O",
				NewAggProcess: true,
			},
		},
	}

	assert.Equal(t, expectedResult, srs[0])
}
