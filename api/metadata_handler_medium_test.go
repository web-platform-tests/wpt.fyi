// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestFilterMetadata(t *testing.T) {
	metadata := shared.MetadataResults(shared.MetadataResults{shared.MetadataResult{Test: "/foo/bar/b.html", URLs: []string{"", "https://aa.com/item", "https://bug.com/item"}}, shared.MetadataResult{Test: "bar", URLs: []string{"", "https://external.com/item", ""}}})
	abstractLink := query.AbstractLink{Pattern: "bug.com"}

	res := filterMetadata(abstractLink, metadata)

	assert.Equal(t, 1, len(res))
	assert.Equal(t, "/foo/bar/b.html", res[0].Test)
	assert.Equal(t, "", res[0].URLs[0])
	assert.Equal(t, "https://aa.com/item", res[0].URLs[1])
	assert.Equal(t, "https://bug.com/item", res[0].URLs[2])
}
