// Copyright 2026 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build small

package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The fields feature-level-interop.js writes, and two dates of scores. The
// sha's leading digits are load-bearing: a value read from the wrong column
// still parses as a number, so a positional mistake shows up as a plausible
// score rather than as an error.
var (
	trendlineFields = []string{
		"sha", "date", "manifest",
		"chrome-version", "chrome",
		"firefox-version", "firefox",
		"safari-version", "safari",
		"interop", "features",
	}
	trendlineOlder = []string{
		"25f5e56ff8faa3aa9c65d1e0ed5b64ee7d4c0a11", "2026-02-21", "merge_pr_54321",
		"145.0.7632.77", "87.4", "147.0.4", "74.5", "26.3", "71.9", "65.7", "812",
	}
	trendlineNewer = []string{
		"9a1b2c3d4e5f60718293a4b5c6d7e8f901234567", "2026-02-28", "merge_pr_54999",
		"146.0.7680.31", "88.1", "148.0", "74.9", "26.3", "71.8", "66.2", "818",
	}
	breakdownFields = []string{"feature", "chrome", "firefox", "safari", "interop", "tests"}
	breakdownGrid   = []string{"grid", "100.0", "98.2", "91.3", "91.3", "412"}
)

func TestNewestScoredDate_LastRowWins(t *testing.T) {
	trendline := [][]string{trendlineFields, trendlineOlder, trendlineNewer}

	date, err := NewestScoredDate(trendline)

	require.NoError(t, err)
	assert.Equal(t, "2026-02-28", date)
}

func TestNewestScoredDate_LocatesDateByName(t *testing.T) {
	// The date is no longer the second column, so a hard-coded index would
	// return the sha here.
	trendline := [][]string{
		{"date", "sha", "interop"},
		{"2026-02-28", "9a1b2c3d", "66.2"},
	}

	date, err := NewestScoredDate(trendline)

	require.NoError(t, err)
	assert.Equal(t, "2026-02-28", date)
}

func TestNewestScoredDate_NoRows(t *testing.T) {
	_, err := NewestScoredDate([][]string{trendlineFields})

	assert.ErrorIs(t, err, errNoScoredDates)
}

func TestNewestScoredDate_NoDateColumn(t *testing.T) {
	trendline := [][]string{
		{"sha", "manifest", "interop"},
		{"9a1b2c3d", "merge_pr_54999", "66.2"},
	}

	_, err := NewestScoredDate(trendline)

	assert.ErrorIs(t, err, errNoDateColumn)
}

func TestExtractFeatureInteropData_ReportsNewestDate(t *testing.T) {
	trendline := [][]string{trendlineFields, trendlineOlder, trendlineNewer}
	breakdown := [][]string{breakdownFields, breakdownGrid}

	result := ExtractFeatureInteropData(trendline, breakdown)

	assert.Equal(t, "9a1b2c3d4e5f60718293a4b5c6d7e8f901234567", result.LastUpdateRevision)
	assert.Equal(t, "merge_pr_54999", result.Manifest)
	assert.Equal(t, "2026-02-28", result.Date)
	assert.Equal(t, trendlineFields, result.Fields)
	assert.Equal(t, [][]string{trendlineOlder, trendlineNewer}, result.Data)
	assert.Equal(t, breakdownFields, result.FeatureFields)
	assert.Equal(t, [][]string{breakdownGrid}, result.Features)
}

func TestExtractFeatureInteropData_NoRows(t *testing.T) {
	result := ExtractFeatureInteropData([][]string{trendlineFields}, [][]string{breakdownFields})

	assert.Equal(t, FeatureInteropData{}, result)
}

func TestExtractFeatureInteropData_NoBreakdown(t *testing.T) {
	// A date can be scored before its breakdown is published, and the graph is
	// still worth drawing without the table.
	trendline := [][]string{trendlineFields, trendlineNewer}

	result := ExtractFeatureInteropData(trendline, nil)

	assert.Equal(t, "2026-02-28", result.Date)
	assert.Equal(t, [][]string{trendlineNewer}, result.Data)
	assert.Nil(t, result.FeatureFields)
	assert.Nil(t, result.Features)
}
