// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBrowserNames(t *testing.T) {
	names, _ := GetBrowserNames()
	assert.True(t, sort.StringsAreSorted(names))
}

func TestIsBrowserName(t *testing.T) {
	assert.True(t, IsBrowserName("chrome"))
	assert.True(t, IsBrowserName("edge"))
	assert.True(t, IsBrowserName("firefox"))
	assert.True(t, IsBrowserName("safari"))
	assert.False(t, IsBrowserName("not-a-browser"))
}

func TestIsBrowserName_Names(t *testing.T) {
	names, _ := GetBrowserNames()
	for _, name := range names {
		assert.True(t, IsBrowserName(name))
	}
}
