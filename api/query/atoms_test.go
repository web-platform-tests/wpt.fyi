//go:build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"go.uber.org/mock/gomock"
)

func TestStructuredQuery_empty(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_missingRunIDs(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"query": {
			"pattern": "/2dcontext/"
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_missingQuery(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2]
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: True{}}, rq)
}

func TestStructuredQuery_emptyRunIDs(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [],
		"query": {
			"pattern": "/2dcontext/"
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_emptyBrowserName(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"status": "PASS"
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: TestStatusEq{Status: shared.TestStatusValueFromString("PASS")}}, rq)
}

func TestStructuredQuery_missingStatus(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"product": "chrome"
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_badStatus(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"product": "chrome",
			"status": "NOT_A_REAL_STATUS"
		}
	}`), &rq)
	assert.NotNil(t, err)
}
func TestStructuredQuery_unknownStatus(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"product": "chrome",
			"status": "UNKNOWN"
		}
	}`), &rq)
	assert.Nil(t, err)
	p := shared.ParseProductSpecUnsafe("chrome")
	assert.Equal(t, RunQuery{
		RunIDs:        []int64{0, 1, 2},
		AbstractQuery: TestStatusEq{&p, shared.TestStatusValueFromString("UNKNOWN")},
	}, rq)
}

func TestStructuredQuery_missingPattern(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_emptyPattern(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"pattern": ""
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: TestNamePattern{""}}, rq)
}

func TestStructuredQuery_pattern(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"pattern": "/2dcontext/"
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: TestNamePattern{"/2dcontext/"}}, rq)
}

func TestStructuredQuery_subtest(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"subtest": "Subtest name"
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: SubtestNamePattern{"Subtest name"}}, rq)
}

func TestStructuredQuery_path(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"path": "/2dcontext/"
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: TestPath{"/2dcontext/"}}, rq)
}

func TestStructuredQuery_legacyBrowserName(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"browser_name": "FiReFoX",
			"status": "PaSs"
		}
	}`), &rq)
	assert.Nil(t, err)
	p := shared.ParseProductSpecUnsafe("firefox")
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: TestStatusEq{&p, shared.TestStatusValueFromString("PASS")},
	}, rq)
}

func TestStructuredQuery_status(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"product": "FiReFoX",
			"status": "PaSs"
		}
	}`), &rq)
	assert.Nil(t, err)
	p := shared.ParseProductSpecUnsafe("firefox")
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: TestStatusEq{&p, shared.TestStatusValueFromString("PASS")},
	}, rq)
}

func TestStructuredQuery_statusNeq(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"product": "FiReFoX",
			"status": {"not": "PaSs"}
		}
	}`), &rq)
	assert.Nil(t, err)
	p := shared.ParseProductSpecUnsafe("firefox")
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: TestStatusNeq{&p, shared.TestStatusValueFromString("PASS")},
	}, rq)
}

func TestStructuredQuery_statusUnsupportedAbstractNot(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"product": "FiReFoX",
			"status": {"not": {"pattern": "cssom"}}
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_not(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"not": {
				"pattern": "cssom"
			}
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: AbstractNot{TestNamePattern{"cssom"}}}, rq)
}

func TestStructuredQuery_or(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"or": [
				{"pattern": "cssom"},
				{"pattern": "html"}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: AbstractOr{[]AbstractQuery{TestNamePattern{"cssom"}, TestNamePattern{"html"}}}}, rq)
}

func TestStructuredQuery_and(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"and": [
				{"pattern": "cssom"},
				{"pattern": "html"}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: AbstractAnd{[]AbstractQuery{TestNamePattern{"cssom"}, TestNamePattern{"html"}}}}, rq)
}

func TestStructuredQuery_exists(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [
				{"pattern": "cssom"},
				{"pattern": "html"}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: AbstractExists{[]AbstractQuery{TestNamePattern{"cssom"}, TestNamePattern{"html"}}}}, rq)
}

func TestStructuredQuery_all(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"all": [
				{"pattern": "cssom"}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{
		RunIDs:        []int64{0, 1, 2},
		AbstractQuery: AbstractAll{[]AbstractQuery{TestNamePattern{"cssom"}}},
	}, rq)
}

func TestStructuredQuery_none(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"none": [
				{"pattern": "cssom"}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{
		RunIDs:        []int64{0, 1, 2},
		AbstractQuery: AbstractNone{[]AbstractQuery{TestNamePattern{"cssom"}}},
	}, rq)
}

func TestStructuredQuery_sequential(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [
				{ "sequential":[
					{"or":[{"status":"PASS"},{"status":"OK"}]},
					{"and":[{"status":{"not":"PASS"}},{"status":{"not":"OK"}},{"status":{"not":"UNKNOWN"}}]}
				]}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(
		t,
		RunQuery{RunIDs: []int64{0, 1, 2},
			AbstractQuery: AbstractExists{[]AbstractQuery{
				AbstractSequential{[]AbstractQuery{
					AbstractOr{[]AbstractQuery{
						TestStatusEq{Status: shared.TestStatusValueFromString("PASS")},
						TestStatusEq{Status: shared.TestStatusValueFromString("OK")},
					}},
					AbstractAnd{[]AbstractQuery{
						TestStatusNeq{Status: shared.TestStatusValueFromString("PASS")},
						TestStatusNeq{Status: shared.TestStatusValueFromString("OK")},
						TestStatusNeq{Status: shared.TestStatusValueFromString("UNKNOWN")},
					}},
				}},
			}}}, rq)
}

func TestStructuredQuery_count(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [{
				"count": 3,
				"where": {
					"or": [{"status":"PASS"},{"status":"OK"}]
				}
			}]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(
		t,
		RunQuery{RunIDs: []int64{0, 1, 2},
			AbstractQuery: AbstractExists{[]AbstractQuery{
				AbstractCount{
					Count: 3,
					Where: AbstractOr{[]AbstractQuery{
						TestStatusEq{Status: shared.TestStatusValueFromString("PASS")},
						TestStatusEq{Status: shared.TestStatusValueFromString("OK")},
					}},
				}},
			}}, rq)
}

func TestStructuredQuery_moreThan(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [{
				"moreThan": 3,
				"where": {"status":"PASS"}
			}]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(
		t,
		RunQuery{RunIDs: []int64{0, 1, 2},
			AbstractQuery: AbstractExists{[]AbstractQuery{
				AbstractMoreThan{
					AbstractCount{
						Count: 3,
						Where: TestStatusEq{Status: shared.TestStatusValueFromString("PASS")},
					},
				}},
			}}, rq)
}

func TestStructuredQuery_lessThan(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [{
				"lessThan": 2,
				"where": {"status":"PASS"}
			}]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(
		t,
		RunQuery{RunIDs: []int64{0, 1, 2},
			AbstractQuery: AbstractExists{[]AbstractQuery{
				AbstractLessThan{
					AbstractCount{
						Count: 2,
						Where: TestStatusEq{Status: shared.TestStatusValueFromString("PASS")},
					},
				}},
			}}, rq)
}

func TestStructuredQuery_link(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [{
				"link": "chromium.bug.com/abc"
			}]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{[]AbstractQuery{
			AbstractLink{
				Pattern: "chromium.bug.com/abc",
			}},
		}}, rq)
}

func TestStructuredQuery_triaged(t *testing.T) {
	p := shared.ParseProductSpecUnsafe("Chrome")
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [{
				"triaged": "chrome"
			}]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{[]AbstractQuery{
			AbstractTriaged{
				Product: &p,
			}},
		}}, rq)
}

func TestStructuredQuery_triagedEmptyProduct(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [{
				"triaged": ""
			}]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{[]AbstractQuery{
			AbstractTriaged{
				Product: nil,
			}},
		}}, rq)
}

func TestStructuredQuery_testlabel(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
        "run_ids": [0, 1, 2],
        "query": {
            "exists": [{
                "label": "interop1"
            }]
        }
    }`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{[]AbstractQuery{
			AbstractTestLabel{
				Label: "interop1",
			}},
		}}, rq)
}

func TestStructuredQuery_combinedTestlabel(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
        "run_ids": [0, 1, 2],
        "query": {
            "exists": [
                {"pattern": "cssom"},
                {"label": "interop"}
            ]
        }
    }`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{[]AbstractQuery{TestNamePattern{"cssom"}, AbstractTestLabel{Label: "interop"}}}}, rq)
}

func TestStructuredQuery_andTestLabels(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"and": [
				{"label": "interop1"},
				{"label": "interop2"}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: AbstractAnd{[]AbstractQuery{AbstractTestLabel{Label: "interop1"}, AbstractTestLabel{Label: "interop2"}}}}, rq)
}

func TestStructuredQuery_testfeature(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
        "run_ids": [0, 1, 2],
        "query": {
            "exists": [{
                "feature": "feature1"
            }]
        }
    }`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{[]AbstractQuery{
			AbstractTestWebFeature{
				TestWebFeatureAtom: TestWebFeatureAtom{
					WebFeature: "feature1",
				},
				manifestFetcher: searchcacheWebFeaturesManifestFetcher{},
			}},
		}}, rq)
}

func TestStructuredQuery_andTestFeatures(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"and": [
				{"feature": "feature1"},
				{"feature": "feature2"}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t,
		RunQuery{
			RunIDs: []int64{0, 1, 2},
			AbstractQuery: AbstractAnd{
				[]AbstractQuery{
					AbstractTestWebFeature{
						TestWebFeatureAtom: TestWebFeatureAtom{
							WebFeature: "feature1",
						},
						manifestFetcher: searchcacheWebFeaturesManifestFetcher{},
					},
					AbstractTestWebFeature{
						TestWebFeatureAtom: TestWebFeatureAtom{
							WebFeature: "feature2",
						},
						manifestFetcher: searchcacheWebFeaturesManifestFetcher{},
					},
				}}},
		rq)
}

func TestStructuredQuery_isDifferent(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [{
				"is": "different"
			}]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{
		RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{[]AbstractQuery{
			MetadataQualityDifferent,
		}},
	}, rq)
}

func TestStructuredQuery_isTentative(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [{
				"is": "tentative"
			}]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{
		RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{[]AbstractQuery{
			MetadataQualityTentative,
		}},
	}, rq)
}

func TestStructuredQuery_isOptional(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [{
				"is": "optional"
			}]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{
		RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{[]AbstractQuery{
			MetadataQualityOptional,
		}},
	}, rq)
}

func TestStructuredQuery_combinedlink(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [
				{"pattern": "cssom"},
				{"link": "chromium"}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{[]AbstractQuery{TestNamePattern{"cssom"}, AbstractLink{Pattern: "chromium"}}}}, rq)
}

func TestStructuredQuery_combinednotlink(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [
				{"and": [
					{"pattern": "cssom"},
					{"not": {"link": "chromium.bug"}
					}
				  ]
				}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{Args: []AbstractQuery{AbstractAnd{Args: []AbstractQuery{TestNamePattern{Pattern: "cssom"}, AbstractNot{Arg: AbstractLink{Pattern: "chromium.bug"}}}}}}}, rq)
}

func TestStructuredQuery_existsSimple(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [
				{"and": [
					{"pattern": "cssom"}
				  ]
				}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractExists{Args: []AbstractQuery{AbstractAnd{Args: []AbstractQuery{TestNamePattern{Pattern: "cssom"}}}}}}, rq)
}

func TestStructuredQuery_existsWithAnd(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"exists": [
				{"pattern": "cssom"}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2}, AbstractQuery: AbstractExists{[]AbstractQuery{TestNamePattern{"cssom"}}}}, rq)
}

func TestStructuredQuery_nested(t *testing.T) {
	var rq RunQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"or": [
				{
					"and": [
						{"not": {"pattern": "cssom"}},
						{"pattern": "html"}
					]
				},
				{
					"product": "eDgE",
					"status": "tImEoUt"
				}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	p := shared.ParseProductSpecUnsafe("edge")
	assert.Equal(t, RunQuery{
		RunIDs: []int64{0, 1, 2},
		AbstractQuery: AbstractOr{
			Args: []AbstractQuery{
				AbstractAnd{
					Args: []AbstractQuery{
						AbstractNot{TestNamePattern{"cssom"}},
						TestNamePattern{"html"},
					},
				},
				TestStatusEq{&p, shared.TestStatusValueFromString("TIMEOUT")},
			},
		},
	}, rq)
}

func TestStructuredQuery_bindPattern(t *testing.T) {
	tnp := TestNamePattern{
		Pattern: "/",
	}
	q := tnp.BindToRuns()
	assert.Equal(t, tnp, q)
}

func TestStructuredQuery_bindBrowserStatusNoRuns(t *testing.T) {
	p := shared.ParseProductSpecUnsafe("Chrome")
	assert.Equal(t, False{}, TestStatusEq{
		Product: &p,
		Status:  1,
	}.BindToRuns())
}

func TestStructuredQuery_bindBrowserStatusSingleRun(t *testing.T) {
	p := shared.ParseProductSpecUnsafe("firefox")
	q := TestStatusEq{
		Product: &p,
		Status:  1,
	}
	runs := []shared.TestRun{
		{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Firefox").ProductAtRevision,
		},
		{
			ID:                2,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Chrome").ProductAtRevision,
		},
	}
	// Only Firefox run ID=1.
	expected := RunTestStatusEq{
		Run:    1,
		Status: 1,
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindBrowserStatusSingleRunNeq(t *testing.T) {
	p := shared.ParseProductSpecUnsafe("firefox")
	q := TestStatusNeq{
		Product: &p,
		Status:  1,
	}
	runs := []shared.TestRun{
		{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Firefox").ProductAtRevision,
		},
		{
			ID:                2,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Chrome").ProductAtRevision,
		},
	}
	// Only Firefox run ID=1.
	expected := RunTestStatusNeq{
		Run:    1,
		Status: 1,
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindStatusSomeRuns(t *testing.T) {
	q := TestStatusNeq{
		Status: 1,
	}
	runs := shared.TestRuns{
		shared.TestRun{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Firefox").ProductAtRevision,
		},
		shared.TestRun{
			ID:                2,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Chrome").ProductAtRevision,
		},
	}
	// Two Firefox runs: ID=1, ID=3.
	expected := Or{
		Args: []ConcreteQuery{
			RunTestStatusNeq{
				Run:    1,
				Status: 1,
			},
			RunTestStatusNeq{
				Run:    2,
				Status: 1,
			},
		},
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindBrowserStatusSomeRuns(t *testing.T) {
	p := shared.ParseProductSpecUnsafe("firefox")
	q := TestStatusNeq{
		Product: &p,
		Status:  1,
	}
	runs := []shared.TestRun{
		{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Firefox").ProductAtRevision,
		},
		{
			ID:                2,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Chrome").ProductAtRevision,
		},
		{
			ID:                3,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Firefox").ProductAtRevision,
		},
	}
	// Two Firefox runs: ID=1, ID=3.
	expected := Or{
		Args: []ConcreteQuery{
			RunTestStatusNeq{
				Run:    1,
				Status: 1,
			},
			RunTestStatusNeq{
				Run:    3,
				Status: 1,
			},
		},
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindExists(t *testing.T) {
	e := shared.ParseProductSpecUnsafe("edge")
	f := shared.ParseProductSpecUnsafe("firefox")
	q := AbstractExists{
		Args: []AbstractQuery{
			AbstractAnd{
				Args: []AbstractQuery{
					TestNamePattern{
						Pattern: "/",
					},
					TestStatusEq{
						Product: &e,
						Status:  1,
					},
				},
			},
			TestStatusNeq{
				Product: &f,
				Status:  1,
			},
		},
	}

	runs := shared.TestRuns{}
	or1 := Or{}
	or2 := Or{}
	products := []shared.ProductSpec{e, f}
	for i := 1; i <= 10; i++ {
		runs = append(
			runs,
			shared.TestRun{
				ID:                int64(i),
				ProductAtRevision: products[i%2].ProductAtRevision,
			})
		if i%2 == 0 { // Evens are edge
			or1.Args = append(or1.Args,
				And{
					Args: []ConcreteQuery{
						TestNamePattern{
							Pattern: "/",
						},
						RunTestStatusEq{
							Run:    int64(i),
							Status: 1,
						},
					},
				})
		} else { // Odds are firefox
			or2.Args = append(or2.Args,
				RunTestStatusNeq{
					Run:    int64(i),
					Status: 1,
				},
			)
		}
	}
	expected := And{
		Args: []ConcreteQuery{or1, or2},
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindExistsWithTwoProducts(t *testing.T) {
	e := shared.ParseProductSpecUnsafe("edge")
	f := shared.ParseProductSpecUnsafe("firefox")
	q := AbstractExists{
		Args: []AbstractQuery{
			AbstractAnd{
				Args: []AbstractQuery{
					TestStatusEq{
						Product: &e,
						Status:  1,
					},
					TestStatusEq{
						Product: &f,
						Status:  1,
					},
				},
			},
		},
	}

	runs := shared.TestRuns{}
	or1 := Or{}
	or2 := Or{}
	products := []shared.ProductSpec{e, f}
	for i := 1; i <= 10; i++ {
		runs = append(
			runs,
			shared.TestRun{
				ID:                int64(i),
				ProductAtRevision: products[i%2].ProductAtRevision,
			})
		if i%2 == 0 { // Evens are edge
			or1.Args = append(or1.Args,
				RunTestStatusEq{
					Run:    int64(i),
					Status: 1,
				},
			)
		} else { // Odds are firefox
			or2.Args = append(or2.Args,
				RunTestStatusEq{
					Run:    int64(i),
					Status: 1,
				},
			)
		}
	}
	expected := And{
		Args: []ConcreteQuery{or1, or2},
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindSequential(t *testing.T) {
	e := shared.ParseProductSpecUnsafe("edge")
	f := shared.ParseProductSpecUnsafe("firefox")
	q := AbstractSequential{
		Args: []AbstractQuery{
			TestStatusEq{Product: &e, Status: 1},
			TestStatusNeq{Product: &f, Status: 1},
		},
	}

	runs := shared.TestRuns{}
	runs = shared.TestRuns{
		{
			ID:                int64(0),
			ProductAtRevision: e.ProductAtRevision,
		},
		{
			ID:                int64(1),
			ProductAtRevision: f.ProductAtRevision,
		},
	}
	seq := And{
		Args: []ConcreteQuery{
			RunTestStatusEq{Run: int64(0), Status: 1},
			RunTestStatusNeq{Run: int64(1), Status: 1},
		},
	}
	expected := Or{
		Args: []ConcreteQuery{seq},
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindCount(t *testing.T) {
	e := shared.ParseProductSpecUnsafe("edge")
	f := shared.ParseProductSpecUnsafe("firefox")
	q := AbstractCount{
		Count: 1,
		Where: TestStatusEq{Status: 1},
	}

	runs := shared.TestRuns{
		{
			ID:                int64(0),
			ProductAtRevision: e.ProductAtRevision,
		},
		{
			ID:                int64(1),
			ProductAtRevision: f.ProductAtRevision,
		},
	}
	expected := Count{
		Count: 1,
		Args: []ConcreteQuery{
			RunTestStatusEq{Run: int64(0), Status: 1},
			RunTestStatusEq{Run: int64(1), Status: 1},
		},
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindLink(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := "sha"
	mockFetcher := sharedtest.NewMockMetadataFetcher(mockCtrl)
	mockFetcher.EXPECT().Fetch().Return(&sha, getMetadataTestData(), nil)

	e := shared.ParseProductSpecUnsafe("safari")
	f := shared.ParseProductSpecUnsafe("firefox")
	q := AbstractLink{
		Pattern:         "bar",
		metadataFetcher: mockFetcher,
	}

	runs := shared.TestRuns{
		{
			ID:                int64(0),
			ProductAtRevision: e.ProductAtRevision,
		},
		{
			ID:                int64(1),
			ProductAtRevision: f.ProductAtRevision,
		},
	}

	// AbstractLink should bind test-level issues too as the pattern might match
	// them. It should not include the Chromium link however, as there is no run
	// for Chromium and thus no reason to include it - the frontend won't show it.
	expect := Link{
		Pattern: "bar",
		Metadata: map[string][]string{
			"/testB/b.html": {"bar.com"},
			"/testC/c.html": {"baz.com"},
		},
	}
	assert.Equal(t, expect, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindTriaged(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := "sha"
	mockFetcher := sharedtest.NewMockMetadataFetcher(mockCtrl)
	mockFetcher.EXPECT().Fetch().Return(&sha, getMetadataTestData(), nil).AnyTimes()

	safari := shared.ParseProductSpecUnsafe("safari")
	firefox := shared.ParseProductSpecUnsafe("firefox")
	q := AbstractTriaged{
		Product:         &firefox,
		metadataFetcher: mockFetcher,
	}

	runs := shared.TestRuns{
		{
			ID:                int64(0),
			ProductAtRevision: safari.ProductAtRevision,
		},
		{
			ID:                int64(1),
			ProductAtRevision: firefox.ProductAtRevision,
		},
	}

	expect := Or{
		Args: []ConcreteQuery{
			Triaged{
				Run: 1,
				Metadata: map[string][]string{
					"/testB/b.html": {"bar.com"},
				},
			},
		},
	}
	assert.Equal(t, expect, q.BindToRuns(runs...))

	// This query doesn't match any of the runs, so should convert to False.
	q = AbstractTriaged{
		Product:         &safari,
		metadataFetcher: mockFetcher,
	}
	assert.Equal(t, False{}, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindTriagedNilProduct(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := "sha"
	mockFetcher := sharedtest.NewMockMetadataFetcher(mockCtrl)
	mockFetcher.EXPECT().Fetch().Return(&sha, getMetadataTestData(), nil).AnyTimes()

	q := AbstractTriaged{
		Product:         nil,
		metadataFetcher: mockFetcher,
	}

	safari := shared.ParseProductSpecUnsafe("safari")
	firefox := shared.ParseProductSpecUnsafe("firefox")
	runs := shared.TestRuns{
		{
			ID:                int64(0),
			ProductAtRevision: safari.ProductAtRevision,
		},
		{
			ID:                int64(1),
			ProductAtRevision: firefox.ProductAtRevision,
		},
	}

	// This is inefficient, but currently a nil product binds to all runs, with
	// the same metadata in all cases.
	expect := Or{
		Args: []ConcreteQuery{
			Triaged{
				Run: 0,
				Metadata: map[string][]string{
					"/testC/c.html": {"baz.com"},
				},
			},
			Triaged{
				Run: 1,
				Metadata: map[string][]string{
					"/testC/c.html": {"baz.com"},
				},
			},
		},
	}
	assert.Equal(t, expect, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindTestLabel(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := "sha"
	mockFetcher := sharedtest.NewMockMetadataFetcher(mockCtrl)
	mockFetcher.EXPECT().Fetch().Return(&sha, getMetadataTestData(), nil).AnyTimes()

	safari := shared.ParseProductSpecUnsafe("safari")
	firefox := shared.ParseProductSpecUnsafe("firefox")
	q := AbstractTestLabel{
		Label:           "interop",
		metadataFetcher: mockFetcher,
	}

	runs := shared.TestRuns{
		{
			ID:                int64(0),
			ProductAtRevision: safari.ProductAtRevision,
		},
		{
			ID:                int64(1),
			ProductAtRevision: firefox.ProductAtRevision,
		},
	}

	expect := TestLabel{
		Label: "interop",
		Metadata: map[string][]string{
			"/testC/c.html": {"labelA"},
		},
	}
	assert.Equal(t, expect, q.BindToRuns(runs...))
}

type testWebFeaturesManifestFetcher struct {
	data shared.WebFeaturesData
	err  error
}

func (t testWebFeaturesManifestFetcher) Fetch() (shared.WebFeaturesData, error) {
	return t.data, t.err
}

func TestStructuredQuery_bindTestWebFeature(t *testing.T) {
	mockManifestFetcher := testWebFeaturesManifestFetcher{
		data: shared.WebFeaturesData{
			"grid": {"/css/css-grid/bar.html": nil},
			"avif": {"/avif/foo.html": nil},
		},
		err: nil,
	}

	safari := shared.ParseProductSpecUnsafe("safari")
	firefox := shared.ParseProductSpecUnsafe("firefox")
	q := AbstractTestWebFeature{
		TestWebFeatureAtom: TestWebFeatureAtom{
			WebFeature: "grid",
		},
		manifestFetcher: mockManifestFetcher,
	}

	runs := shared.TestRuns{
		{
			ID:                int64(0),
			ProductAtRevision: safari.ProductAtRevision,
		},
		{
			ID:                int64(1),
			ProductAtRevision: firefox.ProductAtRevision,
		},
	}

	expect := TestWebFeature{
		WebFeature: "grid",
		WebFeaturesData: shared.WebFeaturesData{
			"grid": {"/css/css-grid/bar.html": nil},
			"avif": {"/avif/foo.html": nil},
		},
	}
	assert.Equal(t, expect, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindIs(t *testing.T) {
	e := shared.ParseProductSpecUnsafe("chrome")
	f := shared.ParseProductSpecUnsafe("safari")
	q := MetadataQualityDifferent

	runs := shared.TestRuns{
		{
			ID:                int64(0),
			ProductAtRevision: e.ProductAtRevision,
		},
		{
			ID:                int64(1),
			ProductAtRevision: f.ProductAtRevision,
		},
	}
	// BindToRuns for MetadataQuality is a no-op, as they are independent
	// of runs.
	assert.Equal(t, q, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindAnd(t *testing.T) {
	p := shared.ParseProductSpecUnsafe("edge")
	q := AbstractAnd{
		Args: []AbstractQuery{
			TestNamePattern{
				Pattern: "/",
			},
			TestStatusEq{
				Product: &p,
				Status:  1,
			},
		},
	}
	runs := []shared.TestRun{
		{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Edge").ProductAtRevision,
		},
	}
	// Only run is Edge, ID=1.
	expected := And{
		Args: []ConcreteQuery{
			TestNamePattern{
				Pattern: "/",
			},
			RunTestStatusEq{
				Run:    1,
				Status: 1,
			},
		},
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindOr(t *testing.T) {
	p := shared.ParseProductSpecUnsafe("edge")
	q := AbstractOr{
		Args: []AbstractQuery{
			TestNamePattern{
				Pattern: "/",
			},
			TestStatusEq{
				Product: &p,
				Status:  1,
			},
		},
	}
	runs := []shared.TestRun{
		{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Edge").ProductAtRevision,
		},
	}
	// Only run is Edge, ID=1.
	expected := Or{
		Args: []ConcreteQuery{
			TestNamePattern{
				Pattern: "/",
			},
			RunTestStatusEq{
				Run:    1,
				Status: 1,
			},
		},
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindNot(t *testing.T) {
	p := shared.ParseProductSpecUnsafe("edge")
	q := AbstractNot{
		Arg: TestStatusEq{
			Product: &p,
			Status:  1,
		},
	}
	runs := []shared.TestRun{
		{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Edge").ProductAtRevision,
		},
	}
	// Only run is Edge, ID=1.
	expected := Not{
		Arg: RunTestStatusEq{
			Run:    1,
			Status: 1,
		},
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindAndReduce(t *testing.T) {
	p := shared.ParseProductSpecUnsafe("safari")
	q := AbstractAnd{
		Args: []AbstractQuery{
			TestNamePattern{
				Pattern: "/",
			},
			TestStatusEq{
				Product: &p,
				Status:  1,
			},
		},
	}
	runs := []shared.TestRun{
		{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Edge").ProductAtRevision,
		},
	}
	// No runs match Safari constraint; it becomes False,
	// False && Pattern="/" => False.
	assert.Equal(t, False{}, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindAndReduceToTrue(t *testing.T) {
	s := shared.ParseProductSpecUnsafe("safari")
	c := shared.ParseProductSpecUnsafe("chrome")
	q := AbstractAnd{
		Args: []AbstractQuery{
			TestStatusEq{
				Product: &c,
				Status:  1,
			},
			TestStatusNeq{
				Product: &s,
				Status:  1,
			},
		},
	}
	runs := []shared.TestRun{
		{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Edge").ProductAtRevision,
		},
	}
	// No runs match any constraint; reduce to False.
	assert.Equal(t, False{}, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindOrReduce(t *testing.T) {
	p := shared.ParseProductSpecUnsafe("safari")
	q := AbstractOr{
		Args: []AbstractQuery{
			TestNamePattern{
				Pattern: "/",
			},
			TestStatusEq{
				Product: &p,
				Status:  1,
			},
		},
	}
	runs := []shared.TestRun{
		{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Edge").ProductAtRevision,
		},
	}
	// No runs match Safari constraint; it becomes False,
	// Pattern="/" || False => Pattern.
	expected := TestNamePattern{"/"}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindComplex(t *testing.T) {
	s := shared.ParseProductSpecUnsafe("safari")
	c := shared.ParseProductSpecUnsafe("chrome")
	q := AbstractOr{
		Args: []AbstractQuery{
			TestNamePattern{
				Pattern: "cssom",
			},
			AbstractAnd{
				Args: []AbstractQuery{
					AbstractNot{
						Arg: TestNamePattern{
							Pattern: "css",
						},
					},
					TestStatusEq{
						Product: &s,
						Status:  1,
					},
					TestStatusNeq{
						Product: &c,
						Status:  1,
					},
				},
			},
		},
	}
	runs := []shared.TestRun{
		{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Chrome").ProductAtRevision,
		},
		{
			ID:                2,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Edge").ProductAtRevision,
		},
		{
			ID:                3,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Chrome").ProductAtRevision,
		},
	}
	// No runs match Safari constraint, so False; two Chrome runs expand to disjunction over
	// their values, but are combined with false in an AND, so False; leaving only
	// Pattern="cssom"
	expected := TestNamePattern{
		Pattern: "cssom",
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func getMetadataTestData() map[string][]byte {
	metadataMap := make(map[string][]byte)
	metadataMap["root/testA"] = []byte(`
    links:
      - product: chrome
        url: foo.com
        results:
        - test: a.html
          status: FAIL
    `)

	metadataMap["testB"] = []byte(`
    links:
      - product: firefox
        url: bar.com
        results:
        - test: b.html
          status: FAIL
    `)

	// A test-level issue, which has no product associated with it.
	metadataMap["testC"] = []byte(`
    links:
      - label: labelA
        url: baz.com
        results:
        - test: c.html
          status: FAIL
    `)

	return metadataMap
}
