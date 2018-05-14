// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestShowResultsUploadForm_not_logged_in(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("GET", "/api/results/upload", new(strings.Reader))
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().isLoggedIn().Return(false)
	mockAE.EXPECT().loginURL("/api/results/upload").Return("/login", nil)

	ShowResultsUploadForm(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusTemporaryRedirect)
	assert.Equal(t, resp.Header().Get("Location"), "/login")
}

func TestShowResultsUploadForm_not_admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("GET", "/api/results/upload", new(strings.Reader))
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().isLoggedIn().Return(true)
	mockAE.EXPECT().isAdmin().Return(false)

	ShowResultsUploadForm(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusUnauthorized)
	assert.NotContains(t, resp.Body.String(), "form")
}

func TestShowResultsUploadForm_admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("GET", "/api/results/upload", new(strings.Reader))
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().isLoggedIn().Return(true)
	mockAE.EXPECT().isAdmin().Return(true)

	ShowResultsUploadForm(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Contains(t, resp.Body.String(), "form")
}

func TestHandleResultsUpload_not_admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/upload", new(strings.Reader))
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().isAdmin().Return(false)

	HandleResultsUpload(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusUnauthorized)
}

func TestHandleFilePayload(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	f := &os.File{}
	mockAE := NewMockAppEngineAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().uploadToGCS(gomock.Any(), f, true).Return("/blade-runner/test.json", nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "single"),
	)

	handleFilePayload(mockAE, "blade-runner", f)
}

func TestHandleURLPayload_single(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	f := &os.File{}
	mockAE := NewMockAppEngineAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().fetchURL("http://wpt.fyi/test.json.gz").Return(f, nil),
		mockAE.EXPECT().uploadToGCS(gomock.Any(), f, true).Return("/blade-runner/test.json", nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "single"),
	)

	handleURLPayload(mockAE, "blade-runner", []string{"http://wpt.fyi/test.json.gz"})
}

func TestHandleURLPayload_multiple(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	urls := []string{"http://wpt.fyi/foo.json.gz", "http://wpt.fyi/bar.json.gz"}
	gcs := []string{"/blade-runner/foo.json", "/blade-runner/bar.json"}

	f := &os.File{}
	mockAE := NewMockAppEngineAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().fetchURL(urls[0]).Return(f, nil),
		mockAE.EXPECT().uploadToGCS(gomock.Any(), f, true).Return(gcs[0], nil),
		mockAE.EXPECT().fetchURL(urls[1]).Return(f, nil),
		mockAE.EXPECT().uploadToGCS(gomock.Any(), f, true).Return(gcs[1], nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", gcs, "multiple"),
	)

	handleURLPayload(mockAE, "blade-runner", urls)
}
