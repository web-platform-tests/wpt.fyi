// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTestRunFilter_NextPage_MaxCount(t *testing.T) {
	ten := 10
	filter := TestRunFilter{
		MaxCount: &ten,
	}
	chrome, _ := ParseProductSpec("chrome")
	loadedRuns := TestRunsByProduct{
		ProductTestRuns{
			Product:  chrome,
			TestRuns: make(TestRuns, 10),
		},
	}
	assert.Equal(t, &TestRunFilter{
		MaxCount: &ten,
		Offset:   &ten,
	}, filter.NextPage(loadedRuns))
}

func TestTestRunFilter_NextPage_From(t *testing.T) {
	now := time.Now()
	aWeekAgo := now.AddDate(0, 0, -7)
	filter := TestRunFilter{
		From: &aWeekAgo,
		To:   &now,
	}
	chrome, _ := ParseProductSpec("chrome")
	loadedRuns := TestRunsByProduct{
		ProductTestRuns{
			Product:  chrome,
			TestRuns: make(TestRuns, 1),
		},
	}
	twoWeeksAgo := aWeekAgo.AddDate(0, 0, -7)
	aWeekAgoMinusAMilli := aWeekAgo.Add(-time.Millisecond)
	assert.Equal(t, &TestRunFilter{
		From: &twoWeeksAgo,
		To:   &aWeekAgoMinusAMilli,
	}, filter.NextPage(loadedRuns))
}
