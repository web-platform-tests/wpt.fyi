// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestHandleResultsCreate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	payload := map[string]interface{}{
		"browser_name":    "firefox",
		"browser_version": "59.0",
		"os_name":         "linux",
		"os_version":      "4.4",
		"revision":        "0123456789",
		"labels":          []string{"foo", "bar"},
		"time_start":      "2018-06-21T18:39:54.218000+00:00",
		"time_end":        "2018-06-21T20:03:49Z",
		// Intentionally missing full_revision_hash; no error should be raised.
		// Unknown parameters should be ignored.
		"_random_extra_key_": "some_value",
	}
	body, err := json.Marshal(payload)
	assert.Nil(t, err)
	req := httptest.NewRequest("POST", "/api/results/create", strings.NewReader(string(body)))
	req.SetBasicAuth("_processor", "secret-token")
	w := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().AuthenticateUploader("_processor", "secret-token").Return(true),
		mockAE.EXPECT().AddTestRun(gomock.Any()).Return(nil, nil),
	)

	HandleResultsCreate(mockAE, w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var testRun shared.TestRun
	body, _ = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &testRun)
	assert.Nil(t, err)
	assert.Equal(t, "firefox", testRun.BrowserName)
	assert.Equal(t, []string{"foo", "bar"}, testRun.Labels)
	assert.False(t, testRun.TimeStart.IsZero())
	assert.False(t, testRun.TimeEnd.IsZero())
}

func TestHandleResultsCreate_NoTimestamps(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	payload := map[string]interface{}{
		"browser_name":    "firefox",
		"browser_version": "59.0",
		"os_name":         "linux",
		"revision":        "0123456789",
	}
	body, err := json.Marshal(payload)
	assert.Nil(t, err)
	req := httptest.NewRequest("POST", "/api/results/create", strings.NewReader(string(body)))
	req.SetBasicAuth("_processor", "secret-token")
	w := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().AuthenticateUploader("_processor", "secret-token").Return(true),
		mockAE.EXPECT().AddTestRun(gomock.Any()).Return(nil, nil),
	)

	HandleResultsCreate(mockAE, w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var testRun shared.TestRun
	body, _ = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &testRun)
	assert.Nil(t, err)
	assert.False(t, testRun.CreatedAt.IsZero())
	assert.False(t, testRun.TimeStart.IsZero())
	assert.Equal(t, testRun.TimeStart, testRun.TimeEnd)
}

func TestHandleResultsCreate_NoBasicAuth(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/create", nil)
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)

	HandleResultsCreate(mockAE, resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestHandleResultsCreate_WrongUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/create", nil)
	req.SetBasicAuth("wrong-user", "secret-token")
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)

	HandleResultsCreate(mockAE, resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestHandleResultsCreate_WrongPassword(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/create", nil)
	req.SetBasicAuth("_processor", "wrong-password")
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().AuthenticateUploader("_processor", "wrong-password").Return(false)

	HandleResultsCreate(mockAE, resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}
