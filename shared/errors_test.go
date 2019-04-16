// +build small

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMultiErrorFromChan_non_empty(t *testing.T) {
	errs := make(chan error, 2)
	errs <- errors.New("test1")
	errs <- errors.New("test2")
	close(errs)
	err := NewMultiErrorFromChan(errs, "testing")
	assert.Equal(t, "2 error(s) occurred when testing:\ntest1\ntest2\n", err.Error())
	_, ok := err.(*MultiError)
	assert.True(t, ok)
}

func TestNewMultiErrorFromChan_nil(t *testing.T) {
	errs := make(chan error)
	close(errs)
	// It is vital to pre-declare the type of err.
	var err error
	err = NewMultiErrorFromChan(errs, "testing")
	// Do NOT use assert.Nil: we use the `nil` literal intentionally here.
	// This is equivalent to `err == nil`. Since err is declared as error,
	// so we are comparing err against (error)(nil), which will fail if
	// NewMultiErrorFromChan incorrectly returns a concrete
	// (*MultiError)(nil).
	assert.Equal(t, nil, err)
	_, ok := err.(*MultiError)
	assert.False(t, ok)
}
