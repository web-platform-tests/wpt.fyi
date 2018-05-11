// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"testing"

	base "github.com/web-platform-tests/wpt.fyi/shared"
)

type Case struct {
	testRun  base.TestRun
	testFile string
	expected string
}

const shortSHA = "abcdef0123"
const resultsURLBase = "https://storage.googleapis.com/wptd/" + shortSHA + "/"
const platform = "chrome-63.0-linux"
const resultsURL = resultsURLBase + "/" + platform + "-summary.json.gz"

func TestGetResultsURL_EmptyFile(t *testing.T) {
	checkResult(
		t,
		Case{
			base.TestRun{
				ResultsURL: resultsURL,
				Revision:   shortSHA,
			},
			"",
			resultsURL,
		})
}

func TestGetResultsURL_TestFile(t *testing.T) {
	file := "css/vendor-imports/mozilla/mozilla-central-reftests/flexbox/flexbox-root-node-001b.html"
	checkResult(
		t,
		Case{
			base.TestRun{
				ResultsURL: resultsURL,
				Revision:   shortSHA,
			},
			file,
			resultsURLBase + platform + "/" + file,
		})
}

func TestGetResultsURL_TrailingSlash(t *testing.T) {
	checkResult(
		t,
		Case{
			base.TestRun{
				ResultsURL: resultsURL,
				Revision:   shortSHA,
			},
			"/",
			resultsURL,
		})
}

func checkResult(t *testing.T, c Case) {
	got := getResultsURL(c.testRun, c.testFile)
	if got != c.expected {
		t.Errorf("\nGot:\n%q\nExpected:\n%q", got, c.expected)
	}
}
