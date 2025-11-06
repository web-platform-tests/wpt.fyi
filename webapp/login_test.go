//go:build medium
// +build medium

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v77/github"
	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"go.uber.org/mock/gomock"
)

var (
	// This is both the blockkey and the hashkey for securecookie for *testing only*. It has 32 bytes.
	secretKey = "cd2a2650545fa9dc9f5aa265133a703a"
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
		(*token).Secret = secretKey
	})

	mockgo := sharedtest.NewMockGitHubOAuth(mockCtrl)
	mockgo.EXPECT().Context().AnyTimes().Return(ctx)
	mockgo.EXPECT().Datastore().AnyTimes().Return(mockStore)
	mockgo.EXPECT().SetRedirectURL("https://foo/oauth?return=%2F")
	mockgo.EXPECT().GetAuthCodeURL(gomock.Any(), gomock.Any()).Return("https://redirect?")

	handleLogin(mockgo, w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	cookies := resp.Header.Get("Set-Cookie")
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode, string(body))
	assert.Equal(t, "https://redirect?", resp.Header.Get("Location"))

	assert.True(t, strings.Contains(cookies, "state="))
	assert.True(t, strings.Contains(cookies, "Path=/"))
	assert.True(t, strings.Contains(cookies, "Max-Age=600"))
	assert.True(t, strings.Contains(cookies, "HttpOnly"))
	assert.True(t, strings.Contains(cookies, "HttpOnly"))
	assert.True(t, strings.Contains(cookies, "SameSite=Lax"))
}

func TestHandleOauth(t *testing.T) {
	state := "YZ6kSZ4PwwHMCcNHwd8xnd9u4ePzv9MmXrNNkYkPZ8Y"
	sc := securecookie.New([]byte(secretKey), []byte(secretKey))
	encodedState, err := sc.Encode("state", state)
	assert.Nil(t, err)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := sharedtest.NewTestContext()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("https://oauth?state=%s&code=bar", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  "state",
		Value: encodedState,
		Path:  "/",
	})

	mockStore := sharedtest.NewMockDatastore(mockCtrl)
	mockStore.EXPECT().NewNameKey("Token", gomock.Any()).AnyTimes().Return(nil)
	mockStore.EXPECT().Get(gomock.Any(), gomock.Any()).AnyTimes().Return(nil).Do(func(key shared.Key, dst interface{}) {
		token, _ := dst.(*shared.Token)
		(*token).Secret = secretKey
	})

	userName := "ufoo"
	userEmail := "ebar"
	secret := "token"
	dummyClient := &github.Client{}
	mockgo := sharedtest.NewMockGitHubOAuth(mockCtrl)
	mockgo.EXPECT().Context().AnyTimes().Return(ctx)
	mockgo.EXPECT().Datastore().AnyTimes().Return(mockStore)
	gomock.InOrder(
		mockgo.EXPECT().NewClient("bar").Return(dummyClient, nil),
		mockgo.EXPECT().GetUser(dummyClient).Return(&github.User{Login: &userName, Email: &userEmail}, nil),
		mockgo.EXPECT().GetAccessToken().Return(secret),
	)

	handleOauth(mockgo, w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	cookies := resp.Header.Get("Set-Cookie")
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode, string(body))
	assert.Equal(t, "/", resp.Header.Get("Location"))

	assert.True(t, strings.Contains(cookies, "session="))
	assert.True(t, strings.Contains(cookies, "Path=/"))
	assert.True(t, strings.Contains(cookies, "Max-Age=2592000"))
	assert.True(t, strings.Contains(cookies, "HttpOnly"))
	assert.True(t, strings.Contains(cookies, "HttpOnly"))
	assert.True(t, strings.Contains(cookies, "SameSite=None"))
}
