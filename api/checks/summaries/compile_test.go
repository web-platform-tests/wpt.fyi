// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompile_Completed(t *testing.T) {
	foo := Completed{
		HostName: "foo.com",
		HostURL:  "foo.com/results/",
		DiffURL:  "foo.com/diff/",
		SHAURL:   "foo.com/sha/",
	}
	s, err := foo.Compile()
	assert.Nil(t, err)
	assert.Contains(t, s, foo.HostName)
	assert.Contains(t, s, foo.HostURL)
	assert.Contains(t, s, foo.DiffURL)
	assert.Contains(t, s, foo.SHAURL)
}

func TestCompile_Pending(t *testing.T) {
	foo := Pending{
		HostName: "foo.com",
		RunsURL:  "foo.com/runs",
	}
	s, err := foo.Compile()
	assert.Nil(t, err)
	assert.Contains(t, s, foo.HostName)
	assert.Contains(t, s, foo.RunsURL)
}
