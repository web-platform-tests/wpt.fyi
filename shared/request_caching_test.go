// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

type failReader struct{}
type okHandler struct{}

var (
	errFailRead = errors.New("Failed read")
	ok          = []byte("OK")
)

func (failReader) Read([]byte) (int, error) { return 0, errFailRead }

func (okHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write(ok)
}

func TestNoCaching404(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	cache := sharedtest.NewMockReadWritable(mockCtrl)
	cache.EXPECT().NewReadCloser("/some/url").Return(ioutil.NopCloser(failReader{}), nil)
	h := shared.NewCachingHandler(
		sharedtest.NewTestContext(),
		http.NotFoundHandler(),
		cache,
		shared.AlwaysCachable,
		shared.URLAsCacheKey,
		shared.CacheStatusOK)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/some/url", nil)
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCaching200(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	cache := sharedtest.NewMockReadWritable(mockCtrl)
	cache.EXPECT().NewReadCloser("/some/url").Return(ioutil.NopCloser(failReader{}), nil)
	wc := sharedtest.NewMockWriteCloser(t)
	cache.EXPECT().NewWriteCloser("/some/url").Return(wc, nil)
	h := shared.NewCachingHandler(
		sharedtest.NewTestContext(),
		okHandler{},
		cache,
		shared.AlwaysCachable,
		shared.URLAsCacheKey,
		shared.CacheStatusOK)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/some/url", nil)
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, wc.IsClosed())
	assert.Equal(t, ok, wc.Bytes())
}
