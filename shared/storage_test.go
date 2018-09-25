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
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine/memcache"
)

func TestGet_cacheHit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var cacheID, storeID interface{}

	cache := NewMockReadWritable(mockCtrl)
	store := NewMockReadable(mockCtrl)
	cs := NewByteCachedStore(NewTestContext(), cache, store)

	data := []byte("{}")
	cr := sharedtest.NewMockReadCloser(t, data)
	cache.EXPECT().NewReadCloser(&cacheID).Return(cr, nil)

	var v []byte
	err := cs.Get(&cacheID, &storeID, &v)
	assert.Nil(t, err)
	assert.Equal(t, data, v)
}

func TestGet_cacheMiss(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var cacheID, storeID interface{}

	cache := NewMockReadWritable(mockCtrl)
	store := NewMockReadable(mockCtrl)
	cs := NewByteCachedStore(NewTestContext(), cache, store)

	data := []byte("{}")
	cw := sharedtest.NewMockWriteCloser(t)
	sr := sharedtest.NewMockReadCloser(t, data)
	cache.EXPECT().NewReadCloser(&cacheID).Return(nil, memcache.ErrCacheMiss)
	store.EXPECT().NewReadCloser(&storeID).Return(sr, nil)
	cache.EXPECT().NewWriteCloser(&cacheID).Return(cw, nil)

	var v []byte
	err := cs.Get(&cacheID, &storeID, &v)
	assert.Nil(t, err)
	assert.Equal(t, data, v)
	assert.True(t, sr.IsClosed())
}

func TestGet_missing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var cacheID, storeID interface{}

	cache := NewMockReadWritable(mockCtrl)
	store := NewMockReadable(mockCtrl)
	cs := NewByteCachedStore(NewTestContext(), cache, store)

	errMissing := errors.New("Failed to fetch from store")
	cache.EXPECT().NewReadCloser(&cacheID).Return(nil, memcache.ErrCacheMiss)
	store.EXPECT().NewReadCloser(&storeID).Return(nil, errMissing)

	var v []byte
	err := cs.Get(&cacheID, &storeID, &v)
	assert.Equal(t, errMissing, err)
}
