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

var lastYear = time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
var thisYear = time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)

func testFarPositive(t *testing.T, e epoch.Epoch) {
	lastYear := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)
	thisYear := time.Date(2018, 5, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, e.IsEpochal(lastYear, thisYear))
	assert.True(t, e.IsEpochal(thisYear, lastYear))
}

var startDay = time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC)
var justAfterStartDay = time.Date(2018, 4, 1, 0, 0, 0, 1, time.UTC)

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

//
// EightHourly
//

func TestIsEightHourly_Close(t *testing.T) {
	testClosePositive(t, eightHourly)
}

func TestIsEightHourly_Far(t *testing.T) {
	testFarPositive(t, eightHourly)
}

func TestIsNotEightHourly_Close(t *testing.T) {
	testCloseNegative(t, eightHourly)
}

func TestIsNotEightHourly_Far(t *testing.T) {
	testFarNegative(t, eightHourly)
}
