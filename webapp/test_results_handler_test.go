// +build small

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"fmt"
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

func TestParseTestRunUIFilter_BeforeAfter_Base64(t *testing.T) {
	enc := "eyJicm93c2VyX25hbWUiOiJGaXJlZm94IiwiYnJvd3Nlcl92ZXJzaW9uIjoiTmlnaHRseSIsIm9zX25hbWUiOiJUcmF2aXMiLCJvc192ZXJzaW9uIjoiSm9iIDU0LjQiLCJyZXZpc2lvbiI6ImYwNzdlMjQyZGUiLCJyZXN1bHRzX3VybCI6Imh0dHBzOi8vcHVsbHMtc3RhZ2luZy53ZWItcGxhdGZvcm0tdGVzdHMub3JnL2pvYi81NC40L3N1bW1hcnkiLCJjcmVhdGVkX2F0IjoiMjAxOC0wMS0wNFQwMDowMDowMFoifQ=="
	f, err := parseTestRunUIFilter(
		httptest.NewRequest("GET", fmt.Sprintf("/results/?before=chrome&after=%s", enc), nil))
	assert.Nil(t, err)
	assert.Equal(t, "[\"chrome\"]", f.Products)
	assert.Equal(t, "Firefox", f.AfterTestRun.BrowserName)
	assert.Equal(t, true, f.Diff)
}
