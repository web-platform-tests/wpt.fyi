// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared/iotest"
	"google.golang.org/appengine/memcache"
)

func TestGet_cacheHit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var cacheID, storeID interface{}

	cache := NewMockReadWritable(mockCtrl)
	store := NewMockReadable(mockCtrl)
	cs := NewCtxCachedStore(NewTestContext(), cache, store)

	data := []byte("{}")
	cr := iotest.NewMockReadCloser(t, data)
	cache.EXPECT().NewReadCloser(&cacheID).Return(cr, nil)

	v, err := cs.Get(&cacheID, &storeID)
	assert.Nil(t, err)
	assert.Equal(t, data, v)
}

func TestGet_cacheMiss(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var cacheID, storeID interface{}

	cache := NewMockReadWritable(mockCtrl)
	store := NewMockReadable(mockCtrl)
	cs := NewCtxCachedStore(NewTestContext(), cache, store)

	data := []byte("{}")
	c := make(chan bool)
	cw := iotest.NewMockWriteCloser(t, c)
	sr := iotest.NewMockReadCloser(t, data)
	cache.EXPECT().NewReadCloser(&cacheID).Return(nil, memcache.ErrCacheMiss)
	store.EXPECT().NewReadCloser(&storeID).Return(sr, nil)
	cache.EXPECT().NewWriteCloser(&cacheID).Return(cw, nil)

	v, err := cs.Get(&cacheID, &storeID)
	assert.Nil(t, err)
	assert.Equal(t, data, v)

	b := <-c
	assert.True(t, sr.IsClosed())
	assert.True(t, b)
}

func TestGet_missing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var cacheID, storeID interface{}

	cache := NewMockReadWritable(mockCtrl)
	store := NewMockReadable(mockCtrl)
	cs := NewCtxCachedStore(NewTestContext(), cache, store)

	errMissing := errors.New("Failed to fetch from store")
	cache.EXPECT().NewReadCloser(&cacheID).Return(nil, memcache.ErrCacheMiss)
	store.EXPECT().NewReadCloser(&storeID).Return(nil, errMissing)

	_, err := cs.Get(&cacheID, &storeID)
	assert.Equal(t, errMissing, err)
}
