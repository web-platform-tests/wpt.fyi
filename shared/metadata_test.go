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
    test: a.html
    url: https://external.com/item
  - product: firefox-2
    test: b.html
    url: https://bug.com/item`)

	metadatamap := parseMetadata(metadataByteMap, NewNilLogger())

	assert.Equal(t, 1, len(metadatamap))
	assert.Equal(t, 2, len(metadatamap[path].Links))
	assert.Equal(t, "chrome", metadatamap[path].Links[0].Product.BrowserName)
	assert.Equal(t, "64", metadatamap[path].Links[0].Product.BrowserVersion)
	assert.Equal(t, "a.html", metadatamap[path].Links[0].TestPath)
	assert.Equal(t, "https://external.com/item", metadatamap[path].Links[0].URL)
	assert.Equal(t, "firefox", metadatamap[path].Links[1].Product.BrowserName)
	assert.Equal(t, "2", metadatamap[path].Links[1].Product.BrowserVersion)
	assert.Equal(t, "b.html", metadatamap[path].Links[1].TestPath)
	assert.Equal(t, "https://bug.com/item", metadatamap[path].Links[1].URL)
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
					Product:  ParseProductSpecUnsafe("ChrOme"),
					TestPath: "a.html",
					URL:      "https://external.com/item",
				},
				MetadataLink{
					Product:  ParseProductSpecUnsafe("Firefox"),
					TestPath: "a.html",
					URL:      "https://bug.com/item",
				},
			},
		},
	}

	MetadataResults := constructMetadataResponse(productSpecs, metadataMap)

	assert.Equal(t, 1, len(MetadataResults))
	assert.Equal(t, MetadataResults[0].Test, "/foo/bar/a.html")
	assert.Equal(t, MetadataResults[0].URLs[0], "https://bug.com/item")
	assert.Equal(t, MetadataResults[0].URLs[1], "https://external.com/item")
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
					Product:  ParseProductSpecUnsafe("ChrOme"),
					TestPath: "a.html",
					URL:      "https://external.com/item",
				},
				MetadataLink{
					Product:  ParseProductSpecUnsafe("safari"),
					TestPath: "a.html",
					URL:      "https://bug.com/item",
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
					Product:  ParseProductSpecUnsafe("ChrOme"),
					TestPath: "b.html",
					URL:      "https://external.com/item",
				},
				MetadataLink{
					Product:  ParseProductSpecUnsafe("Firefox"),
					TestPath: "a.html",
					URL:      "https://bug.com/item",
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
					Product:  ParseProductSpecUnsafe("ChrOme-2"),
					TestPath: "b.html",
					URL:      "https://external.com/item",
				},
				MetadataLink{
					Product:  ParseProductSpecUnsafe("Firefox-54"),
					TestPath: "a.html",
					URL:      "https://bug.com/item",
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
					Product:  ParseProductSpecUnsafe("ChrOme"),
					TestPath: "b.html",
					URL:      "https://external.com/item",
				},
				MetadataLink{
					Product:  ProductSpec{},
					TestPath: "a.html",
					URL:      "https://bug.com/item",
				},
			},
		},
	}

	MetadataResults := constructMetadataResponse(productSpecs, metadataMap)

	assert.Equal(t, 2, len(MetadataResults))
	assert.Equal(t, MetadataResults[0].Test, "/foo/bar/a.html")
	assert.Equal(t, MetadataResults[0].URLs[0], "https://bug.com/item")
	assert.Equal(t, MetadataResults[0].URLs[1], "https://bug.com/item")
	assert.Equal(t, MetadataResults[0].URLs[1], "https://bug.com/item")
	assert.Equal(t, MetadataResults[1].Test, "/foo/bar/b.html")
	assert.Equal(t, MetadataResults[1].URLs[0], "")
	assert.Equal(t, MetadataResults[1].URLs[1], "https://external.com/item")
	assert.Equal(t, MetadataResults[1].URLs[2], "")
}
