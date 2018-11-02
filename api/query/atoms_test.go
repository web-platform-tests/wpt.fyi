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
	var rq runQuery
	err := json.Unmarshal([]byte(`{}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_missingRunIDs(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"query": {
			"pattern": "/2dcontext/"
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_missingQuery(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2]
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_emptyRunIDs(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [],
		"query": {
			"pattern": "/2dcontext/"
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_emptyBrowserName(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"browser_name": "",
			"status": "PASS"
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_missingStatus(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"browser_name": "chrome"
		}
	}`), &rq)
	assert.NotNil(t, err)
}

func TestStructuredQuery_badStatus(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"browser_name": "chrome",
			"status": "NOT_A_REAL_STATUS"
		}
	}`), &rq)
	assert.NotNil(t, err)
}
func TestStructuredQuery_unknownStatus(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"browser_name": "chrome",
			"status": "UNKNOWN"
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: testStatusConstraint{"chrome", shared.TestStatusValueFromString("UNKNOWN")}}, rq)
}

func TestStructuredQuery_pattern(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"pattern": "/2dcontext/"
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: testNamePattern{"/2dcontext/"}}, rq)
}

func TestStructuredQuery_status(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"browser_name": "FiReFoX",
			"status": "PaSs"
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: testStatusConstraint{"firefox", shared.TestStatusValueFromString("PASS")}}, rq)
}

func TestStructuredQuery_not(t *testing.T) {
	var rq runQuery
	err := json.Unmarshal([]byte(`{
		"run_ids": [0, 1, 2],
		"query": {
			"not": {
				"pattern": "cssom"
			}
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: not{testNamePattern{"cssom"}}}, rq)
}

func TestStructuredQuery_or(t *testing.T) {
	var rq runQuery
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
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: or{[]query{testNamePattern{"cssom"}, testNamePattern{"html"}}}}, rq)
}

func TestStructuredQuery_and(t *testing.T) {
	var rq runQuery
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
	assert.Equal(t, runQuery{runIDs: []int64{0, 1, 2}, query: and{[]query{testNamePattern{"cssom"}, testNamePattern{"html"}}}}, rq)
}

func TestStructuredQuery_nested(t *testing.T) {
	var rq runQuery
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
					"browser_name": "eDgE",
					"status": "tImEoUt"
				}
			]
		}
	}`), &rq)
	assert.Nil(t, err)
	assert.Equal(t, runQuery{
		runIDs: []int64{0, 1, 2},
		query: or{
			or: []query{
				and{
					and: []query{
						not{testNamePattern{"cssom"}},
						testNamePattern{"html"},
					},
				},
				testStatusConstraint{"edge", shared.TestStatusValueFromString("TIMEOUT")},
			},
		},
	}, rq)
}
