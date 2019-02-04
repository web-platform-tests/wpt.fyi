// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

//
// Shared data types used for string WPT test results in query cache.
//

// TestStatus represents the possible test result statuses.
type TestStatus int64

const (
	// TestStatusUnknown is an uninitialized TestStatus and should
	// not be used.
	TestStatusUnknown TestStatus = 0

	// TestStatusPass indicates that all tests completed successfully and passed.
	TestStatusPass TestStatus = 1

	// TestStatusOK indicates that all tests completed successfully.
	TestStatusOK TestStatus = 2

	// TestStatusError indicates that some tests did not complete
	// successfully.
	TestStatusError TestStatus = 3

	// TestStatusTimeout indicates that some tests timed out.
	TestStatusTimeout TestStatus = 4

	// TestStatusNotRun indicates that a test was not run.
	TestStatusNotRun TestStatus = 5

	// TestStatusFail indicates that a test failed.
	TestStatusFail TestStatus = 6

	// TestStatusCrash indicates that the WPT test runner crashed attempting to run the test.
	TestStatusCrash TestStatus = 7

	// TestStatusSkip indicates that the test was disabled for this test run.
	TestStatusSkip TestStatus = 8

	// TestStatusAssert indicates that a non-fatal assertion failed. This test
	// status is supported by, at least, Mozilla.
	TestStatusAssert TestStatus = 9

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

	// TestStatusNameSkip is the string representation of a test result where the
	// test was disabled for this test run.
	TestStatusNameSkip string = "SKIP"

	// TestStatusNameAssert is the string representation of a test result where
	// a non-fatal assertion failed. This test status is supported by, at least,
	// Mozilla.
	TestStatusNameAssert string = "ASSERT"

	// TestStatusDefault is the default value used when a status string cannot be
	// interpreted.
	TestStatusDefault TestStatus = TestStatusUnknown

	// TestStatusNameDefault is the default string used when a status value cannot
	// be interpreted.
	TestStatusNameDefault string = TestStatusNameUnknown
)

var testStatusValues = map[string]TestStatus{
	TestStatusNameUnknown: TestStatusUnknown,
	TestStatusNamePass:    TestStatusPass,
	TestStatusNameOK:      TestStatusOK,
	TestStatusNameError:   TestStatusError,
	TestStatusNameTimeout: TestStatusTimeout,
	TestStatusNameNotRun:  TestStatusNotRun,
	TestStatusNameFail:    TestStatusFail,
	TestStatusNameCrash:   TestStatusCrash,
	TestStatusNameSkip:    TestStatusSkip,
	TestStatusNameAssert:  TestStatusAssert,
}

var testStatusNames = map[TestStatus]string{
	TestStatusUnknown: TestStatusNameUnknown,
	TestStatusPass:    TestStatusNamePass,
	TestStatusOK:      TestStatusNameOK,
	TestStatusError:   TestStatusNameError,
	TestStatusTimeout: TestStatusNameTimeout,
	TestStatusNotRun:  TestStatusNameNotRun,
	TestStatusFail:    TestStatusNameFail,
	TestStatusCrash:   TestStatusNameCrash,
	TestStatusSkip:    TestStatusNameSkip,
	TestStatusAssert:  TestStatusNameAssert,
}

// IsPassOrOK is true if the value is TestStatusPass or TestStatusOK
func (s TestStatus) IsPassOrOK() bool {
	return s == TestStatusOK || s == TestStatusPass
}

// TestStatusValueFromString returns the enum value associated with str (if
// any), or else TestStatusDefault.
func TestStatusValueFromString(str string) TestStatus {
	v, ok := testStatusValues[str]
	if !ok {
		return TestStatusDefault
	}
	return v
}

// TestStatusStringFromValue returns the string associated with s (if any), or
// else TestStatusStringDefault.
func TestStatusStringFromValue(s TestStatus) string {
	str, ok := testStatusNames[s]
	if !ok {
		return TestStatusNameDefault
	}
	return str
}
