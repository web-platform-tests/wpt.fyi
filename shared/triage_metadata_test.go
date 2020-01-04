// +build small
// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestAppendTestName(t *testing.T) {
	var actual, expected MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/bar.html": [
			{
				"url": "bugs.bar?id=456",
				"product": "chrome",
				"results": [
					{"status": 6 }
				]
			}
		],
		"/foo1/bar1.html": [
			{
				"product": "chrome",
				"url": "bugs.bar",
				"results": [
					{"status": 6 },
					{"status": 3 }
				]}
		]
	}`), &actual)

	json.Unmarshal([]byte(`{
		"/foo/bar.html": [
			{
				"url": "bugs.bar?id=456",
				"product": "chrome",
				"results": [
					{"status": 6 }
				]
			}
		],
		"/foo1/bar1.html": [
			{
				"product": "chrome",
				"url": "bugs.bar",
				"results": [
					{"status": 6, "test": "bar1.html"},
					{"status": 3, "test": "bar1.html"}
				]}
		]
	}`), &expected)
	test := "/foo1/bar1.html"

	appendTestName(test, actual)

	assert.Equal(t, expected, actual)
}

func TestAddToFiles_AddNewFile(t *testing.T) {
	var amendment MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/foo1/bar.html": [
			{
				"url": "bugs.bar?id=456",
				"product": "chrome",
				"results": [
					{"status": 6 }
				]
			}
		]
	}`), &amendment)

	var path = "a"
	var fileMap = make(map[string]Metadata)
	fileInBytes := []byte(`
links:
  - product: chrome-64
    url: https://external.com/item
    results:
    - test: a.html
  - product: firefox-2
    url: https://bug.com/item
    results:
    - test: b.html
      subtest: Something should happen
      status: FAIL
    - test: c.html
`)
	var file Metadata
	yaml.Unmarshal(fileInBytes, &file)
	fileMap[path] = file

	actualMap := addToFiles(amendment, fileMap, NewNilLogger())

	assert.Equal(t, 1, len(actualMap))
	actualInBytes, ok := actualMap["foo/foo1"]
	assert.True(t, ok)

	var actual Metadata
	yaml.Unmarshal(actualInBytes, &actual)
	assert.Equal(t, 1, len(actual.Links))
	assert.Equal(t, "chrome", actual.Links[0].Product.BrowserName)
	assert.Equal(t, "bugs.bar?id=456", actual.Links[0].URL)
	assert.Equal(t, 1, len(actual.Links[0].Results))
	assert.Equal(t, "bar.html", actual.Links[0].Results[0].TestPath)
	assert.Equal(t, TestStatusFail, *actual.Links[0].Results[0].Status)
}

func TestAddToFiles_AddNewMetadataResult(t *testing.T) {
	var amendment MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/foo1/a.html": [
			{
				"url": "foo",
				"product": "chrome",
				"results": [
					{"status": 6 }
				]
			}
		]
	}`), &amendment)

	var path = "foo/foo1"
	var fileMap = make(map[string]Metadata)
	fileInBytes := []byte(`
links:
  - product: chrome
    url: foo
    results:
    - test: b.html
  - product: firefox-2
    url: https://bug.com/item
    results:
    - test: b.html
      subtest: Something should happen
      status: FAIL
    - test: c.html
`)
	var file Metadata
	yaml.Unmarshal(fileInBytes, &file)
	fileMap[path] = file

	actualMap := addToFiles(amendment, fileMap, NewNilLogger())

	assert.Equal(t, 1, len(actualMap))
	actualInBytes, ok := actualMap["foo/foo1"]
	assert.True(t, ok)

	var actual Metadata
	yaml.Unmarshal(actualInBytes, &actual)
	assert.Equal(t, 2, len(actual.Links))
	assert.Equal(t, "chrome", actual.Links[0].Product.BrowserName)
	assert.Equal(t, "foo", actual.Links[0].URL)
	assert.Equal(t, 2, len(actual.Links[0].Results))
	assert.Equal(t, "b.html", actual.Links[0].Results[0].TestPath)
	assert.Equal(t, "a.html", actual.Links[0].Results[1].TestPath)
	assert.Equal(t, TestStatusFail, *actual.Links[0].Results[1].Status)
	assert.Equal(t, "firefox", actual.Links[1].Product.BrowserName)
	assert.Equal(t, "https://bug.com/item", actual.Links[1].URL)
}

func TestAddToFiles_AddNewMetadataLink(t *testing.T) {
	var amendment MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/foo1/a.html": [
			{
				"url": "foo1",
				"product": "chrome",
				"results": [
					{"status": 6 }
				]
			}
		]
	}`), &amendment)

	var path = "foo/foo1"
	var fileMap = make(map[string]Metadata)
	fileInBytes := []byte(`
links:
  - product: chrome
    url: foo
    results:
    - test: b.html
  - product: firefox-2
    url: https://bug.com/item
    results:
    - test: b.html
      subtest: Something should happen
      status: FAIL
    - test: c.html
`)
	var file Metadata
	yaml.Unmarshal(fileInBytes, &file)
	fileMap[path] = file

	actualMap := addToFiles(amendment, fileMap, NewNilLogger())

	assert.Equal(t, 1, len(actualMap))
	actualInBytes, ok := actualMap["foo/foo1"]
	assert.True(t, ok)

	var actual Metadata
	yaml.Unmarshal(actualInBytes, &actual)
	assert.Equal(t, 3, len(actual.Links))
	assert.Equal(t, "chrome", actual.Links[0].Product.BrowserName)
	assert.Equal(t, "foo", actual.Links[0].URL)
	assert.Equal(t, 1, len(actual.Links[0].Results))
	assert.Equal(t, "b.html", actual.Links[0].Results[0].TestPath)
	assert.Equal(t, "firefox", actual.Links[1].Product.BrowserName)
	assert.Equal(t, "https://bug.com/item", actual.Links[1].URL)
	assert.Equal(t, "chrome", actual.Links[2].Product.BrowserName)
	assert.Equal(t, "foo1", actual.Links[2].URL)
}

func TestGenerateRandomInt(t *testing.T) {
	int1 := generateRandomInt()
	int2 := generateRandomInt()
	assert.True(t, int1 != int2)
}
