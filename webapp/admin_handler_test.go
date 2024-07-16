// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestCheckAdmin_not_logged_in(t *testing.T) {
	resp := httptest.NewRecorder()
	assert.False(t, checkAdmin(nil, shared.NewNilLogger(), resp))
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestCheckAdmin_not_admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACL := sharedtest.NewMockGitHubAccessControl(mockCtrl)
	mockACL.EXPECT().IsValidAdmin().Return(false, nil)

	resp := httptest.NewRecorder()
	assert.False(t, checkAdmin(mockACL, shared.NewNilLogger(), resp))
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestCheckAdmin_error(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACL := sharedtest.NewMockGitHubAccessControl(mockCtrl)
	mockACL.EXPECT().IsValidAdmin().Return(true, errors.New("error"))

	resp := httptest.NewRecorder()
	assert.False(t, checkAdmin(mockACL, shared.NewNilLogger(), resp))
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestCheckAdmin_admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACL := sharedtest.NewMockGitHubAccessControl(mockCtrl)
	mockACL.EXPECT().IsValidAdmin().Return(true, nil)

	resp := httptest.NewRecorder()
	assert.True(t, checkAdmin(mockACL, shared.NewNilLogger(), resp))
	assert.Equal(t, http.StatusOK, resp.Code)
}
