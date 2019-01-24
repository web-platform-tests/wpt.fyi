// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import (
	"flag"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// To output the rendered content during execution of the test(s), set this flag, e.g.
// go test ./api/checks/summaries -tags="medium" -print_output -test.v
var renderOutputToConsole = flag.Bool("print_output", false, "Whether to render compiled markdown during test execution.")

func TestGetSummary_Completed(t *testing.T) {
	master := shared.TestRun{}
	master.BrowserName = "chrome"
	master.Revision = "abcdef0123"
	master.FullRevisionHash = strings.Repeat(master.Revision, 4)
	pr := shared.TestRun{}
	pr.BrowserName = "chrome"
	pr.Revision = "0123456789"
	pr.FullRevisionHash = strings.Repeat(pr.Revision, 4)
	foo := Completed{}
	foo.BaseRun = master
	foo.HeadRun = pr
	foo.DiffURL = "https://foo.com/diff?before=chrome[master]&after=chrome@0123456789"
	foo.HostName = "foo.com"
	foo.HostURL = "https://foo.com/"
	testName := "/foo.html?exclude=(Document|window|HTML.*)"
	foo.Results = map[string][]int{
		testName: []int{2, 3},
	}
	foo.More = 1

	s, err := foo.GetSummary()
	printOutput(s)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Contains(t, s, foo.HostName)
	assert.Contains(t, s, foo.HostURL)
	assert.Contains(t, s, foo.DiffURL)
	assert.Contains(t, s, "2 / 3")
	assert.Contains(t, s, "And 1 others...")
	assert.Contains(t, s, foo.FileIssueURL().String())

	// And with MasterDiffURL
	foo.MasterDiffURL = "https://foo.com/?products=chrome[master],chrome@0123456789&diff"
	s, err = foo.GetSummary()
	printOutput(s)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Contains(t, s, foo.MasterDiffURL)

	// With PRNumbers
	foo.PRNumbers = []int{123}
	s, err = foo.GetSummary()
	printOutput(s)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Contains(t, s, escapeMD(testName))
	assert.Contains(t, s, "https://foo.com/runs/?pr=123")
	assert.Contains(t, s, "https://foo.com/results/?pr=123")
}

func TestGetSummary_Pending(t *testing.T) {
	foo := Pending{
		HostName: "https://foo.com",
		RunsURL:  "https://foo.com/runs?products=chrome&sha=0123456789",
	}
	s, err := foo.GetSummary()
	printOutput(s)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Contains(t, s, foo.HostName)
	assert.Contains(t, s, foo.RunsURL)
	assert.Contains(t, s, foo.FileIssueURL().String())
}

func TestGetSummary_Regressed(t *testing.T) {
	master := shared.TestRun{}
	master.BrowserName = "chrome"
	master.Revision = "abcdef0123"
	master.FullRevisionHash = strings.Repeat(master.Revision, 4)
	pr := shared.TestRun{}
	pr.BrowserName = "chrome"
	pr.Revision = "0123456789"
	pr.FullRevisionHash = strings.Repeat(pr.Revision, 4)
	foo := Regressed{}
	foo.BaseRun = master
	foo.HeadRun = pr
	foo.HostName = "foo.com"
	foo.HostURL = "https://foo.com/"
	foo.DiffURL = "https://foo.com/?products=chrome@0000000000,chrome@0123456789&diff"
	testName := "/foo.html?exclude=(Document|window|HTML.*)"
	foo.Regressions = map[string]BeforeAndAfter{
		testName: BeforeAndAfter{
			PassingBefore: 1,
			TotalBefore:   1,
			PassingAfter:  0,
			TotalAfter:    1,
		},
	}
	foo.More = 1

	s, err := foo.GetSummary()
	printOutput(s)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Contains(t, s, foo.HostName)
	assert.Contains(t, s, foo.HostURL)
	assert.Contains(t, s, foo.DiffURL)
	assert.Contains(t, s, master.String())
	assert.Contains(t, s, pr.String())
	assert.Contains(t, s, "0 / 1")
	assert.Contains(t, s, "1 / 1")
	assert.Contains(t, s, "And 1 others...")
	assert.Contains(t, s, foo.FileIssueURL().String())

	// And with MasterDiffURL
	foo.MasterDiffURL = "https://foo.com/?products=chrome[master],chrome@0123456789&diff"
	s, err = foo.GetSummary()
	printOutput(s)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Contains(t, s, foo.MasterDiffURL)

	// With PRNumbers
	foo.PRNumbers = []int{123}
	s, err = foo.GetSummary()
	printOutput(s)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Contains(t, s, "https://foo.com/runs/?pr=123")
	assert.Contains(t, s, "https://foo.com/results/?pr=123")
}

func printOutput(s string) {
	if *renderOutputToConsole {
		log.Printf("MD output:\n-----------\n%s", s)
	}
}
