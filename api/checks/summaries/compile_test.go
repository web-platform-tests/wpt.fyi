// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
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

func TestGetSummary_Regressed(t *testing.T) {
	master := shared.TestRun{}
	master.Revision = "abcdef0123"
	master.FullRevisionHash = strings.Repeat(master.Revision, 4)
	pr := shared.TestRun{}
	pr.Revision = "0123456789"
	pr.FullRevisionHash = strings.Repeat(master.Revision, 4)
	foo := Regressed{
		MasterRun:     master,
		PRRun:         pr,
		HostName:      "foo.com",
		HostURL:       "https://foo.com/",
		DiffURL:       "https://foo.com/?products=chrome@0000000000,chrome@0123456789&diff",
		MasterDiffURL: "https://foo.com/?products=chrome[master],chrome@0123456789&diff",
		Regressions: map[string]BeforeAndAfter{
			"/foo.html": BeforeAndAfter{
				PassingBefore: 1,
				TotalBefore:   1,
				PassingAfter:  0,
				TotalAfter:    1,
			},
		},
		More: 1,
	}
	s, err := foo.GetSummary()
	assert.Nil(t, err)
	assert.Contains(t, s, foo.HostName)
	assert.Contains(t, s, foo.HostURL)
	assert.Contains(t, s, foo.DiffURL)
	assert.Contains(t, s, master.String())
	assert.Contains(t, s, pr.String())
	assert.Contains(t, s, "0 / 1")
	assert.Contains(t, s, "1 / 1")
	assert.Contains(t, s, "And 1 others...")
}
