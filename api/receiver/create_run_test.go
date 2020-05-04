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
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/checks/mock_checks"
	"github.com/web-platform-tests/wpt.fyi/api/receiver/mock_receiver"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestHandleResultsCreate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := "0123456789012345678901234567890123456789"
	payload := map[string]interface{}{
		"id":                 12345,
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
	pAtR := shared.ProductAtRevision{
		Product: shared.Product{
			BrowserName:    "firefox",
			BrowserVersion: "59.0",
			OSName:         "linux",
			OSVersion:      "4.4",
		},
		Revision:         sha[:10],
		FullRevisionHash: sha,
	}
	testRunIn := &shared.TestRun{
		ID:                12345,
		TimeStart:         time.Date(2018, time.June, 21, 18, 39, 54, 218000000, time.UTC),
		TimeEnd:           time.Date(2018, time.June, 21, 20, 3, 49, 0, time.UTC),
		Labels:            []string{"foo", "bar"},
		ProductAtRevision: pAtR,
	}
	testKey := &sharedtest.MockKey{TypeName: "TestRun", ID: 12345}
	pendingRun := shared.PendingTestRun{
		ID:                12345,
		Stage:             shared.StageValid,
		ProductAtRevision: pAtR,
	}

	mockAE := mock_receiver.NewMockAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	mockS := mock_checks.NewMockAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().GetUploader("_processor").Return(shared.Uploader{"_processor", "secret-token"}, nil),
		mockAE.EXPECT().AddTestRun(sharedtest.SameProductSpec(testRunIn.String())).Return(testKey, nil),
		mockS.EXPECT().ScheduleResultsProcessing(sha, sharedtest.SameProductSpec("firefox")).Return(nil),
		mockAE.EXPECT().UpdatePendingTestRun(pendingRun).Return(nil),
	)

	w := httptest.NewRecorder()
	HandleResultsCreate(mockAE, mockS, w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var testRunOut shared.TestRun
	body, _ = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &testRunOut)
	assert.Nil(t, err)
	// Fields outside of ProductAtRevision are not included in the matcher, so check them now:
	assert.Equal(t, testRunIn.ID, testRunOut.ID)
	assert.Equal(t, testRunIn.Labels, testRunOut.Labels)
	assert.Equal(t, testRunIn.TimeStart, testRunOut.TimeStart)
	assert.Equal(t, testRunIn.TimeEnd, testRunOut.TimeEnd)
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
	pAtR := shared.ProductAtRevision{
		Product: shared.Product{
			BrowserName:    "firefox",
			BrowserVersion: "59.0",
			OSName:         "linux",
		},
		Revision:         sha[:10],
		FullRevisionHash: sha,
	}
	testRunIn := &shared.TestRun{ProductAtRevision: pAtR}
	testKey := &sharedtest.MockKey{TypeName: "TestRun", ID: 123}
	pendingRun := shared.PendingTestRun{
		ID:                123,
		Stage:             shared.StageValid,
		ProductAtRevision: pAtR,
	}

	mockAE := mock_receiver.NewMockAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	mockS := mock_checks.NewMockAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().GetUploader("_processor").Return(shared.Uploader{"_processor", "secret-token"}, nil),
		mockAE.EXPECT().AddTestRun(sharedtest.SameProductSpec(testRunIn.String())).Return(testKey, nil),
		mockS.EXPECT().ScheduleResultsProcessing(sha, sharedtest.SameProductSpec("firefox")).Return(nil),
		mockAE.EXPECT().UpdatePendingTestRun(pendingRun).Return(nil),
	)

	w := httptest.NewRecorder()
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
	mockS := mock_checks.NewMockAPI(mockCtrl)
	mockAE := mock_receiver.NewMockAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	mockAE.EXPECT().GetUploader("_processor").AnyTimes().Return(shared.Uploader{"_processor", "secret-token"}, nil)

	payload := map[string]interface{}{
		"browser_name":    "firefox",
		"browser_version": "59.0",
		"os_name":         "linux",
		"revision":        "0123456789",
	}
	body, err := json.Marshal(payload)
	assert.Nil(t, err)
	t.Run("Missing full_revision_hash", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/results/create", strings.NewReader(string(body)))
		req.SetBasicAuth("_processor", "secret-token")
		w := httptest.NewRecorder()

		HandleResultsCreate(mockAE, mockS, w, req)
		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	payload["full_revision_hash"] = "9876543210987654321098765432109876543210"
	body, err = json.Marshal(payload)
	assert.Nil(t, err)
	t.Run("Incorrect full_revision_hash", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/results/create", strings.NewReader(string(body)))
		req.SetBasicAuth("_processor", "secret-token")
		w := httptest.NewRecorder()

		HandleResultsCreate(mockAE, mockS, w, req)
		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandleResultsCreate_NoBasicAuth(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/create", nil)
	resp := httptest.NewRecorder()
	mockAE := mock_receiver.NewMockAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	mockS := mock_checks.NewMockAPI(mockCtrl)

	HandleResultsCreate(mockAE, mockS, resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestHandleResultsCreate_WrongUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/create", nil)
	req.SetBasicAuth("wrong-user", "secret-token")
	resp := httptest.NewRecorder()
	mockAE := mock_receiver.NewMockAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	mockAE.EXPECT().GetUploader("wrong-user").Return(shared.Uploader{"wrong-user", "secret-token"}, nil)
	mockS := mock_checks.NewMockAPI(mockCtrl)

	HandleResultsCreate(mockAE, mockS, resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestHandleResultsCreate_WrongPassword(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/create", nil)
	req.SetBasicAuth("_processor", "wrong-password")
	resp := httptest.NewRecorder()
	mockAE := mock_receiver.NewMockAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	mockAE.EXPECT().GetUploader("_processor").Return(shared.Uploader{"_processor", "secret-token"}, nil)
	mockS := mock_checks.NewMockAPI(mockCtrl)

	HandleResultsCreate(mockAE, mockS, resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}
