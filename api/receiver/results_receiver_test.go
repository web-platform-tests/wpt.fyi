// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine/taskqueue"
)

func TestHandleResultsUpload_not_admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/upload", nil)
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().IsAdmin().Return(false)

	HandleResultsUpload(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusUnauthorized)
}

func TestHandleResultsUpload_http_basic_auth_invalid(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/upload", nil)
	req.SetBasicAuth("not_a_user", "123")
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().IsAdmin().Return(false),
		mockAE.EXPECT().AuthenticateUploader("not_a_user", "123").Return(false),
	)

	HandleResultsUpload(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusUnauthorized)
}

func TestHandleResultsUpload_success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	payload := url.Values{
		"result_url": {"http://wpt.fyi/test.json.gz"},
	}
	req := httptest.NewRequest("POST", "/api/results/upload", strings.NewReader(payload.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("blade-runner", "123")
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	f := &os.File{}
	task := &taskqueue.Task{Name: "task"}
	gomock.InOrder(
		mockAE.EXPECT().IsAdmin().Return(false),
		mockAE.EXPECT().AuthenticateUploader("blade-runner", "123").Return(true),
		mockAE.EXPECT().fetchURL("http://wpt.fyi/test.json.gz").Return(f, nil),
		mockAE.EXPECT().uploadToGCS(gomock.Any(), f, true).Return("/blade-runner/test.json", nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "single", gomock.Any()).Return(task, nil),
	)

	HandleResultsUpload(mockAE, resp, req)
	assert.Equal(t, resp.Code, http.StatusOK)
}

func TestHandleResultsUpload_extra_params(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	payload := url.Values{
		"result_url":    {"http://wpt.fyi/test.json.gz"},
		"browser_name":  {"firefox"},
		"labels":        {"stable"},
		"invalid_param": {"should be ignored"},
	}
	req := httptest.NewRequest("POST", "/api/results/upload", strings.NewReader(payload.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("blade-runner", "123")
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	f := &os.File{}
	extraParams := map[string]string{
		"browser_name":    "firefox",
		"labels":          "stable",
		"revision":        "",
		"browser_version": "",
		"os_name":         "",
		"os_version":      "",
	}
	task := &taskqueue.Task{Name: "task"}
	gomock.InOrder(
		mockAE.EXPECT().IsAdmin().Return(false),
		mockAE.EXPECT().AuthenticateUploader("blade-runner", "123").Return(true),
		mockAE.EXPECT().fetchURL("http://wpt.fyi/test.json.gz").Return(f, nil),
		mockAE.EXPECT().uploadToGCS(gomock.Any(), f, true).Return("/blade-runner/test.json", nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "single", extraParams).Return(task, nil),
	)

	HandleResultsUpload(mockAE, resp, req)
	assert.Equal(t, resp.Code, http.StatusOK)
}

func TestHandleFilePayload(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	f := &os.File{}
	extraParams := map[string]string{
		"browser_name": "firefox",
	}

	mockAE := NewMockAppEngineAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().uploadToGCS(gomock.Any(), f, true).Return("/blade-runner/test.json", nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "single", extraParams),
	)

	handleFilePayload(mockAE, "blade-runner", f, extraParams)
}

func TestHandleURLPayload_single(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	f := &os.File{}
	extraParams := map[string]string{
		"browser_name": "firefox",
	}

	mockAE := NewMockAppEngineAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().fetchURL("http://wpt.fyi/test.json.gz").Return(f, nil),
		mockAE.EXPECT().uploadToGCS(gomock.Any(), f, true).Return("/blade-runner/test.json", nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "single", extraParams),
	)

	handleURLPayload(mockAE, "blade-runner", []string{"http://wpt.fyi/test.json.gz"}, extraParams)
}

func TestHandleURLPayload_multiple(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	f := &os.File{}
	urls := []string{"http://wpt.fyi/foo.json.gz", "http://wpt.fyi/bar.json.gz"}
	gcs := []string{"/blade-runner/foo.json", "/blade-runner/bar.json"}

	mockAE := NewMockAppEngineAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().fetchURL(urls[0]).Return(f, nil),
		mockAE.EXPECT().uploadToGCS(gomock.Any(), f, true).Return(gcs[0], nil),
		mockAE.EXPECT().fetchURL(urls[1]).Return(f, nil),
		mockAE.EXPECT().uploadToGCS(gomock.Any(), f, true).Return(gcs[1], nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", gcs, "multiple", nil),
	)

	handleURLPayload(mockAE, "blade-runner", urls, nil)
}
