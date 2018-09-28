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

// MockWriteCloser tracks write+close state for testing purposes.
type MockWriteCloser struct {
	b      bytes.Buffer
	closed bool
	t      *testing.T
	c      chan bool
}

// Write ensures MockWriteCloser isn't closed and delegates to an underlying
// buffer for writing.
func (mwc *MockWriteCloser) Write(p []byte) (n int, err error) {
	assert.False(mwc.t, mwc.closed)
	return mwc.b.Write(p)
}

// Close stores "closed" state and synchronizes by sending true to
// MockWriteCloser.c iff it is not nil. Note that the synchronization message
// is sent from this goroutine; i.e., Close() will not return until another
// goroutine receives the message.
func (mwc *MockWriteCloser) Close() error {
	mwc.closed = true
	if mwc.c != nil {
		mwc.c <- true
	}
	return nil
}

// NewMockWriteCloser constructs a MockWriteCloser for a given test and optional
// on-close synchronization channel. MockWriteCloser will send true to the
// channel on Close().
func NewMockWriteCloser(t *testing.T, c chan bool) *MockWriteCloser {
	return &MockWriteCloser{
		b:      bytes.Buffer{},
		closed: false,
		t:      t,
		c:      c,
	}
}

// MockReadCloser implements reading from a predefined byte slice and tracks
// closed state for testing.
type MockReadCloser struct {
	rc     io.ReadCloser
	closed bool
	t      *testing.T
}

// Read ensures that MockReadCloser has not be closed, then delegates to a
// reader that wraps a predefined byte slice.
func (mrc *MockReadCloser) Read(p []byte) (n int, err error) {
	assert.False(mrc.t, mrc.closed)
	return mrc.rc.Read(p)
}

// Close tracks closed state and returns nil.
func (mrc *MockReadCloser) Close() error {
	mrc.closed = true
	return nil
}

// NewMockReadCloser constructs a ReadCloser bound to the given test and byte
// slice.
func NewMockReadCloser(t *testing.T, data []byte) *MockReadCloser {
	return &MockReadCloser{
		rc:     ioutil.NopCloser(bytes.NewReader(data)),
		closed: false,
		t:      t,
	}
}

// IsClosed allows test code to query whether or not a MockReadCloser has been
// closed.
func (mrc *MockReadCloser) IsClosed() bool {
	return mrc.closed
}
