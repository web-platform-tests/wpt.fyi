// +build small

// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"compress/gzip"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMetadataFromGZip_Success(t *testing.T) {
	f, _ := os.Open("metadata_testdata/util_gzip_testfile.tar.gz")
	defer f.Close()
	reader, _ := gzip.NewReader(f)
	expectedValIndexedDB :=
		`links:
  - product: chrome
    test: bindings-inject-key.html
    status: MISSING
    url: bugs.chromium.org/p/chromium/issues/detail?id=934844
`

	expectedValTheHistoryInterface :=
		`links:
  - product: chrome
    test: 007.html
    status: FAIL
    url: bugs.chromium.org/p/chromium/issues/detail?id=592874
`
	metadataMapRes, err := parseMetadataFromGZip(reader)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(metadataMapRes))
	val, exist := metadataMapRes["IndexedDB"]
	assert.True(t, exist)
	assert.Equal(t, expectedValIndexedDB, string(val))
	val, exist = metadataMapRes["html/browsers/history/the-history-interface"]
	assert.True(t, exist)
	assert.Equal(t, expectedValTheHistoryInterface, string(val))
}
