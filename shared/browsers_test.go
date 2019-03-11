// +build small

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultBrowserNames(t *testing.T) {
	names := GetDefaultBrowserNames()
	assert.True(t, sort.StringsAreSorted(names))
	// Non default browser names:
	for _, n := range names {
		assert.NotEqual(t, "uc", n)
	}
}

func TestIsBrowserName(t *testing.T) {
	assert.True(t, IsBrowserName("chrome"))
	assert.True(t, IsBrowserName("edge"))
	assert.True(t, IsBrowserName("firefox"))
	assert.True(t, IsBrowserName("safari"))
	assert.True(t, IsBrowserName("uc"))
	assert.False(t, IsBrowserName("not-a-browser"))
}

func TestIsBrowserName_DefaultBrowsers(t *testing.T) {
	names := GetDefaultBrowserNames()
	for _, name := range names {
		assert.True(t, IsBrowserName(name))
	}
}
