//go:build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestForRun_fail(t *testing.T) {
	rs := NewResults()
	rrs := rs.ForRun(RunID(0))
	assert.Nil(t, rrs)
}

func TestAdd_fail(t *testing.T) {
	rs := NewResults()
	rrs1 := NewRunResults()
	rrs2 := NewRunResults()
	err := rs.Add(RunID(0), rrs1)
	assert.Nil(t, err)
	err = rs.Add(RunID(0), rrs2)
	assert.NotNil(t, err)
}

func TestAddForRunGetResult(t *testing.T) {
	rs := NewResults()
	rrs := NewRunResults()

	ru := RunID(0)
	re := ResultID(shared.TestStatusOK)
	te := TestID{}

	rrs.Add(re, te)
	err := rs.Add(ru, rrs)
	assert.Nil(t, err)
	rrs = rs.ForRun(ru)
	assert.NotNil(t, rrs)
	assert.Equal(t, re, rrs.GetResult(te))
	assert.Equal(t, ResultID(shared.TestStatusUnknown), rrs.GetResult(TestID{1, 1}))
}
