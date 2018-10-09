// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

//
// Shared data types used for string WPT test results in Cloud Spanner.
//

const (
	// TestStatusUnknown is an uninitialized TestStatus and should
	// not be used.
	TestStatusUnknown int64 = 0

	// TestStatusPass indicates that all tests completed successfully and passed.
	TestStatusPass int64 = 1

	// TestStatusOK indicates that all tests completed successfully.
	TestStatusOK int64 = 2

	// TestStatusError indicates that some tests did not complete
	// successfully.
	TestStatusError int64 = 3

	// TestStatusTimeout indicates that some tests timed out.
	TestStatusTimeout int64 = 4

	// TestStatusNotRun indicates that a test was not run.
	TestStatusNotRun int64 = 5

	// TestStatusFail indicates that a test failed.
	TestStatusFail int64 = 6

	// TestStatusCrash indicates that the WPT test runner crashed attempting to run the test.
	TestStatusCrash int64 = 7

	// TestStatusDefault is the default value used when a status string cannot be
	// interpreted.
	TestStatusDefault int64 = 0

	// TestStatusNameDefault is the default string used when a status value cannot
	// be interpreted.
	TestStatusNameDefault string = "UNKNOWN"
)

var testStatusValues = map[string]int64{
	"UNKNOWN": 0,
	"PASS":    1,
	"OK":      2,
	"ERROR":   3,
	"TIMEOUT": 4,
	"NOT_RUN": 5,
	"FAIL":    6,
	"CRASH":   7,
}

var testStatusNames = map[int64]string{
	0: "UNKNOWN",
	1: "PASS",
	2: "OK",
	3: "ERROR",
	4: "TIMEOUT",
	5: "NOT_RUN",
	6: "FAIL",
	7: "CRASH",
}

// TestStatusValueFromString returns the enum value associated with str (if
// any), or else TestStatusDefault.
func TestStatusValueFromString(str string) int64 {
	v, ok := testStatusValues[str]
	if !ok {
		return TestStatusDefault
	}
	return v
}

// TestStatusStringFromValue returns the string associated with s (if any), or
// else TestStatusStringDefault.
func TestStatusStringFromValue(s int64) string {
	str, ok := testStatusNames[s]
	if !ok {
		return TestStatusNameDefault
	}
	return str
}
