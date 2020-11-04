// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"fmt"
	"strings"
)

// MultiError is a convenient wrapper of multiple errors and is itself an
// implementation of the error interface.
type MultiError struct {
	errors []error
	when   string
}

// NewMultiErrorFromChan creates a MultiError by reading from an error channel.
// The "when" parameter will be included in the error string in a "when" clause.
// If there is no error in the channel, nil will be returned.
//
// Note that it uses `range` over the channel, so users need to close the
// channel before calling this function or running it in a goroutine.
func NewMultiErrorFromChan(errors chan error, when string) error {
	var multiError MultiError
	for err := range errors {
		multiError.errors = append(multiError.errors, err)
	}
	if multiError.errors != nil {
		multiError.when = when
		return multiError
	}
	return nil
}

// NewMultiError creates a MultiError from a slice of errors. The "when"
// parameter will be included in the error string in a "when" clause.
// If the slice is empty, nil will be returned.
func NewMultiError(errors []error, when string) error {
	if len(errors) == 0 {
		return nil
	}
	return MultiError{errors, when}
}

func (e MultiError) Error() string {
	if e.Count() == 0 {
		return ""
	}
	errStrs := make([]string, len(e.errors))
	for i, err := range e.errors {
		errStrs[i] = err.Error()
	}
	return fmt.Sprintf("%d error(s) occurred when %s:\n%s",
		len(e.errors), e.when, strings.Join(errStrs, "\n"))
}

// Count returns the number of errors in this MultiError.
func (e MultiError) Count() int {
	return len(e.errors)
}

// Errors returns the inner error slice of a MultiError.
func (e MultiError) Errors() []error {
	return e.errors
}
