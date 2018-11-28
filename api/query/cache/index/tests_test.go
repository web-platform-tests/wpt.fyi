// +build small

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
	var subName *string
	id, err := ts.Add(name, subName)
	assert.Nil(t, err)
	actualName, actualSubName, err := ts.GetName(id)
	assert.Nil(t, err)
	assert.Equal(t, name, actualName)
	assert.Equal(t, subName, actualSubName)

	subNameValue := "some sub name"
	subName = &subNameValue
	id, err = ts.Add(name, subName)
	assert.Nil(t, err)
	actualName, actualSubName, err = ts.GetName(id)
	assert.Nil(t, err)
	assert.Equal(t, name, actualName)
	assert.Equal(t, *subName, *actualSubName)
}
