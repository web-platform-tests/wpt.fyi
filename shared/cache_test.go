// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared_test

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestGet_cacheHit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var cacheID, storeID interface{}

	cache := sharedtest.NewMockReadWritable(mockCtrl)
	store := sharedtest.NewMockReadable(mockCtrl)
	cs := shared.NewByteCachedStore(sharedtest.NewTestContext(), cache, store)

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

	cache := sharedtest.NewMockReadWritable(mockCtrl)
	store := sharedtest.NewMockReadable(mockCtrl)
	cs := shared.NewByteCachedStore(sharedtest.NewTestContext(), cache, store)

	data := []byte("{}")
	errMissing := errors.New("failed to fetch from store")
	cw := sharedtest.NewMockWriteCloser(t)
	sr := sharedtest.NewMockReadCloser(t, data)
	cache.EXPECT().NewReadCloser(&cacheID).Return(nil, errMissing)
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

	cache := sharedtest.NewMockReadWritable(mockCtrl)
	store := sharedtest.NewMockReadable(mockCtrl)
	cs := shared.NewByteCachedStore(sharedtest.NewTestContext(), cache, store)

	errMissing := errors.New("failed to fetch from store")
	cache.EXPECT().NewReadCloser(&cacheID).Return(nil, errMissing)
	store.EXPECT().NewReadCloser(&storeID).Return(nil, errMissing)

	var v []byte
	err := cs.Get(&cacheID, &storeID, &v)
	assert.Equal(t, errMissing, err)
}
