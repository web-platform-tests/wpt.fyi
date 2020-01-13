// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestHandleMetadataTriage_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := sharedtest.NewTestContext()
	w := httptest.NewRecorder()

	body :=
		`{
		"/bar/foo.html": [
			{
				"product":"chrome",
				"url":"bugs.bar",
				"results":[{"status":6}]
			}
		]}`
	bodyReader := strings.NewReader(body)
	req := httptest.NewRequest("PATCH", "https://foo/metadata", bodyReader)
	req.Header.Set("Content-Type", "application/json")

	mockgac := sharedtest.NewMockGitHubAccessControl(mockCtrl)
	mockgac.EXPECT().IsValidAccessToken().Return(http.StatusOK, nil)
	mockgac.EXPECT().IsValidWPTMember().Return(http.StatusOK, nil)

	mocktm := sharedtest.NewMockTriageMetadataInterface(mockCtrl)
	mocktm.EXPECT().Triage(gomock.Any()).Return("", nil)

	handleMetadataTriage(ctx, mockgac, mocktm, w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleMetadataTriage_NonSimpleRequests(t *testing.T) {
	w := httptest.NewRecorder()
	body :=
		`{
		"/bar/foo.html": [
			{
				"product":"chrome",
				"url":"bugs.bar",
				"results":[{"status":6}]
			}
		]}`
	bodyReader := strings.NewReader(body)
	req := httptest.NewRequest("GET", "https://foo/metadata", bodyReader)
	req.Header.Set("Content-Type", "application/json")

	handleMetadataTriage(nil, nil, nil, w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "https://foo/metadata", bodyReader)

	handleMetadataTriage(nil, nil, nil, w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleMetadataTriage_WrongContentType(t *testing.T) {
	w := httptest.NewRecorder()
	body :=
		`{
	"/bar/foo.html": [
		{
			"product":"chrome",
			"url":"bugs.bar",
			"results":[{"status":6}]
		}
	]}`
	bodyReader := strings.NewReader(body)
	req := httptest.NewRequest("PATCH", "https://foo/metadata", bodyReader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handleMetadataTriage(nil, nil, nil, w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleMetadataTriage_InvalidBody(t *testing.T) {
	w := httptest.NewRecorder()
	bodyReader := strings.NewReader("abc")
	req := httptest.NewRequest("PATCH", "https://foo/metadata", bodyReader)
	req.Header.Set("Content-Type", "application/json")

	handleMetadataTriage(nil, nil, nil, w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

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

	var expected, actual shared.MetadataResults
	json.Unmarshal([]byte(`{
		"/randomfolder1/innerfolder1/innerfolder2/innerfolder3/foo1.html": [
			{
				"product":"chrome",
				"url":"bugs.bar?id=456",
				"results":[
					{ "status":6 }
				]
			}
		],
		"/randomfolder2/foo.html": [
			{
				"product": "safari",
				"url":"safari.foo.com",
				"results":[{"status":0}]
			}
		],
		"/randomfolder3/innerfolder1/random3foo.html": [
			{
				"product":"chrome",
				"url":"bugs.bar",
				"results":[{"status":6}]
			}
		]}`), &expected)
	json.Unmarshal([]byte(res), &actual)
	assert.Equal(t, expected, actual)
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

	var expected, actual shared.MetadataResults
	json.Unmarshal([]byte(`{
		"/randomfolder1/innerfolder1/innerfolder2/innerfolder3/foo1.html": [
			{
				"url": "bugs.bar?id=456",
				"product": "chrome",
				"results": [
					{"status": 6 }
				]
			}
		],
		"/randomfolder3/innerfolder1/random3foo.html": [
			{
				"product": "chrome",
				"url": "bugs.bar",
				"results": [
					{"status": 6 }
				]}
		]
	}`), &expected)
	json.Unmarshal([]byte(res), &actual)
	assert.EqualValues(t, expected, actual)
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
		"/foo/bar/b.html": shared.MetadataLinks{
			shared.MetadataLink{
				URL: "https://aa.com/item",
			},
			shared.MetadataLink{
				URL: "https://bug.com/item",
			},
		},
		"bar": shared.MetadataLinks{
			shared.MetadataLink{
				URL: "https://external.com/item",
			},
		},
	})
	abstractLink := query.AbstractLink{Pattern: "bug.com"}

	res := filterMetadata(abstractLink, metadata)

	assert.Equal(t, 1, len(res))
	assert.Equal(t, "https://aa.com/item", res["/foo/bar/b.html"][0].URL)
}
