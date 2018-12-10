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

func TestFilterManifest_Reftest(t *testing.T) {
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
	filtered, err := filterManifest(bytes, []string{"/css/css-images/tiled-gradients.html"})
	assert.Nil(t, err)
	unmarshalled := shared.Manifest{}
	json.Unmarshal(filtered, &unmarshalled)
	assert.NotNil(t, unmarshalled.Items.Reftest)
	assert.Equal(t, 1, len(unmarshalled.Items.Reftest))

	// Prefix
	filtered, err = filterManifest(bytes, []string{"/css/css-images/"})
	assert.Nil(t, err)
	unmarshalled = shared.Manifest{}
	json.Unmarshal(filtered, &unmarshalled)
	assert.NotNil(t, unmarshalled.Items.Reftest)
	assert.Equal(t, 2, len(unmarshalled.Items.Reftest))

	// No matches
	filtered, err = filterManifest(bytes, []string{"/not-a-folder/test.html"})
	assert.Nil(t, err)
	unmarshalled = shared.Manifest{}
	json.Unmarshal(filtered, &unmarshalled)
	assert.Equal(t, 0, len(unmarshalled.Items.Reftest))
}
