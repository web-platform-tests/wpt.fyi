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

func TestMetadataHanlder_GET_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	r := httptest.NewRequest("GET", "/abd/api/metadata?product=firefox", nil)
	w := httptest.NewRecorder()
	sha := "sha"
	mockFetcher := sharedtest.NewMockMetadataFetcher(mockCtrl)
	mockFetcher.EXPECT().Fetch().Return(&sha, getMetadataTestData(), nil)

	metadataHandler := MetadataHandler{shared.NewNilLogger(), mockFetcher}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	res := w.Body.String()

	var expected, actual shared.MetadataResults
	json.Unmarshal([]byte(`{
        "/testB/b.html": [
            {
                "product": "firefox",
                "url":"bar.com",
                "results":[{"status":6}]
            }
        ]}`), &expected)
	json.Unmarshal([]byte(res), &actual)
	assert.Equal(t, expected, actual)
}

func TestMetadataHanlder_GET_MissingProducts(t *testing.T) {
	r := httptest.NewRequest("GET", "/abd/api/metadata?", nil)
	w := httptest.NewRecorder()

	metadataHandler := MetadataHandler{shared.NewNilLogger(), nil}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMetadataHandler_POST_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	body :=
		`{
        "exists": [{
            "link": "foo"
        }]
    }`
	bodyReader := strings.NewReader(body)
	r := httptest.NewRequest("POST", "/abd/api/metadata?product=chrome&product=safari", bodyReader)
	w := httptest.NewRecorder()

	sha := "shaA"
	mockFetcher := sharedtest.NewMockMetadataFetcher(mockCtrl)
	mockFetcher.EXPECT().Fetch().Return(&sha, getMetadataTestData(), nil)

	metadataHandler := MetadataHandler{shared.NewNilLogger(), mockFetcher}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	res := w.Body.String()

	var expected, actual shared.MetadataResults
	json.Unmarshal([]byte(`{
        "/root/testA/a.html": [
            {
                "product":"chrome",
                "url":"foo.com",
                "results":[
                    { "status":6 }
                ]
            }
        ]
    }`), &expected)
	json.Unmarshal([]byte(res), &actual)
	assert.EqualValues(t, expected, actual)
}

func TestMetadataHandler_POST_MissingProducts(t *testing.T) {
	body :=
		`{
        "exists": [{
            "link": "bugs.chromium.org"
        }]
    }`
	bodyReader := strings.NewReader(body)
	r := httptest.NewRequest("GET", "/abd/api/metadata?", bodyReader)
	w := httptest.NewRecorder()

	metadataHandler := MetadataHandler{shared.NewNilLogger(), nil}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMetadataHandler_POST_NotLink(t *testing.T) {
	body :=
		`{
        "exists": [{
            "pattern": "bugs.chromium.org"
        }]
    }`
	bodyReader := strings.NewReader(string(body))
	r := httptest.NewRequest("POST", "/abd/api/metadata?product=chrome&product=safari", bodyReader)
	w := httptest.NewRecorder()

	metadataHandler := MetadataHandler{shared.NewNilLogger(), nil}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMetadataHandler_POST_NotJustLink(t *testing.T) {
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

	metadataHandler := MetadataHandler{shared.NewNilLogger(), nil}
	metadataHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMetadata(t *testing.T) {
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

func getMetadataTestData() map[string][]byte {
	metadataMap := make(map[string][]byte)
	metadataMap["root/testA"] = []byte(`
    links:
      - product: chrome
        url: foo.com
        results:
        - test: a.html
          status: FAIL
    `)

	metadataMap["testB"] = []byte(`
    links:
      - product: firefox
        url: bar.com
        results:
        - test: b.html
          status: FAIL
    `)
	return metadataMap
}
