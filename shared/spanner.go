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

	// TestStatusNameUnknown is the string representation for an uninitialized
	// TestStatus and should not be used.
	TestStatusNameUnknown string = "UNKNOWN"

	// TestStatusNamePass is the string representation of a test result where the
	// test passed.
	TestStatusNamePass string = "PASS"

	// TestStatusNameOK is the string represnetation of a test result where the
	// test ran completely but may not have passed (and/or not all of its subtests
	// passed).
	TestStatusNameOK string = "OK"

	// TestStatusNameError is the string representation for a test result where
	// a test harness error was encountered at test runtime.
	TestStatusNameError string = "ERROR"

	// TestStatusNameTimeout is the string representation for a test result where
	// the test timed out.
	TestStatusNameTimeout string = "TIMEOUT"

	// TestStatusNameNotRun is  the string representation for a test result where
	// the test exists but was not run.
	TestStatusNameNotRun string = "NOTRUN"

	// TestStatusNameFail is the string representation of a test result where the
	// test failed.
	TestStatusNameFail string = "FAIL"

	// TestStatusNameCrash is the string representation of a test result where the
	// test runner crashed.
	TestStatusNameCrash string = "CRASH"

	// TestStatusDefault is the default value used when a status string cannot be
	// interpreted.
	TestStatusDefault int64 = TestStatusUnknown

	// TestStatusNameDefault is the default string used when a status value cannot
	// be interpreted.
	TestStatusNameDefault string = TestStatusNameUnknown
)

var testStatusValues = map[string]int64{
	TestStatusNameUnknown: TestStatusUnknown,
	TestStatusNamePass:    TestStatusPass,
	TestStatusNameOK:      TestStatusOK,
	TestStatusNameError:   TestStatusError,
	TestStatusNameTimeout: TestStatusTimeout,
	TestStatusNameNotRun:  TestStatusNotRun,
	TestStatusNameFail:    TestStatusFail,
	TestStatusNameCrash:   TestStatusCrash,
}

var testStatusNames = map[int64]string{
	TestStatusUnknown: TestStatusNameUnknown,
	TestStatusPass:    TestStatusNamePass,
	TestStatusOK:      TestStatusNameOK,
	TestStatusError:   TestStatusNameError,
	TestStatusTimeout: TestStatusNameTimeout,
	TestStatusNotRun:  TestStatusNameNotRun,
	TestStatusFail:    TestStatusNameFail,
	TestStatusCrash:   TestStatusNameCrash,
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
