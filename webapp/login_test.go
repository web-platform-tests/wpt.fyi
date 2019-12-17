// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"github.com/web-platform-tests/wpt.fyi/webapp/mock_webapp"
)

func TestHandleLogin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockgo := mock_webapp.NewMockGithubOAuth(mockCtrl)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/login", nil)
	mockgo.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())

	handleLogin(mockgo, w, req)

}
