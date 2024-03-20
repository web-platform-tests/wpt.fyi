//go:build small
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
		assert.NotEqual(t, "android_webview", n)
		assert.NotEqual(t, "chrome_android", n)
		assert.NotEqual(t, "chrome_ios", n)
		assert.NotEqual(t, "chromium", n)
		assert.NotEqual(t, "deno", n)
		assert.NotEqual(t, "epiphany", n)
		assert.NotEqual(t, "firefox_android", n)
		assert.NotEqual(t, "flow", n)
		assert.NotEqual(t, "node.js", n)
		assert.NotEqual(t, "servo", n)
		assert.NotEqual(t, "wktr", n)
		assert.NotEqual(t, "webkitgtk", n)
		assert.NotEqual(t, "uc", n)
	}
}

func TestIsBrowserName(t *testing.T) {
	assert.True(t, IsBrowserName("chrome"))
	assert.True(t, IsBrowserName("chromium"))
	assert.True(t, IsBrowserName("deno"))
	assert.True(t, IsBrowserName("edge"))
	assert.True(t, IsBrowserName("firefox"))
	assert.True(t, IsBrowserName("firefox_android"))
	assert.True(t, IsBrowserName("flow"))
	assert.True(t, IsBrowserName("node.js"))
	assert.True(t, IsBrowserName("safari"))
	assert.True(t, IsBrowserName("chrome_android"))
	assert.True(t, IsBrowserName("chrome_ios"))
	assert.True(t, IsBrowserName("android_webview"))
	assert.True(t, IsBrowserName("epiphany"))
	assert.True(t, IsBrowserName("servo"))
	assert.True(t, IsBrowserName("wktr"))
	assert.True(t, IsBrowserName("webkitgtk"))
	assert.True(t, IsBrowserName("uc"))
	assert.False(t, IsBrowserName("not-a-browser"))
}

func TestIsBrowserName_DefaultBrowsers(t *testing.T) {
	names := GetDefaultBrowserNames()
	for _, name := range names {
		assert.True(t, IsBrowserName(name))
	}
}
