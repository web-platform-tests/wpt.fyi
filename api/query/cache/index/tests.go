// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"fmt"
	"sync"

	farm "github.com/dgryski/go-farm"
)

// TestID is a unique identifier for a WPT test or sub-test.
type TestID struct {
	testID uint64
	subID  uint64
}

// Tests is an indexing component that provides fast test name lookup by TestID.
type Tests interface {
	Add(name string, subName *string) (TestID, error)
	GetName(id TestID) (string, *string, error)

	// TODO: Add filter binding function:
	// TestFilter(q string) UnboundFilter
}

// Tests is an indexing component that provides fast test name lookup by TestID.
type testsMap struct {
	tests sync.Map
}

type testName struct {
	name    string
	subName *string
}

// NewTests constructs an empty Tests instance.
func NewTests() Tests {
	return &testsMap{tests: sync.Map{}}
}

func (ts *testsMap) Add(name string, subName *string) (TestID, error) {
	id, err := computeID(name, subName)
	if err != nil {
		return id, err
	}
	ts.tests.Store(id, testName{name, subName})
	return id, nil
}

func (ts *testsMap) GetName(id TestID) (string, *string, error) {
	v, ok := ts.tests.Load(id)
	if !ok {
		return "", nil, fmt.Errorf(`Test not found; ID: %v`, id)
	}
	name := v.(testName)
	return name.name, name.subName, nil
}

// TODO: Add filter binding function:
// func TestFilter(q string) UnboundFilter {
// 	return NewTestNameFilter(q)
// }

func computeID(name string, subPtr *string) (TestID, error) {
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
