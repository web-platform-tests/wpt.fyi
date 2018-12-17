// +build small

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestFilter_Reftest(t *testing.T) {
	bytes := []byte(`{
	"items": {
		"reftest": {
      "css/css-images/linear-gradient-2.html": [
        [
					"/css/css-images/linear-gradient-2.html",
					[ ["/css/css-images/linear-gradient-ref.html","=="] ],
          {}
        ]
      ],
      "css/css-images/tiled-gradients.html": [
        [
					"/css/css-images/tiled-gradients.html",
					[ ["/css/css-images/tiled-gradients-ref.html","=="] ],
          {}
        ]
			]
		}
	}
}`)

	// Specific file
	filtered, err := Filter(bytes, []string{"/css/css-images/tiled-gradients.html"})
	assert.Nil(t, err)
	unmarshalled := shared.Manifest{}
	json.Unmarshal(filtered, &unmarshalled)
	assert.NotNil(t, unmarshalled.Items.Reftest)
	assert.Equal(t, 1, len(unmarshalled.Items.Reftest))

	// Prefix
	filtered, err = Filter(bytes, []string{"/css/css-images/"})
	assert.Nil(t, err)
	unmarshalled = shared.Manifest{}
	json.Unmarshal(filtered, &unmarshalled)
	assert.NotNil(t, unmarshalled.Items.Reftest)
	assert.Equal(t, 2, len(unmarshalled.Items.Reftest))

	// No matches
	filtered, err = Filter(bytes, []string{"/not-a-folder/test.html"})
	assert.Nil(t, err)
	unmarshalled = shared.Manifest{}
	json.Unmarshal(filtered, &unmarshalled)
	assert.Equal(t, 0, len(unmarshalled.Items.Reftest))
}

func TestExplodePossibleFilename_AnyJS(t *testing.T) {
	filename := "/test/file.any.js"
	exploded := []string{
		"/test/file.any.html",
		"/test/file.any.worker.html",
		"/test/file.any.serviceworker.html",
		"/test/file.any.sharedworker.html",
	}
	assert.Equal(t, ExplodePossibleFilenames(filename), exploded)
}

func TestExplodePossibleFilename_WindowJS(t *testing.T) {
	filename := "/test/file.window.js"
	exploded := []string{
		"/test/file.window.html",
	}
	assert.Equal(t, ExplodePossibleFilenames(filename), exploded)
}

func TestExplodePossibleFilename_Standard(t *testing.T) {
	filename := "/test/file.html"
	assert.Nil(t, ExplodePossibleFilenames(filename))
}

func TestExplodePossibleRenames_AnyJS(t *testing.T) {
	before, after := "/test/file.any.js", "/test/file.https.any.js"
	renames := map[string]string{
		before:                              after,
		"/test/file.any.html":               "/test/file.https.any.html",
		"/test/file.any.worker.html":        "/test/file.https.any.worker.html",
		"/test/file.any.serviceworker.html": "/test/file.https.any.serviceworker.html",
		"/test/file.any.sharedworker.html":  "/test/file.https.any.sharedworker.html",
	}
	assert.Equal(t, ExplodePossibleRenames(before, after), renames)
}

func TestExplodePossibleRenames_WindowJS(t *testing.T) {
	before, after := "/test/file.window.js", "/test/file.https.window.js"
	renames := map[string]string{
		before: after,
		"/test/file.window.html": "/test/file.https.window.html",
	}
	assert.Equal(t, ExplodePossibleRenames(before, after), renames)
}
