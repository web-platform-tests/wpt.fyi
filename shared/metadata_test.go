// +build small

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"testing"

	"github.com/go-yaml/yaml"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	var metadata Metadata
	err := yaml.Unmarshal([]byte(`
links:
  - product: chrome-64
    test: a.html
    url: https://external.com/item`), &metadata)
	assert.Nil(t, err)
	assert.Equal(t, "chrome", metadata.Links[0].Product.BrowserName)
	assert.Equal(t, "64", metadata.Links[0].Product.BrowserVersion)
	assert.Equal(t, "a.html", metadata.Links[0].TestPath)
	assert.Equal(t, "https://external.com/item", metadata.Links[0].URL)
}
