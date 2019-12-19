// +build medium
// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v28/github"
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
		sc := "dIS6V5HAQppr4QyLCSTEyg=="
		(*token).Secret = sc
	})

	mockgo := mock_webapp.NewMockGithubOAuth(mockCtrl)
	mockgo.EXPECT().Context().AnyTimes().Return(ctx)
	mockgo.EXPECT().SetRedirectURL("https://foo/oauth?return=%2F").Return()
	mockgo.EXPECT().GetAuthCodeURL(gomock.Any(), gomock.Any()).Return("https://redirect?")
	mockgo.EXPECT().Datastore().AnyTimes().Return(mockStore)

	handleLogin(mockgo, w, req)

	resp := w.Result()
	cookies := w.Header().Get("Set-Cookie")
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	assert.Equal(t, "https://redirect?", w.Header().Get("Location"))

	assert.True(t, strings.Contains(cookies, "state="))
	assert.True(t, strings.Contains(cookies, "Path=/"))
	assert.True(t, strings.Contains(cookies, "Max-Age=600"))
	assert.True(t, strings.Contains(cookies, "HttpOnly"))
	assert.True(t, strings.Contains(cookies, "HttpOnly"))
	assert.True(t, strings.Contains(cookies, "SameSite=Lax"))
}

func TestHandleOauth(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := sharedtest.NewTestContext()
	w := httptest.NewRecorder()
	sessionVal := "MTU3Njc4MDA0N3xpSXVpeXJyWFRFZUZBRTBDdGVISG00Q1h3YWlqM3ZmdkFORmk1Z1pWTFQtd24yQjlXM0ZsTkF5Ti1sS1ozZS1EWjZub1dSRXVyMHd5TnprZV9ZeHF0Zz09fNRqZ-8jwop8p39BpJTXlpNrRsfMeWMTH4CuRfA0QS0e"
	req := httptest.NewRequest("GET", "https://oauth?state=YZ6kSZ4PwwHMCcNHwd8xnd9u4ePzv9MmXrNNkYkPZ8Y=&code=bar", nil)
	req.AddCookie(&http.Cookie{
		Name:  "state",
		Value: sessionVal,
		Path:  "/",
	})

	mockStore := sharedtest.NewMockDatastore(mockCtrl)
	mockStore.EXPECT().NewNameKey("Token", gomock.Any()).AnyTimes().Return(nil)
	mockStore.EXPECT().Get(gomock.Any(), gomock.Any()).AnyTimes().Return(nil).Do(func(key shared.Key, dst interface{}) {
		token, _ := dst.(*shared.Token)
		// Need to create a 16 bytes fake secrete for securecookie.
		sc := "dIS6V5HAQppr4QyLCSTEyg=="
		(*token).Secret = sc
	})

	userName := "ufoo"
	userEmail := "ebar"
	secrete := "token"
	mockgo := mock_webapp.NewMockGithubOAuth(mockCtrl)
	mockgo.EXPECT().Context().AnyTimes().Return(ctx)
	mockgo.EXPECT().GetNewClient(gomock.Any()).Return(nil, nil)
	mockgo.EXPECT().GetGithubUser(gomock.Any()).Return(&github.User{Login: &userName, Email: &userEmail}, nil)
	mockgo.EXPECT().Datastore().AnyTimes().Return(mockStore)
	mockgo.EXPECT().GetAccessToken().Return(&secrete)

	handleOauth(mockgo, w, req)

	resp := w.Result()
	cookies := w.Header().Get("Set-Cookie")
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	assert.Equal(t, "/", w.Header().Get("Location"))

	assert.True(t, strings.Contains(cookies, "session="))
	assert.True(t, strings.Contains(cookies, "Path=/"))
	assert.True(t, strings.Contains(cookies, "Max-Age=2592000"))
	assert.True(t, strings.Contains(cookies, "HttpOnly"))
	assert.True(t, strings.Contains(cookies, "HttpOnly"))
	assert.True(t, strings.Contains(cookies, "SameSite=None"))
}
