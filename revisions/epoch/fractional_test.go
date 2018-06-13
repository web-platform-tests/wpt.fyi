// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package epoch_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
)

var eightHourly = epoch.EightHourly{}
var fourHourly = epoch.FourHourly{}
var twoHourly = epoch.TwoHourly{}

func testClosePositive(t *testing.T, e epoch.Epoch) {
	n := int(e.GetData().MaxDuration.Hours())
	justPrior := time.Date(2018, 4, 1, n-1, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 1, n, 0, 0, 0, time.UTC)
	assert.True(t, e.IsEpochal(justPrior, justAfter))
	assert.True(t, e.IsEpochal(justAfter, justPrior))
}

func testFarPositive(t *testing.T, e epoch.Epoch) {
	thisYear := time.Now()
	lastYear := thisYear.AddDate(-1, 0, 0)
	assert.True(t, e.IsEpochal(lastYear, thisYear))
	assert.True(t, e.IsEpochal(thisYear, lastYear))
}

func testTZPositive(t *testing.T, e epoch.Epoch) {
	n := int(e.GetData().MaxDuration.Hours())
	justPrior := time.Date(2018, 4, 1, n-1, 59, 59, 999999999, time.UTC)
	justAfter := time.Date(2018, 4, 1, n-1, 0, 0, 0, time.FixedZone("UTC-1", -60*60))
	assert.True(t, e.IsEpochal(justPrior, justAfter))
	assert.True(t, e.IsEpochal(justAfter, justPrior))
}

var startDay = time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
var justAfterStartDay = time.Date(2018, 4, 1, 0, 0, 0, 1, time.UTC)
var justAfterStartDayTZ = time.Date(2018, 3, 31, 23, 0, 0, 1, time.FixedZone("UTC-1", -60*60))

func testCloseNegative(t *testing.T, e epoch.Epoch) {
	assert.False(t, e.IsEpochal(startDay, justAfterStartDay))
	assert.False(t, e.IsEpochal(justAfterStartDay, startDay))
}

func testFarNegative(t *testing.T, e epoch.Epoch) {
	n := int(e.GetData().MaxDuration.Hours())
	start := time.Date(2018, 4, 1, n, 0, 0, 0, time.UTC)
	justBeforeEnd := time.Date(2018, 4, 1, 2*n-1, 59, 59, 999999999, time.UTC)
	assert.False(t, e.IsEpochal(start, justBeforeEnd))
	assert.False(t, e.IsEpochal(justBeforeEnd, start))
}

func testTZNegative(t *testing.T, e epoch.Epoch) {
	assert.False(t, e.IsEpochal(startDay, justAfterStartDayTZ))
	assert.False(t, e.IsEpochal(justAfterStartDayTZ, startDay))
}

//
// EightHourly
//

func TestIsEightHourly_Close(t *testing.T) {
	testClosePositive(t, eightHourly)
}

func TestIsEightHourly_Far(t *testing.T) {
	testFarPositive(t, eightHourly)
}

func TestIsEightHourly_TZ(t *testing.T) {
	testTZPositive(t, eightHourly)
}

func TestIsNotEightHourly_Close(t *testing.T) {
	testCloseNegative(t, eightHourly)
}

func TestIsNotEightHourly_Far(t *testing.T) {
	testFarNegative(t, eightHourly)
}

func TestIsNotEightHourly_TZ(t *testing.T) {
	testTZNegative(t, eightHourly)
}

//
// FourHourly
//

func TestIsFourHourly_Close(t *testing.T) {
	testClosePositive(t, fourHourly)
}

func TestIsFourHourly_Far(t *testing.T) {
	testFarPositive(t, fourHourly)
}

func TestIsFourHourly_TZ(t *testing.T) {
	testTZPositive(t, fourHourly)
}

func TestIsNotFourHourly_Close(t *testing.T) {
	testCloseNegative(t, fourHourly)
}

func TestIsNotFourHourly_Far(t *testing.T) {
	testFarNegative(t, fourHourly)
}

func TestIsNotFourHourly_TZ(t *testing.T) {
	testTZNegative(t, fourHourly)
}

//
// TwoHourly
//

func TestIsTwoHourly_Close(t *testing.T) {
	testClosePositive(t, twoHourly)
}

func TestIsTwoHourly_Far(t *testing.T) {
	testFarPositive(t, twoHourly)
}

func TestIsTwoHourly_TZ(t *testing.T) {
	testTZPositive(t, twoHourly)
}

func TestIsNotTwoHourly_Close(t *testing.T) {
	testCloseNegative(t, twoHourly)
}

func TestIsNotTwoHourly_Far(t *testing.T) {
	testFarNegative(t, twoHourly)
}

func TestIsNotTwoHourly_TZ(t *testing.T) {
	testTZNegative(t, twoHourly)
}
