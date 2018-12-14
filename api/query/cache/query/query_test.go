// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/query"
)

func TestPrepareUserQuery_basic(t *testing.T) {
	runIDs := []int64{1, 2, 3, 4}
	q := query.TestNamePattern{Pattern: ""}
	q2 := PrepareUserQuery(runIDs, q)

	// Assert structrure: AND(OR(...), q)
	and, ok := q2.(query.And)
	assert.True(t, ok)
	assert.Equal(t, 2, len(and.Args))
	assert.Equal(t, q, and.Args[1])
	_, ok = and.Args[0].(query.Or)
	assert.True(t, ok)
}

func TestPrepareUserQuery_and(t *testing.T) {
	runIDs := []int64{1, 2, 3, 4}
	q := query.And{
		Args: []query.ConcreteQuery{
			query.TestNamePattern{Pattern: "a"},
			query.TestNamePattern{Pattern: "b"},
		},
	}
	q2 := PrepareUserQuery(runIDs, q)

	// Assert structrure: AND(OR(...), q...)
	and, ok := q2.(query.And)
	assert.True(t, ok)
	assert.Equal(t, 1+len(q.Args), len(and.Args))
	_, ok = and.Args[0].(query.Or)
	assert.True(t, ok)
	for i, arg := range q.Args {
		assert.Equal(t, arg, and.Args[1+i])
	}
}
