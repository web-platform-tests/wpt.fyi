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
	f, err := parseTestRunUIFilter(httptest.NewRequest("GET", "/results/?max-count=3", nil))
	assert.Nil(t, err)
	assert.True(t, f.MaxCount != nil && *f.MaxCount == 3)

	f, err = parseTestRunUIFilter(httptest.NewRequest("GET", "/results/?products=chrome,safari&diff", nil))
	assert.Nil(t, err)
	assert.Equal(t, "[\"chrome\",\"safari\"]", f.Products)
	assert.Equal(t, true, f.Diff)
}

func TestParseTestRunUIFilter_BeforeAfter(t *testing.T) {
	f, err := parseTestRunUIFilter(httptest.NewRequest("GET", "/results/?before=chrome&after=chrome[experimental]", nil))
	assert.Nil(t, err)
	assert.Equal(t, "[\"chrome\",\"chrome[experimental]\"]", f.Products)
	assert.Equal(t, true, f.Diff)

	f, err = parseTestRunUIFilter(httptest.NewRequest("GET", "/results/?after=chrome&before=edge", nil))
	assert.Nil(t, err)
	assert.Equal(t, "[\"edge\",\"chrome\"]", f.Products)
	assert.Equal(t, true, f.Diff)
}
