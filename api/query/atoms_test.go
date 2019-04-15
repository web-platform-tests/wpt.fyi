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
	assert.Equal(t, expected, q.BindToRuns(runs...))
}

func TestStructuredQuery_bindBrowserStatusSingleRunNeq(t *testing.T) {
	p := shared.ParseProductSpecUnsafe("firefox")
	q := TestStatusNeq{
		Product: &p,
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
	expected := Count{
		Count: 1,
		Args: []ConcreteQuery{
			RunTestStatusEq{Run: int64(0), Status: 1},
			RunTestStatusEq{Run: int64(1), Status: 1},
		},
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
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
		shared.TestRun{
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
		shared.TestRun{
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
		shared.TestRun{
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
	// No runs match Safari constraint, so False; two Chrome runs expand to disjunction over
	// their values, but are combined with false in an AND, so False; leaving only
	// Pattern="cssom"
	expected := TestNamePattern{
		Pattern: "cssom",
	}
	assert.Equal(t, expected, q.BindToRuns(runs...))
}
