// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSummary_Completed(t *testing.T) {
	foo := Completed{
		DiffURL:  "foo.com/diff?before=chrome[master]&after=chrome@0123456789",
		HostName: "foo.com",
		HostURL:  "foo.com/results/",
		SHAURL:   "foo.com/sha/",
	}
	s, err := foo.GetSummary()
	assert.Nil(t, err)
	assert.Contains(t, s, foo.HostName)
	assert.Contains(t, s, foo.HostURL)
	assert.Contains(t, s, foo.DiffURL)
	assert.Contains(t, s, foo.SHAURL)
}

func TestGetSummary_Pending(t *testing.T) {
	foo := Pending{
		HostName: "foo.com",
		RunsURL:  "foo.com/runs?products=chrome&sha=0123456789",
	}
	s, err := foo.GetSummary()
	assert.Nil(t, err)
	assert.Contains(t, s, foo.HostName)
	assert.Contains(t, s, foo.RunsURL)
}
