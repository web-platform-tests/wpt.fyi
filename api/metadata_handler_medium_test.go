// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"testing"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)


func TestFilterMetadataHanlder(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "metadata_testdata/gzip_testfile.tar.gz")
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	r := httptest.NewRequest("GET", "/abd/api/metadata?product=chrome&product=safari", nil)
	w := httptest.NewRecorder()
	client := &http.Client{}

	metadataHandler := MetadataHandler{nil, client, server.URL}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	res := w.Body.String()
	assert.Equal(t, "[{\"test\":\"/IndexedDB/bindings-inject-key.html\",\"urls\":[\"bugs.chromium.org/p/chromium/issues/detail?id=934844\",\"\"]},{\"test\":\"/html/browsers/history/the-history-interface/007.html\",\"urls\":[\"bugs.chromium.org/p/chromium/issues/detail?id=592874\",\"\"]}]", res)
}

func TestFilterMetadata(t *testing.T) {
	metadata := shared.MetadataResults(shared.MetadataResults{shared.MetadataResult{Test: "/foo/bar/b.html", URLs: []string{"", "https://aa.com/item", "https://bug.com/item"}}, shared.MetadataResult{Test: "bar", URLs: []string{"", "https://external.com/item", ""}}})
	abstractLink := query.AbstractLink{Pattern: "bug.com"}

	res := filterMetadata(abstractLink, metadata)

	assert.Equal(t, 1, len(res))
	assert.Equal(t, "/foo/bar/b.html", res[0].Test)
	assert.Equal(t, "", res[0].URLs[0])
	assert.Equal(t, "https://aa.com/item", res[0].URLs[1])
	assert.Equal(t, "https://bug.com/item", res[0].URLs[2])
}
