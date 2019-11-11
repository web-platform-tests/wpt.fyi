// +build small

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMetadata(t *testing.T) {
	var path = "foo/bar"
	var metadataByteMap = make(map[string][]byte)
	metadataByteMap[path] = []byte(`
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

	metadatamap := parseMetadata(metadataByteMap, NewNilLogger())

	assert.Len(t, metadatamap, 1)
	assert.Len(t, metadatamap[path].Links, 2)
	assert.Equal(t, "chrome", metadatamap[path].Links[0].Product.BrowserName)
	assert.Equal(t, "64", metadatamap[path].Links[0].Product.BrowserVersion)
	assert.Equal(t, "a.html", metadatamap[path].Links[0].Results[0].TestPath)
	assert.Equal(t, "https://external.com/item", metadatamap[path].Links[0].URL)
	assert.Equal(t, "firefox", metadatamap[path].Links[1].Product.BrowserName)
	assert.Equal(t, "2", metadatamap[path].Links[1].Product.BrowserVersion)
	assert.Equal(t, "b.html", metadatamap[path].Links[1].Results[0].TestPath)
	assert.Equal(t, "Something should happen", *(metadatamap[path].Links[1].Results[0].SubtestName))
	assert.Equal(t, TestStatusFail, *(metadatamap[path].Links[1].Results[0].Status))
	assert.Equal(t, "https://bug.com/item", metadatamap[path].Links[1].URL)
	assert.Len(t, metadatamap[path].Links[1].Results, 2)
	assert.Equal(t, "b.html", metadatamap[path].Links[1].Results[0].TestPath)
	assert.Equal(t, "Something should happen", *(metadatamap[path].Links[1].Results[0].SubtestName))
	assert.Equal(t, TestStatusFail, *(metadatamap[path].Links[1].Results[0].Status))
}

func TestConstructMetadataResponse_OneLink(t *testing.T) {
	productSpecs := []ProductSpec{
		ParseProductSpecUnsafe("Firefox-54"),
		ParseProductSpecUnsafe("Chrome"),
	}
	subtestName := "Something should happen"
	fail := TestStatusFail
	metadataMap := map[string]Metadata{
		"foo/bar": Metadata{
			Links: []MetadataLink{
				MetadataLink{
					Product: ParseProductSpecUnsafe("ChrOme"),
					URL:     "https://external.com/item",
					Results: []MetadataTestResult{{
						TestPath: "a.html",
					}},
				},
				MetadataLink{
					Product: ParseProductSpecUnsafe("Firefox"),
					URL:     "https://bug.com/item",
					Results: []MetadataTestResult{{
						TestPath:    "a.html",
						SubtestName: &subtestName,
						Status:      &fail,
					}},
				},
			},
		},
	}

	MetadataResults := constructMetadataResponse(productSpecs, metadataMap)

	assert.Equal(t, 1, len(MetadataResults))
	assert.Equal(t, 2, len(MetadataResults["/foo/bar/a.html"]))
	assert.Equal(t, "https://external.com/item", MetadataResults["/foo/bar/a.html"][0].URL)
	assert.True(t, ParseProductSpecUnsafe("chrome").MatchesProductSpec(MetadataResults["/foo/bar/a.html"][0].Product))
	assert.Equal(t, "https://bug.com/item", MetadataResults["/foo/bar/a.html"][1].URL)
	assert.True(t, ParseProductSpecUnsafe("firefox").MatchesProductSpec(MetadataResults["/foo/bar/a.html"][1].Product))
}

func TestConstructMetadataResponse_NoMatchingLink(t *testing.T) {
	productSpecs := []ProductSpec{
		ParseProductSpecUnsafe("Firefox-54"),
		ParseProductSpecUnsafe("Firefox"),
	}
	metadataMap := map[string]Metadata{
		"foo/bar": Metadata{
			Links: []MetadataLink{
				MetadataLink{
					Product: ParseProductSpecUnsafe("ChrOme"),
					URL:     "https://external.com/item",
					Results: []MetadataTestResult{{
						TestPath: "a.html",
					}},
				},
				MetadataLink{
					Product: ParseProductSpecUnsafe("safari"),
					URL:     "https://bug.com/item",
					Results: []MetadataTestResult{{
						TestPath: "a.html",
					}},
				},
			},
		},
	}

	MetadataResults := constructMetadataResponse(productSpecs, metadataMap)

	assert.Equal(t, 0, len(MetadataResults))
}

func TestConstructMetadataResponse_MultipleLinks(t *testing.T) {
	productSpecs := []ProductSpec{
		ParseProductSpecUnsafe("Firefox-54"),
		ParseProductSpecUnsafe("Chrome"),
	}
	metadataMap := map[string]Metadata{
		"foo/bar": Metadata{
			Links: []MetadataLink{
				MetadataLink{
					Product: ParseProductSpecUnsafe("ChrOme"),
					URL:     "https://external.com/item",
					Results: []MetadataTestResult{{
						TestPath: "b.html",
					}},
				},
				MetadataLink{
					Product: ParseProductSpecUnsafe("Firefox"),
					URL:     "https://bug.com/item",
					Results: []MetadataTestResult{{
						TestPath: "a.html",
					}},
				},
			},
		},
	}

	MetadataResults := constructMetadataResponse(productSpecs, metadataMap)

	assert.Equal(t, 2, len(MetadataResults))
	assert.Equal(t, MetadataResults["/foo/bar/a.html"][0].URL, "https://bug.com/item")
	assert.Equal(t, MetadataResults["/foo/bar/b.html"][0].URL, "https://external.com/item")
}

func TestConstructMetadataResponse_OneMatchingBrowserVersion(t *testing.T) {
	productSpecs := []ProductSpec{
		ParseProductSpecUnsafe("Firefox-54"),
		ParseProductSpecUnsafe("Chrome-1"),
	}
	metadataMap := map[string]Metadata{
		"foo/bar": Metadata{
			Links: []MetadataLink{
				MetadataLink{
					Product: ParseProductSpecUnsafe("ChrOme-2"),
					URL:     "https://external.com/item",
					Results: []MetadataTestResult{{
						TestPath: "b.html",
					}},
				},
				MetadataLink{
					Product: ParseProductSpecUnsafe("Firefox-54"),
					URL:     "https://bug.com/item",
					Results: []MetadataTestResult{{
						TestPath: "a.html",
					}},
				},
			},
		},
	}

	MetadataResults := constructMetadataResponse(productSpecs, metadataMap)

	assert.Equal(t, 1, len(MetadataResults))
	assert.Equal(t, MetadataResults["/foo/bar/a.html"][0].URL, "https://bug.com/item")
}

func TestConstructMetadataResponse_WithEmptyProductSpec(t *testing.T) {
	productSpecs := []ProductSpec{
		ParseProductSpecUnsafe("Firefox-54"),
		ParseProductSpecUnsafe("Chrome"),
		ParseProductSpecUnsafe("Safari"),
	}
	metadataMap := map[string]Metadata{
		"foo/bar": Metadata{
			Links: []MetadataLink{
				MetadataLink{
					Product: ParseProductSpecUnsafe("ChrOme"),
					URL:     "https://external.com/item",
					Results: []MetadataTestResult{{
						TestPath: "b.html",
					}},
				},
				MetadataLink{
					Product: ProductSpec{},
					URL:     "https://bug.com/item",
					Results: []MetadataTestResult{{
						TestPath: "a.html",
					}},
				},
			},
		},
	}

	MetadataResults := constructMetadataResponse(productSpecs, metadataMap)

	assert.Equal(t, 2, len(MetadataResults))
	assert.Equal(t, 1, len(MetadataResults["/foo/bar/a.html"]))
	assert.Equal(t, MetadataResults["/foo/bar/a.html"][0].URL, "https://bug.com/item")
	assert.Equal(t, 1, len(MetadataResults["/foo/bar/b.html"]))
	assert.Equal(t, MetadataResults["/foo/bar/b.html"][0].URL, "https://external.com/item")
}

func TestGetWPTTestPath(t *testing.T) {
	actual := GetWPTTestPath("foo", "bar")
	assert.Equal(t, "/foo/bar", actual)
}

func TestGetWPTTestPath_EmptyFolder(t *testing.T) {
	actual := GetWPTTestPath("", "bar")
	assert.Equal(t, "/bar", actual)
}

func TestSplitWPTTestPath_InvalidPath(t *testing.T) {
	folderPath, testPath := SplitWPTTestPath("foo/bar")
	assert.Equal(t, "", folderPath)
	assert.Equal(t, "", testPath)
}

func TestSplitWPTTestPath_EmptyPath(t *testing.T) {
	folderPath, testPath := SplitWPTTestPath("/")
	assert.Equal(t, "", folderPath)
	assert.Equal(t, "", testPath)
}

func TestSplitWPTTestPath_NoFolderPath(t *testing.T) {
	folderPath, testPath := SplitWPTTestPath("/foo")
	assert.Equal(t, "", folderPath)
	assert.Equal(t, "foo", testPath)
}

func TestSplitWPTTestPath_Success(t *testing.T) {
	folderPath, testPath := SplitWPTTestPath("/foo/bar/foo1")
	assert.Equal(t, "foo/bar", folderPath)
	assert.Equal(t, "foo1", testPath)

	folderPath, testPath = SplitWPTTestPath("/foo/bar")
	assert.Equal(t, "foo", folderPath)
	assert.Equal(t, "bar", testPath)
}

func TestGetMetadataFilePath(t *testing.T) {
	actual := GetMetadataFilePath("foo")
	assert.Equal(t, "foo/META.yml", actual)
}
