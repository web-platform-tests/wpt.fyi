// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const shortSHA = "abcdef0123"
const resultsURLBase = "https://storage.googleapis.com/wptd/" + shortSHA + "/"
const product = "chrome-63.0-linux"
const resultsURL = resultsURLBase + "/" + product + "-summary.json.gz"

func TestMapStringKeys(t *testing.T) {
	m := map[string]int{"foo": 1}
	keys, err := MapStringKeys(m)
	if err != nil {
		assert.FailNow(t, "Error getting map string keys")
	}
	assert.Equal(t, []string{"foo"}, keys)

	m2 := map[string]interface{}{"bar": "baz"}
	keys, err = MapStringKeys(m2)
	if err != nil {
		assert.FailNow(t, "Error getting map string keys")
	}
	assert.Equal(t, []string{"bar"}, keys)
}

func TestMapStringKeys_NotAMap(t *testing.T) {
	one := 1
	keys, err := MapStringKeys(one)
	assert.Nil(t, keys)
	assert.NotNil(t, err)
}

func TestMapStringKeys_NotAStringKeyedMap(t *testing.T) {
	m := map[int]int{1: 1}
	keys, err := MapStringKeys(m)
	assert.Nil(t, keys)
	assert.NotNil(t, err)
}

func TestProductChannelToLabel(t *testing.T) {
	assert.Equal(t, StableLabel, ProductChannelToLabel("release"))
	assert.Equal(t, StableLabel, ProductChannelToLabel("stable"))
	assert.Equal(t, BetaLabel, ProductChannelToLabel("beta"))
	assert.Equal(t, ExperimentalLabel, ProductChannelToLabel("dev"))
	assert.Equal(t, ExperimentalLabel, ProductChannelToLabel("nightly"))
	assert.Equal(t, ExperimentalLabel, ProductChannelToLabel("preview"))
	assert.Equal(t, ExperimentalLabel, ProductChannelToLabel("experimental"))
	assert.Equal(t, "", ProductChannelToLabel("not-a-channel"))
}

func TestGetResultsURL_EmptyFile(t *testing.T) {
	run := TestRun{ResultsURL: resultsURL}
	run.Revision = shortSHA
	checkResult(t, run, "", resultsURL)
}

func TestGetResultsURL_TestFile(t *testing.T) {
	run := TestRun{ResultsURL: resultsURL}
	run.Revision = shortSHA
	file := "css/vendor-imports/mozilla/mozilla-central-reftests/flexbox/flexbox-root-node-001b.html"
	checkResult(t, run, file, resultsURLBase+product+"/"+file)
}

func TestGetResultsURL_TrailingSlash(t *testing.T) {
	run := TestRun{ResultsURL: resultsURL}
	run.Revision = shortSHA
	checkResult(t, run, "/", resultsURL)
}

func checkResult(t *testing.T, testRun TestRun, testFile string, expected string) {
	got := GetResultsURL(testRun, testFile)
	if got != expected {
		t.Errorf("\nGot:\n%q\nExpected:\n%q", got, expected)
	}
}

func TestGetSharedPath(t *testing.T) {
	assert.Equal(t, "/a/b/c.html", GetSharedPath("/a/b/c.html"))
	assert.Equal(t, "/a/b/", GetSharedPath("/a/b/c.html", "/a/b/d.html"))
	assert.Equal(t, "/", GetSharedPath("/a/b/c.html", "/d/e/f.html"))
	assert.Equal(t, "/a/", GetSharedPath("/a/z.html", "/a/b/x.html", "/a/b/y.html"))
}
