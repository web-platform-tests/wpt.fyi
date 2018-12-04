// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapStringKeys(t *testing.T) {
	m := map[string]int{"foo": 1}
	keys, err := MapStringKeys(m)
	if err != nil {
		assert.FailNow(t, "Error getting map string keys")
	}
	assert.Equal(t, []string{"foo"}, keys)

	m2 := map[string]interface{}{"bar": "baz"}
	keys, err = MapStringKeys(m2)
	if err != nil {
		assert.FailNow(t, "Error getting map string keys")
	}
	assert.Equal(t, []string{"bar"}, keys)
}

func TestMapStringKeys_NotAMap(t *testing.T) {
	one := 1
	keys, err := MapStringKeys(one)
	assert.Nil(t, keys)
	assert.NotNil(t, err)
}

func TestMapStringKeys_NotAStringKeyedMap(t *testing.T) {
	m := map[int]int{1: 1}
	keys, err := MapStringKeys(m)
	assert.Nil(t, keys)
	assert.NotNil(t, err)
}

func TestProductChannelToLabel(t *testing.T) {
	assert.Equal(t, StableLabel, ProductChannelToLabel("release"))
	assert.Equal(t, StableLabel, ProductChannelToLabel("stable"))
	assert.Equal(t, BetaLabel, ProductChannelToLabel("beta"))
	assert.Equal(t, ExperimentalLabel, ProductChannelToLabel("dev"))
	assert.Equal(t, ExperimentalLabel, ProductChannelToLabel("nightly"))
	assert.Equal(t, ExperimentalLabel, ProductChannelToLabel("preview"))
	assert.Equal(t, ExperimentalLabel, ProductChannelToLabel("experimental"))
	assert.Equal(t, "", ProductChannelToLabel("not-a-channel"))
}
