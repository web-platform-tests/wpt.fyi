// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/checks"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

var testDatastoreKey = &DatastoreKey{"TestRun", 1}

func TestHandleResultsCreate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := "0123456789012345678901234567890123456789"
	payload := map[string]interface{}{
		"browser_name":       "firefox",
		"browser_version":    "59.0",
		"os_name":            "linux",
		"os_version":         "4.4",
		"revision":           sha[:10],
		"full_revision_hash": sha,
		"labels":             []string{"foo", "bar"},
		"time_start":         "2018-06-21T18:39:54.218000+00:00",
		"time_end":           "2018-06-21T20:03:49Z",
		"_random_extra_key_": "some_value",
	}
	body, err := json.Marshal(payload)
	assert.Nil(t, err)
	req := httptest.NewRequest("POST", "/api/results/create", strings.NewReader(string(body)))
	req.SetBasicAuth("_processor", "secret-token")
	w := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(context.Background())
	mockS := checks.NewMockAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().authenticateUploader("_processor", "secret-token").Return(true),
		mockAE.EXPECT().addTestRun(gomock.Any()).Return(testDatastoreKey, nil),
		mockS.EXPECT().ScheduleResultsProcessing(sha, sharedtest.SameProductSpec("firefox")).Return(nil),
	)

	HandleResultsCreate(mockAE, mockS, w, req)
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

	sha := "0123456789012345678901234567890123456789"
	payload := map[string]interface{}{
		"browser_name":       "firefox",
		"browser_version":    "59.0",
		"os_name":            "linux",
		"revision":           sha[:10],
		"full_revision_hash": sha,
	}
	body, err := json.Marshal(payload)
	assert.Nil(t, err)
	req := httptest.NewRequest("POST", "/api/results/create", strings.NewReader(string(body)))
	req.SetBasicAuth("_processor", "secret-token")
	w := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(context.Background())
	mockS := checks.NewMockAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().authenticateUploader("_processor", "secret-token").Return(true),
		mockAE.EXPECT().addTestRun(gomock.Any()).Return(testDatastoreKey, nil),
		mockS.EXPECT().ScheduleResultsProcessing(sha, sharedtest.SameProductSpec("firefox")).Return(nil),
	)

	HandleResultsCreate(mockAE, mockS, w, req)
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

func TestHandleResultsCreate_BadRevision(t *testing.T) {
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
	mockAE.EXPECT().Context().AnyTimes().Return(context.Background())
	mockS := checks.NewMockAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().authenticateUploader("_processor", "secret-token").Return(true),
	)

	HandleResultsCreate(mockAE, mockS, w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	payload["full_revision_hash"] = "9876543210987654321098765432109876543210"
	gomock.InOrder(
		mockAE.EXPECT().authenticateUploader("_processor", "secret-token").Return(true),
	)
	body, err = json.Marshal(payload)
	assert.Nil(t, err)
	req = httptest.NewRequest("POST", "/api/results/create", strings.NewReader(string(body)))
	req.SetBasicAuth("_processor", "secret-token")

	HandleResultsCreate(mockAE, mockS, w, req)
	resp = w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleResultsCreate_NoBasicAuth(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/create", nil)
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(context.Background())
	mockS := checks.NewMockAPI(mockCtrl)

	HandleResultsCreate(mockAE, mockS, resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestHandleResultsCreate_WrongUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/create", nil)
	req.SetBasicAuth("wrong-user", "secret-token")
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(context.Background())
	mockS := checks.NewMockAPI(mockCtrl)

	HandleResultsCreate(mockAE, mockS, resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestHandleResultsCreate_WrongPassword(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/create", nil)
	req.SetBasicAuth("_processor", "wrong-password")
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(context.Background())
	mockAE.EXPECT().authenticateUploader("_processor", "wrong-password").Return(false)
	mockS := checks.NewMockAPI(mockCtrl)

	HandleResultsCreate(mockAE, mockS, resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}
