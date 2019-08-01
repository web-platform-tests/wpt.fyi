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
	assert.Equal(t, "Something should happen", metadatamap[path].Links[1].Results[0].SubtestName)
	assert.Equal(t, TestStatusFail, metadatamap[path].Links[1].Results[0].Status)
	assert.Equal(t, "https://bug.com/item", metadatamap[path].Links[1].URL)
	assert.Len(t, metadatamap[path].Links[1].Results, 2)
	assert.Equal(t, "b.html", metadatamap[path].Links[1].Results[0].TestPath)
	assert.Equal(t, "Something should happen", metadatamap[path].Links[1].Results[0].SubtestName)
	assert.Equal(t, TestStatusFail, metadatamap[path].Links[1].Results[0].Status)
}

func TestConstructMetadataResponse_OneLink(t *testing.T) {
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
						TestPath: "a.html",
					}},
				},
				MetadataLink{
					Product: ParseProductSpecUnsafe("Firefox"),
					URL:     "https://bug.com/item",
					Results: []MetadataTestResult{{
						TestPath:    "a.html",
						SubtestName: "subtest-a",
						Status:      TestStatusFail,
					}},
				},
			},
		},
	}

	MetadataResults := constructMetadataResponse(productSpecs, metadataMap)

	assert.Equal(t, 2, len(MetadataResults))
	assert.Equal(t, MetadataResults[0].Test, "/foo/bar/a.html")
	assert.Equal(t, MetadataResults[0].Subtest, "")
	assert.Equal(t, MetadataResults[0].URLs[0], "")
	assert.Equal(t, MetadataResults[0].URLs[1], "https://external.com/item")
	assert.Equal(t, MetadataResults[0].Status[0], TestStatus(0))
	assert.Equal(t, MetadataResults[0].Status[1], TestStatus(0))
	assert.Equal(t, MetadataResults[1].Test, "/foo/bar/a.html")
	assert.Equal(t, MetadataResults[1].Subtest, "subtest-a")
	assert.Equal(t, MetadataResults[1].URLs[0], "https://bug.com/item")
	assert.Equal(t, MetadataResults[1].URLs[1], "")
	assert.Equal(t, MetadataResults[1].Status[0], TestStatusFail)
	assert.Equal(t, MetadataResults[1].Status[1], TestStatus(0))
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
	assert.Equal(t, MetadataResults[0].Test, "/foo/bar/a.html")
	assert.Equal(t, MetadataResults[0].URLs[0], "https://bug.com/item")
	assert.Equal(t, MetadataResults[0].URLs[1], "")
	assert.Equal(t, MetadataResults[1].Test, "/foo/bar/b.html")
	assert.Equal(t, MetadataResults[1].URLs[0], "")
	assert.Equal(t, MetadataResults[1].URLs[1], "https://external.com/item")
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
	assert.Equal(t, MetadataResults[0].Test, "/foo/bar/a.html")
	assert.Equal(t, MetadataResults[0].URLs[0], "https://bug.com/item")
	assert.Equal(t, MetadataResults[0].URLs[1], "")
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
						SubtestName: "subtest-a",
						Status: TestStatusFail,
					}},
				},
				MetadataLink{
					Product: ProductSpec{},
					URL:     "https://bug.com/item",
					Results: []MetadataTestResult{{
						TestPath: "a.html",
					}},
				},
				MetadataLink{
					Product: ParseProductSpecUnsafe("safari"),
					URL:     "https:/apple/item",
					Results: []MetadataTestResult{{
						TestPath: "b.html",
						SubtestName: "subtest-a",
						Status: TestStatusCrash,
					}},
				},
			},
		},
	}

	MetadataResults := constructMetadataResponse(productSpecs, metadataMap)

	assert.Equal(t, 2, len(MetadataResults))
	assert.Equal(t, MetadataResults[0].Test, "/foo/bar/a.html")
	assert.Equal(t, MetadataResults[0].Subtest, "")
	assert.Equal(t, MetadataResults[0].URLs[0], "https://bug.com/item")
	assert.Equal(t, MetadataResults[0].URLs[1], "https://bug.com/item")
	assert.Equal(t, MetadataResults[0].URLs[1], "https://bug.com/item")
	assert.Equal(t, MetadataResults[1].Test, "/foo/bar/b.html")
	assert.Equal(t, MetadataResults[1].Subtest, "subtest-a")
	assert.Equal(t, MetadataResults[1].URLs[0], "")
	assert.Equal(t, MetadataResults[1].URLs[1], "https://external.com/item")
	assert.Equal(t, MetadataResults[1].URLs[2], "https:/apple/item")
	assert.Equal(t, MetadataResults[1].Status[0], TestStatus(0))
	assert.Equal(t, MetadataResults[1].Status[1], TestStatusFail)
	assert.Equal(t, MetadataResults[1].Status[2], TestStatusCrash)
}
