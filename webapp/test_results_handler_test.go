// +build small

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTestRunUIFilter(t *testing.T) {
	r := httptest.NewRequest("GET", "/results/?max-count=3", nil)
	f, err := parseTestRunUIFilter(r)
	assert.Nil(t, err)
	assert.True(t, f.MaxCount != nil && *f.MaxCount == 3)

	r = httptest.NewRequest("GET", "/results/?products=chrome,safari&diff", nil)
	f, err = parseTestRunUIFilter(r)
	assert.Nil(t, err)
	assert.Equal(t, f.Products, "[\"chrome\",\"safari\"]")
	assert.Equal(t, f.Diff, true)
}
