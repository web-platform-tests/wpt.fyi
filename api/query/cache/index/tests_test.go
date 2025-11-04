//go:build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetName_fail(t *testing.T) {
	ts := NewTests()
	_, _, err := ts.GetName(TestID{})
	assert.NotNil(t, err)
}

func TestAddGetName(t *testing.T) {
	ts := NewTests()
	name := "/a/b/c"
	id, err := computeTestID(name, nil)
	assert.Nil(t, err)
	ts.Add(id, name, nil)
	actualName, actualSubName, err := ts.GetName(id)
	assert.Nil(t, err)
	assert.Equal(t, name, actualName)
	assert.Nil(t, actualSubName)

	subNameValue := "some sub name"
	subName := &subNameValue
	id, err = computeTestID(name, subName)
	assert.Nil(t, err)
	ts.Add(id, name, subName)
	actualName, actualSubName, err = ts.GetName(id)
	assert.Nil(t, err)
	assert.Equal(t, name, actualName)
	assert.Equal(t, *subName, *actualSubName)
}
