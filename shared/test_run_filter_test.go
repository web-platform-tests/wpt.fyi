// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"
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
	// Use UTC to avoid DST craziness.
	now := time.Now().UTC()
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
	nextPage := filter.NextPage(loadedRuns)
	assert.Equal(t, &TestRunFilter{
		From: &twoWeeksAgo,
		To:   &aWeekAgoMinusAMilli,
	}, nextPage)
}

func TestTestRunFilter_JSONRoundTrip(t *testing.T) {
	one := 1
	chrome, _ := ParseProductSpec("chrome[experimental]")
	page := TestRunFilter{
		MaxCount: &one,
		Offset:   &one,
		Labels:   mapset.NewSet(MasterLabel),
		Products: ProductSpecs{chrome},
	}

	// Test a JSON roundtrip.
	m, err := json.Marshal(page)
	assert.Nil(t, err)
	var jsonRoundTrip TestRunFilter
	err = json.Unmarshal(m, &jsonRoundTrip)
	assert.Nil(t, err)
	assert.EqualValues(t, &one, jsonRoundTrip.MaxCount)
	assert.EqualValues(t, &one, jsonRoundTrip.Offset)
	assert.Contains(t, ToStringSlice(jsonRoundTrip.Labels), MasterLabel)
}
