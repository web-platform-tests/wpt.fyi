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
	srcs, runs, err := getTestRunsAndSources(r, "latest")
	assert.Nil(t, err)
	assert.Nil(t, runs)
	assert.Equal(t, []string{"/api/runs?complete=true"}, srcs)
}
