// +build small

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

func TestHandleResultsUpload_not_admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/upload", new(strings.Reader))
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().IsAdmin().Return(false)

	HandleResultsUpload(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusUnauthorized)
}

func TestHandleResultsUpload_http_basic_auth_invalid(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/upload", new(strings.Reader))
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
