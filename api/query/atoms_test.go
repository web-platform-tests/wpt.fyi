// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
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
	assert.NotNil(t, err)
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
			"product": "",
			"status": "PASS"
		}
	}`), &rq)
	assert.NotNil(t, err)
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
	assert.Equal(t, RunQuery{
		RunIDs:        []int64{0, 1, 2},
		AbstractQuery: TestStatusEq{shared.ParseProductSpecUnsafe("chrome"), shared.TestStatusValueFromString("UNKNOWN")},
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
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: TestStatusEq{shared.ParseProductSpecUnsafe("firefox"), shared.TestStatusValueFromString("PASS")},
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
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: TestStatusEq{shared.ParseProductSpecUnsafe("firefox"), shared.TestStatusValueFromString("PASS")},
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
	assert.Equal(t, RunQuery{RunIDs: []int64{0, 1, 2},
		AbstractQuery: TestStatusNeq{shared.ParseProductSpecUnsafe("firefox"), shared.TestStatusValueFromString("PASS")},
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
				TestStatusEq{shared.ParseProductSpecUnsafe("edge"), shared.TestStatusValueFromString("TIMEOUT")},
			},
		},
	}, rq)
}

func TestStructuredQuery_bindPattern(t *testing.T) {
	tnp := TestNamePattern{
		Pattern: "/",
	}
	q := tnp.BindToRuns([]shared.TestRun{})
	assert.Equal(t, tnp, q)
}

func TestStructuredQuery_bindStatusNoRuns(t *testing.T) {
	assert.Equal(t, True{}, TestStatusEq{
		Product: shared.ParseProductSpecUnsafe("Chrome"),
		Status:  1,
	}.BindToRuns([]shared.TestRun{}))
}

func TestStructuredQuery_bindStatusSingleRun(t *testing.T) {
	q := TestStatusEq{
		Product: shared.ParseProductSpecUnsafe("Firefox"),
		Status:  1,
	}
	runs := []shared.TestRun{
		shared.TestRun{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Firefox").ProductAtRevision,
		},
		shared.TestRun{
			ID:                2,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Chrome").ProductAtRevision,
		},
	}
	// Only Firefox run ID=1.
	expected := RunTestStatusEq{
		Run:    1,
		Status: 1,
	}
	assert.Equal(t, expected, q.BindToRuns(runs))
}

func TestStructuredQuery_bindStatusSingleRunNeq(t *testing.T) {
	q := TestStatusNeq{
		Product: shared.ParseProductSpecUnsafe("Firefox"),
		Status:  1,
	}
	runs := []shared.TestRun{
		shared.TestRun{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Firefox").ProductAtRevision,
		},
		shared.TestRun{
			ID:                2,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Chrome").ProductAtRevision,
		},
	}
	// Only Firefox run ID=1.
	expected := RunTestStatusNeq{
		Run:    1,
		Status: 1,
	}
	assert.Equal(t, expected, q.BindToRuns(runs))
}

func TestStructuredQuery_bindStatusSomeRuns(t *testing.T) {
	q := TestStatusNeq{
		Product: shared.ParseProductSpecUnsafe("Firefox"),
		Status:  1,
	}
	runs := []shared.TestRun{
		shared.TestRun{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Firefox").ProductAtRevision,
		},
		shared.TestRun{
			ID:                2,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Chrome").ProductAtRevision,
		},
		shared.TestRun{
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
	assert.Equal(t, expected, q.BindToRuns(runs))
}

func TestStructuredQuery_bindAnd(t *testing.T) {
	q := AbstractAnd{
		Args: []AbstractQuery{
			TestNamePattern{
				Pattern: "/",
			},
			TestStatusEq{
				Product: shared.ParseProductSpecUnsafe("Edge"),
				Status:  1,
			},
		},
	}
	runs := []shared.TestRun{
		shared.TestRun{
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
	assert.Equal(t, expected, q.BindToRuns(runs))
}

func TestStructuredQuery_bindOr(t *testing.T) {
	q := AbstractOr{
		Args: []AbstractQuery{
			TestNamePattern{
				Pattern: "/",
			},
			TestStatusEq{
				Product: shared.ParseProductSpecUnsafe("Edge"),
				Status:  1,
			},
		},
	}
	runs := []shared.TestRun{
		shared.TestRun{
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
	assert.Equal(t, expected, q.BindToRuns(runs))
}

func TestStructuredQuery_bindNot(t *testing.T) {
	q := AbstractNot{
		Arg: TestStatusEq{
			Product: shared.ParseProductSpecUnsafe("Edge"),
			Status:  1,
		},
	}
	runs := []shared.TestRun{
		shared.TestRun{
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
	assert.Equal(t, expected, q.BindToRuns(runs))
}

func TestStructuredQuery_bindAndReduce(t *testing.T) {
	q := AbstractAnd{
		Args: []AbstractQuery{
			TestNamePattern{
				Pattern: "/",
			},
			TestStatusEq{
				Product: shared.ParseProductSpecUnsafe("Safari"),
				Status:  1,
			},
		},
	}
	runs := []shared.TestRun{
		shared.TestRun{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Edge").ProductAtRevision,
		},
	}
	// No runs match Safari constraint; it becomes True,
	// True && Pattern="/" => Pattern="/".
	expected := TestNamePattern{
		Pattern: "/",
	}
	assert.Equal(t, expected, q.BindToRuns(runs))
}

func TestStructuredQuery_bindAndReduceToTrue(t *testing.T) {
	q := AbstractAnd{
		Args: []AbstractQuery{
			TestStatusEq{
				Product: shared.ParseProductSpecUnsafe("Chrome"),
				Status:  1,
			},
			TestStatusNeq{
				Product: shared.ParseProductSpecUnsafe("Safari"),
				Status:  1,
			},
		},
	}
	runs := []shared.TestRun{
		shared.TestRun{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Edge").ProductAtRevision,
		},
	}
	// No runs match any constraint; reduce to True.
	expected := True{}
	assert.Equal(t, expected, q.BindToRuns(runs))
}

func TestStructuredQuery_bindOrReduce(t *testing.T) {
	q := AbstractOr{
		Args: []AbstractQuery{
			TestNamePattern{
				Pattern: "/",
			},
			TestStatusEq{
				Product: shared.ParseProductSpecUnsafe("Safari"),
				Status:  1,
			},
		},
	}
	runs := []shared.TestRun{
		shared.TestRun{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Edge").ProductAtRevision,
		},
	}
	// No runs match Safari constraint; it becomes True,
	// Pattern="/" || True => True.
	expected := True{}
	assert.Equal(t, expected, q.BindToRuns(runs))
}

func TestStructuredQuery_bindComplex(t *testing.T) {
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
						Product: shared.ParseProductSpecUnsafe("Safari"),
						Status:  1,
					},
					TestStatusNeq{
						Product: shared.ParseProductSpecUnsafe("Chrome"),
						Status:  1,
					},
				},
			},
		},
	}
	runs := []shared.TestRun{
		shared.TestRun{
			ID:                1,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Chrome").ProductAtRevision,
		},
		shared.TestRun{
			ID:                2,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Edge").ProductAtRevision,
		},
		shared.TestRun{
			ID:                3,
			ProductAtRevision: shared.ParseProductSpecUnsafe("Chrome").ProductAtRevision,
		},
	}
	// No runs match Safari constraint; two Chrome runs expand to disjunction over
	// their values:
	// Pattern="cssom" || (!Pattern="css" && Safari(status=1) && Chrome(status=1))
	// => Pattern="cssom" || (!Pattern="css" && (RunID=1(status=1) ||
	//                                           RunID=3(status=1))
	expected := Or{
		Args: []ConcreteQuery{
			TestNamePattern{
				Pattern: "cssom",
			},
			And{
				Args: []ConcreteQuery{
					Not{
						Arg: TestNamePattern{
							Pattern: "css",
						},
					},
					Or{
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
					},
				},
			},
		},
	}
	assert.Equal(t, expected, q.BindToRuns(runs))
}
