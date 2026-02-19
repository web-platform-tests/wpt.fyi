// +build small

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestAppendTestName(t *testing.T) {
	var actual, expected MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/bar.html": [
			{
				"url": "https://bugs.bar?id=456",
				"product": "chrome",
				"results": [
					{"status": 6 }
				]
			}
		],
		"/foo1/bar1.html": [
			{
				"product": "chrome",
				"url": "https://bugs.bar",
				"results": [
					{"status": 6, "subtest": "sub-bar1" },
					{"status": 3 }
				]}
		]
	}`), &actual)

	json.Unmarshal([]byte(`{
		"/foo/bar.html": [
			{
				"url": "https://bugs.bar?id=456",
				"product": "chrome",
				"results": [
					{"status": 6 }
				]
			}
		],
		"/foo1/bar1.html": [
			{
				"product": "chrome",
				"url": "https://bugs.bar",
				"results": [
					{"status": 6, "test": "bar1.html", "subtest": "sub-bar1" },
					{"status": 3, "test": "bar1.html"}
				]}
		]
	}`), &expected)
	test := "/foo1/bar1.html"

	appendTestName(test, actual)

	assert.Equal(t, expected, actual)
}

func TestAppendTestName_EmptyResults(t *testing.T) {
	var actual, expected MetadataResults
	json.Unmarshal([]byte(`{
		"/foo1/bar1.html": [
			{
				"product": "chrome",
				"url": "https://bugs.bar"
			}
		]
	}`), &actual)

	json.Unmarshal([]byte(`{
		"/foo1/bar1.html": [
			{
				"product": "chrome",
				"url": "https://bugs.bar",
				"results": [
					{"test": "bar1.html"}
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
				"url": "https://bugs.bar?id=456",
				"product": "chrome",
				"results": [
					{"subtest": "sub-bar1"}
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
  - product: firefox
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
	assert.Equal(t, "https://bugs.bar?id=456", actual.Links[0].URL)
	assert.Equal(t, 1, len(actual.Links[0].Results))
	assert.Equal(t, "bar.html", actual.Links[0].Results[0].TestPath)
	assert.Equal(t, "sub-bar1", *actual.Links[0].Results[0].SubtestName)
}

func TestAddToFiles_AddNewMetadataResult(t *testing.T) {
	var amendment MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/foo1/a.html": [
			{
				"url": "https://foo",
				"product": "chrome",
				"results": [
					{"status": 6, "subtest": "sub-a" }
				]
			}
		]
	}`), &amendment)

	var path = "foo/foo1"
	var fileMap = make(map[string]Metadata)
	fileInBytes := []byte(`
links:
  - product: chrome
    url: https://foo
    results:
    - test: b.html
  - product: firefox
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
	assert.Equal(t, "https://foo", actual.Links[0].URL)
	assert.Equal(t, 2, len(actual.Links[0].Results))
	assert.Equal(t, "b.html", actual.Links[0].Results[0].TestPath)
	assert.Equal(t, "a.html", actual.Links[0].Results[1].TestPath)
	assert.Equal(t, "sub-a", *actual.Links[0].Results[1].SubtestName)
	assert.Equal(t, TestStatusFail, *actual.Links[0].Results[1].Status)
	assert.Equal(t, "firefox", actual.Links[1].Product.BrowserName)
	assert.Equal(t, "https://bug.com/item", actual.Links[1].URL)
}

func TestAddToFiles_AddNewMetadataLink(t *testing.T) {
	var amendment MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/foo1/a.html": [
			{
				"url": "https://foo1",
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
    url: https://foo
    results:
    - test: b.html
  - product: firefox
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
	assert.Equal(t, "https://foo", actual.Links[0].URL)
	assert.Equal(t, 1, len(actual.Links[0].Results))
	assert.Equal(t, "b.html", actual.Links[0].Results[0].TestPath)
	assert.Equal(t, "firefox", actual.Links[1].Product.BrowserName)
	assert.Equal(t, "https://bug.com/item", actual.Links[1].URL)
	assert.Equal(t, "chrome", actual.Links[2].Product.BrowserName)
	assert.Equal(t, "https://foo1", actual.Links[2].URL)
	assert.Equal(t, "a.html", actual.Links[2].Results[0].TestPath)
}

func TestAddToFiles_AddNewMetadataLink_Label(t *testing.T) {
	var amendment MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/foo1/abc.html": [
			{
				"label": "interop"
			}
		]
	}`), &amendment)

	var path = "foo/foo1"
	var fileMap = make(map[string]Metadata)
	fileInBytes := []byte(`
links:
  - product: chrome
    url: https://foo
    results:
    - test: b.html
  - product: firefox
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
	assert.Equal(t, "https://foo", actual.Links[0].URL)
	assert.Equal(t, 1, len(actual.Links[0].Results))
	assert.Equal(t, "b.html", actual.Links[0].Results[0].TestPath)
	assert.Equal(t, "firefox", actual.Links[1].Product.BrowserName)
	assert.Equal(t, "https://bug.com/item", actual.Links[1].URL)
	assert.Equal(t, "", actual.Links[2].Product.String())
	assert.Equal(t, "", actual.Links[2].URL)
	assert.Equal(t, "interop", actual.Links[2].Label)
}

func TestAddToFiles_AddNewMetadataResults_Label(t *testing.T) {
	var amendment MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/foo1/abc.html": [
			{
				"label": "interop"
			}
		]
	}`), &amendment)

	var path = "foo/foo1"
	var fileMap = make(map[string]Metadata)
	fileInBytes := []byte(`
links:
  - label: interop
    results:
    - test: b.html
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
	assert.Equal(t, "", actual.Links[0].Product.BrowserName)
	assert.Equal(t, "interop", actual.Links[0].Label)
	assert.Equal(t, 2, len(actual.Links[0].Results))
	assert.Equal(t, "b.html", actual.Links[0].Results[0].TestPath)
	assert.Equal(t, "abc.html", actual.Links[0].Results[1].TestPath)
}

func TestAddToFiles_AddNewMetadataLink_Asterisk(t *testing.T) {
	var amendment MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/foo1/*": [
			{
				"url": "https://foo1",
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
    url: https://foo
    results:
    - test: b.html
  - product: firefox
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
	assert.Equal(t, "https://foo", actual.Links[0].URL)
	assert.Equal(t, 1, len(actual.Links[0].Results))
	assert.Equal(t, "b.html", actual.Links[0].Results[0].TestPath)
	assert.Equal(t, "firefox", actual.Links[1].Product.BrowserName)
	assert.Equal(t, "https://bug.com/item", actual.Links[1].URL)
	assert.Equal(t, "chrome", actual.Links[2].Product.BrowserName)
	assert.Equal(t, "https://foo1", actual.Links[2].URL)
	assert.Equal(t, "*", actual.Links[2].Results[0].TestPath)
}

func TestNewTriageMetadata_email_fallback(t *testing.T) {
	// They won't be used in the constructor.
	ctx := context.Background()
	fetcher := MetadataFetcher(nil)

	m := NewTriageMetadata(ctx, nil, "testuser", "testemail@example.com", fetcher).(triageMetadata)
	assert.Equal(t, m.authorName, "testuser")
	assert.Equal(t, m.authorEmail, "testemail@example.com")

	m = NewTriageMetadata(ctx, nil, "testuser", "", fetcher).(triageMetadata)
	assert.Equal(t, m.authorName, "testuser")
	assert.Equal(t, m.authorEmail, "testuser@users.noreply.github.com")
}

func TestContainsInterop_True(t *testing.T) {
	var amendment MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/foo1/abc.html": [
			{
				"label": "interop-x"
			}
		]
	}`), &amendment)

	actual := containsInterop(amendment)

	assert.True(t, actual)
}

func TestContainsInterop_NotInteropLabel(t *testing.T) {
	var amendment MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/foo1/abc.html": [
			{
				"label": "lets-go-interoperability"
			}
		]
	}`), &amendment)

	actual := containsInterop(amendment)

	assert.False(t, actual)
}

func TestContainsInterop_False(t *testing.T) {
	var amendment MetadataResults
	json.Unmarshal([]byte(`{
		"/foo/foo1/*": [
			{
				"url": "https://foo1",
				"product": "chrome",
				"results": [
					{"status": 6 }
				]
			}
		]
	}`), &amendment)

	actual := containsInterop(amendment)

	assert.False(t, actual)
}

