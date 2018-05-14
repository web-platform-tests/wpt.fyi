// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/receiver"
)

func TestShowAdminUploadForm_not_logged_in(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("GET", "/admin/results/upload", new(strings.Reader))
	resp := httptest.NewRecorder()
	mockAE := receiver.NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().IsLoggedIn().Return(false)
	mockAE.EXPECT().LoginURL("/admin/results/upload").Return("/login", nil)

	showAdminUploadForm(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusTemporaryRedirect)
	assert.Equal(t, resp.Header().Get("Location"), "/login")
}

func TestShowAdminUploadForm_not_admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("GET", "/admin/results/upload", new(strings.Reader))
	resp := httptest.NewRecorder()
	mockAE := receiver.NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().IsLoggedIn().Return(true)
	mockAE.EXPECT().IsAdmin().Return(false)

	showAdminUploadForm(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusUnauthorized)
	assert.NotContains(t, resp.Body.String(), "form")
}

func TestShowAdminUploadForm_admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("GET", "/admin/results/upload", new(strings.Reader))
	resp := httptest.NewRecorder()
	mockAE := receiver.NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().IsLoggedIn().Return(true)
	mockAE.EXPECT().IsAdmin().Return(true)

	showAdminUploadForm(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Contains(t, resp.Body.String(), "form")
}
