//go:build small
// +build small

// Copyright 2026 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestConvertTestRunUIFilter_DatesInUTC(t *testing.T) {
	// +24:00 is a valid offset for Go, but not for JavaScript's Date.
	from := time.Date(2024, 1, 2, 3, 4, 5, 0, time.FixedZone("", 24*60*60))
	to := time.Date(2024, 2, 3, 4, 5, 6, 0, time.FixedZone("", 2*60*60))
	filter := convertTestRunUIFilter(shared.TestRunFilter{From: &from, To: &to})
	assert.Equal(t, "2024-01-01T03:04:05Z", filter.From)
	assert.Equal(t, "2024-02-03T02:05:06Z", filter.To)
}

func TestConvertTestRunUIFilter_NoDates(t *testing.T) {
	filter := convertTestRunUIFilter(shared.TestRunFilter{})
	assert.Equal(t, "", filter.From)
	assert.Equal(t, "", filter.To)
}
