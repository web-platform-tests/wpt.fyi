// +build medium
// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"github.com/web-platform-tests/wpt.fyi/webapp/mock_webapp"
)

func TestHandleLogin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := sharedtest.NewTestContext()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "https://foo/login", nil)

	mockStore := sharedtest.NewMockDatastore(mockCtrl)
	mockStore.EXPECT().NewNameKey("Token", gomock.Any()).AnyTimes().Return(nil)
	mockStore.EXPECT().Get(gomock.Any(), gomock.Any()).AnyTimes().Return(nil).Do(func(key shared.Key, dst interface{}) {
		token, _ := dst.(*shared.Token)
		// Need to create a 16 bytes fake secrete for securecookie.
		sc, _ := generateRandomState(16)
		(*token).Secret = sc
	})

	mockgo := mock_webapp.NewMockGithubOAuth(mockCtrl)
	mockgo.EXPECT().Context().AnyTimes().Return(ctx)
	mockgo.EXPECT().SetRedirectURL("https://foo/oauth?return=%2F").Return()
	mockgo.EXPECT().GetAuthCodeURL(gomock.Any(), gomock.Any()).Return("foo")
	mockgo.EXPECT().Datastore().AnyTimes().Return(mockStore)

	handleLogin(mockgo, w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	assert.Equal(t, nil, w.Header().Get("Set-Cookie"))
}
