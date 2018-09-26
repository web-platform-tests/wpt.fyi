// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package test

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockWriteCloser struct {
	b      bytes.Buffer
	closed bool
	t      *testing.T
}

func (mwc *MockWriteCloser) Write(p []byte) (n int, err error) {
	assert.False(mwc.t, mwc.closed)
	return mwc.b.Write(p)
}

func (mwc *MockWriteCloser) Close() error {
	mwc.closed = true
	return nil
}

func NewMockWriteCloser(t *testing.T) *MockWriteCloser {
	return &MockWriteCloser{
		b:      bytes.Buffer{},
		closed: false,
		t:      t,
	}
}

type MockReadCloser struct {
	rc     io.ReadCloser
	closed bool
	t      *testing.T
}

func (mrc *MockReadCloser) Read(p []byte) (n int, err error) {
	assert.False(mrc.t, mrc.closed)
	return mrc.rc.Read(p)
}

func (mrc *MockReadCloser) Close() error {
	mrc.closed = true
	return nil
}

func NewMockReadCloser(t *testing.T, data []byte) *MockReadCloser {
	return &MockReadCloser{
		rc:     ioutil.NopCloser(bytes.NewReader(data)),
		closed: false,
		t:      t,
	}
}

func (mrc *MockReadCloser) IsClosed() bool {
	return mrc.closed
}
