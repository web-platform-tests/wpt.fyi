//go:build small

// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package client

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"go.uber.org/mock/gomock"
)

func TestCreateRun(t *testing.T) {
	// To make sure we actually hit the assertion in the handler.
	visited := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "blade-runner", user)
		assert.Equal(t, "password", pass)
		assert.Nil(t, r.ParseForm())
		assert.Equal(t, []string{"https://wpt.fyi/results.json.gz"}, r.PostForm["result_url"])
		assert.Equal(t, []string{"https://wpt.fyi/screenshots.db.gz"}, r.PostForm["screenshot_url"])
		assert.Equal(t, "foo,bar", r.PostForm.Get("labels"))
		w.Write([]byte("OK"))
		visited = true
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()
	serverURL, _ := url.Parse(server.URL)

	mockC := gomock.NewController(t)
	defer mockC.Finish()
	aeAPI := sharedtest.NewMockAppEngineAPI(mockC)
	gomock.InOrder(
		aeAPI.EXPECT().GetVersionedHostname().Return("localhost:8080"),
		aeAPI.EXPECT().GetResultsUploadURL().Return(serverURL),
		aeAPI.EXPECT().GetHTTPClientWithTimeout(UploadTimeout).Return(server.Client()),
	)

	uc := NewClient(aeAPI)
	assert.Nil(t, uc.CreateRun(
		"abcdef1234abcdef1234abcdef1234abcdef1234",
		"blade-runner",
		"password",
		[]string{"https://wpt.fyi/results.json.gz"},
		[]string{"https://wpt.fyi/screenshots.db.gz"},
		nil,
		[]string{"foo", "bar"},
	))
	assert.True(t, visited)
}
