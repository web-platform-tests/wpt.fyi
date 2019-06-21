// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestFilterMetadataHanlder_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../shared/metadata_testdata/gzip_testfile.tar.gz")
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	r := httptest.NewRequest("GET", "/abd/api/metadata?product=chrome&product=safari", nil)
	w := httptest.NewRecorder()
	client := server.Client()

	metadataHandler := MetadataHandler{shared.NewNilLogger(), client, server.URL}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	res := w.Body.String()

	assert.Equal(t, `[{"test":"/randomfolder1/innerfolder1/innerfolder2/innerfolder3/foo1.html","urls":["bugs.bar?id=456",""]},{"test":"/randomfolder2/foo.html","urls":["","safari.foo.com"]},{"test":"/randomfolder3/innerfolder1/random3foo.html","urls":["bugs.bar",""]}]`, res)
}

func TestFilterMetadataHanlder_MissingProducts(t *testing.T) {
	r := httptest.NewRequest("GET", "/abd/api/metadata?", nil)
	w := httptest.NewRecorder()

	metadataHandler := MetadataHandler{shared.NewNilLogger(), nil, ""}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFilterMetadataHandlerPost_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../shared/metadata_testdata/gzip_testfile.tar.gz")
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	body :=
		`{
		"exists": [{
			"link": "bugs.bar"
		}]
	}`
	bodyReader := strings.NewReader(body)
	r := httptest.NewRequest("POST", "/abd/api/metadata?product=chrome&product=safari", bodyReader)
	w := httptest.NewRecorder()
	client := server.Client()

	metadataHandler := MetadataHandler{shared.NewNilLogger(), client, server.URL}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	res := w.Body.String()

	assert.Equal(t, `[{"test":"/randomfolder1/innerfolder1/innerfolder2/innerfolder3/foo1.html","urls":["bugs.bar?id=456",""]},{"test":"/randomfolder3/innerfolder1/random3foo.html","urls":["bugs.bar",""]}]`, res)
}

func TestFilterMetadataHandlerPost_MissingProducts(t *testing.T) {
	body :=
		`{
		"exists": [{
			"link": "bugs.chromium.org"
		}]
	}`
	bodyReader := strings.NewReader(body)
	r := httptest.NewRequest("GET", "/abd/api/metadata?", bodyReader)
	w := httptest.NewRecorder()

	metadataHandler := MetadataHandler{shared.NewNilLogger(), nil, ""}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFilterMetadataHandlerPost_NotLink(t *testing.T) {
	body :=
		`{
		"exists": [{
			"pattern": "bugs.chromium.org"
		}]
	}`
	bodyReader := strings.NewReader(string(body))
	r := httptest.NewRequest("POST", "/abd/api/metadata?product=chrome&product=safari", bodyReader)
	w := httptest.NewRecorder()

	metadataHandler := MetadataHandler{shared.NewNilLogger(), nil, ""}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFilterMetadataHandlerPost_NotJustLink(t *testing.T) {
	body :=
		`{
		"exists": [{
			"and": [
				{"pattern": "bugs.chromium.org"},
				{"link": "abc"}
			]
		}]
	}`
	bodyReader := strings.NewReader(string(body))
	r := httptest.NewRequest("POST", "/abd/api/metadata?product=chrome&product=safari", bodyReader)
	w := httptest.NewRecorder()

	metadataHandler := MetadataHandler{shared.NewNilLogger(), nil, ""}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFilterMetadata(t *testing.T) {
	metadata := shared.MetadataResults(shared.MetadataResults{
		shared.MetadataResult{
			Test: "/foo/bar/b.html",
			URLs: []string{"", "https://aa.com/item", "https://bug.com/item"}},
		shared.MetadataResult{
			Test: "bar",
			URLs: []string{"", "https://external.com/item", ""}}})
	abstractLink := query.AbstractLink{Pattern: "bug.com"}

	res := filterMetadata(abstractLink, metadata)

	assert.Equal(t, 1, len(res))
	assert.Equal(t, "/foo/bar/b.html", res[0].Test)
	assert.Equal(t, "", res[0].URLs[0])
	assert.Equal(t, "https://aa.com/item", res[0].URLs[1])
	assert.Equal(t, "https://bug.com/item", res[0].URLs[2])
}
