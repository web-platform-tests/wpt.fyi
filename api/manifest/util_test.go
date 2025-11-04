//go:build small

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	// This is a smoke test; see shared/manifest_test.go for more tests.
	bytes := []byte(`{
"items": {
	"testharness": {
		"foo": {
			"bar": {
				"test.html": [
					"1d3166465cc6f2e8f9f18f53e499ca61e12d59bd",
					[null, {}]
				]

			}
		},
		"foobar": {
			"mytest.html": [
				"2d3166465cc6f2e8f9f18f53e499ca61e12d59bd",
				[null, {}]
			]
		}
	},
	"manual": {
		"foobar": {
			"test-manual.html": [
				"3d3166465cc6f2e8f9f18f53e499ca61e12d59bd",
				[null, {}]
			]
		}
	}
},
"version": 8
}`)

	filtered, err := Filter(bytes, []string{"/foo/bar"})
	assert.Nil(t, err)
	assert.Equal(t, `{"items":{"testharness":{"foo":{"bar":{"test.html":["1d3166465cc6f2e8f9f18f53e499ca61e12d59bd",[null,{}]]}}}},"version":8}`, string(filtered))
}
