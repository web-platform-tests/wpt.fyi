// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"fmt"

	farm "github.com/dgryski/go-farm"
)

// TestID is a unique identifier for a WPT test or sub-test.
type TestID struct {
	testID uint64
	subID  uint64
}

// Tests is an indexing component that provides fast test name lookup by TestID.
type Tests interface {
	// Add adds a new named test/subtest to a tests index, using the given TestID
	// to identify the test/subtest in the index.
	Add(TestID, string, *string)
	// GetName retrieves the name/subtest name associated with a given TestID. If
	// the index does not recognize the TestID, then an error is returned.
	GetName(TestID) (string, *string, error)

	Range(func(TestID) bool)
}

// Tests is an indexing component that provides fast test name lookup by TestID.
type testsMap struct {
	tests map[TestID]testName
}

type testName struct {
	name    string
	subName *string
}

// NewTests constructs an empty Tests instance.
func NewTests() Tests {
	return &testsMap{tests: make(map[TestID]testName)}
}

func (ts *testsMap) Add(t TestID, name string, subName *string) {
	ts.tests[t] = testName{name, subName}
}

func (ts *testsMap) GetName(id TestID) (string, *string, error) {
	name, ok := ts.tests[id]
	if !ok {
		return "", nil, fmt.Errorf(`Test not found; ID: %v`, id)
	}
	return name.name, name.subName, nil
}

func (ts *testsMap) Range(f func(TestID) bool) {
	for t := range ts.tests {
		if !f(t) {
			break
		}
	}
}

func computeTestID(name string, subPtr *string) (TestID, error) {
	var s uint64
	t := farm.Fingerprint64([]byte(name))
	if subPtr != nil && *subPtr != "" {
		s = farm.Fingerprint64([]byte(*subPtr))
		if s == 0 {
			return TestID{}, fmt.Errorf(`Subtest ID for string "%s" is 0`, *subPtr)
		}
	}
	return TestID{t, s}, nil
}
