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

func TestGetTestRunsAndSources(t *testing.T) {
	r := httptest.NewRequest("GET", "/results/?max-count=3", nil)
	srcs, runs, err := getTestRunsAndSources(r)
	assert.Nil(t, err)
	assert.Nil(t, runs)
	assert.Equal(t, []string{"/api/runs?complete=true&max-count=3"}, srcs)

	r = httptest.NewRequest("GET", "/results/?max-count=5&product=chrome-69&sha=abcdef0123", nil)
	srcs, runs, err = getTestRunsAndSources(r)
	assert.Nil(t, err)
	assert.Nil(t, runs)
	assert.Equal(t, []string{"/api/runs?max-count=5&product=chrome-69&sha=abcdef0123"}, srcs)
}
