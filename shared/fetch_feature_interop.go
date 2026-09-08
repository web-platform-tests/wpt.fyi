// Copyright 2026 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -build_flags=--mod=mod -destination sharedtest/fetch_feature_interop_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared FetchFeatureInterop

package shared

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
)

const (
	// featureInteropTrendlineURLFormat is the GitHub URL for fetching the interop
	// score of every date scored, taking the channel name. One row per date. The
	// file name is CSV_NAME in results-analysis' feature-level-interop.js, which
	// is the one thing the two repositories have to agree on.
	featureInteropTrendlineURLFormat = "https://raw.githubusercontent.com/web-platform-tests/results-analysis/gh-pages/data/%s-browser-feature-interop-exclusions.csv"
	// featureInteropBreakdownURLFormat is the GitHub URL for fetching one date's
	// score of every web feature, taking the channel name and the date. The
	// breakdown is published per date, so it cannot be fetched until the
	// trendline says which date is newest.
	featureInteropBreakdownURLFormat = "https://raw.githubusercontent.com/web-platform-tests/results-analysis/gh-pages/data/%s-browser-feature-interop-exclusions-%s.csv"
)

// errNoDateColumn occurs when the trendline has no column named "date", which
// every one of its rows is keyed by.
var errNoDateColumn = errors.New("no date column in feature interop trendline")

// errNoScoredDates occurs when the trendline holds its fields but no rows, so
// there is no newest date to fetch a breakdown for.
var errNoScoredDates = errors.New("no dates scored in feature interop trendline")

// FeatureInteropData stores the interop score of every date scored, plus the
// per-web-feature breakdown of the newest of them.
type FeatureInteropData struct {
	// The wpt revision the newest scored date ran at.
	LastUpdateRevision string `json:"lastUpdateRevision"`
	// The wpt release whose web features manifest the newest date was scored against.
	Manifest string `json:"manifest"`
	// The newest date scored, which Features breaks down.
	Date string `json:"date"`
	// Fields correspond to the fields (columns) in the trendline table.
	Fields []string `json:"fields"`
	// Trendline table, one row per date, in chronological order.
	Data [][]string `json:"data"`
	// FeatureFields correspond to the fields (columns) in the breakdown table.
	FeatureFields []string `json:"featureFields"`
	// Breakdown table, one row per web feature.
	Features [][]string `json:"features"`
}

// fieldValue returns the value `row` holds in the column named `name`, or the
// empty string if `fields` has no such column. Columns are located by name so
// that adding one to the published CSV cannot shift a value read here.
func fieldValue(fields []string, row []string, name string) string {
	for i, field := range fields {
		if field == name && i < len(row) {
			return row[i]
		}
	}

	return ""
}

// NewestScoredDate returns the date of the last row of `trendline`, which
// feature-level-interop.js writes in chronological order with its fields at the
// 0th index. It is the date whose breakdown is worth showing, so failing to
// find one is an error rather than an empty result.
func NewestScoredDate(trendline [][]string) (string, error) {
	if len(trendline) < 2 {
		return "", errNoScoredDates
	}

	newest := trendline[len(trendline)-1]
	date := fieldValue(trendline[0], newest, "date")
	if date == "" {
		return "", errNoDateColumn
	}

	return date, nil
}

// ExtractFeatureInteropData generates FeatureInteropData for
// feature_interop_handler from `trendline` and `breakdown`, both [][]string
// with the 0th index as fields and the rest as the table; e.g.
// [[sha,date,manifest,chrome-version,chrome,firefox-version,firefox,safari-version,safari,interop,features],
// [25f5e56ff8faa3aa9c65d1e0ed5b64ee7d4c0a11,2026-02-21,merge_pr_54321,145.0.7632.77,87.4,147.0.4,74.5,26.3,71.9,65.7,812],
// ...]
// The revision, manifest and date reported are the newest row's, so that they
// describe the breakdown as well as the end of the trendline.
func ExtractFeatureInteropData(trendline [][]string, breakdown [][]string) FeatureInteropData {
	if len(trendline) < 2 {
		return FeatureInteropData{}
	}

	var response FeatureInteropData
	response.Fields = trendline[0]
	response.Data = trendline[1:]

	newest := trendline[len(trendline)-1]
	response.LastUpdateRevision = fieldValue(trendline[0], newest, "sha")
	response.Manifest = fieldValue(trendline[0], newest, "manifest")
	response.Date = fieldValue(trendline[0], newest, "date")

	// A trendline without a breakdown still draws a graph, so the table is left
	// empty rather than the whole response discarded.
	if len(breakdown) > 1 {
		response.FeatureFields = breakdown[0]
		response.Features = breakdown[1:]
	}

	return response
}

// FetchFeatureInterop encapsulates the Fetch(ctx, isExperimental) method for testing.
type FetchFeatureInterop interface {
	Fetch(ctx context.Context, isExperimental bool) (trendline [][]string, breakdown [][]string, err error)
}

type fetchFeatureInterop struct{}

// Fetch() fetches the interop score trendline in CSV from GitHub for the given
// channel, in chronological order, then the breakdown of its newest date. A
// breakdown that cannot be fetched is returned empty rather than as an error,
// as a trendline without one still draws a graph.
func (f fetchFeatureInterop) Fetch(ctx context.Context, isExperimental bool) ([][]string, [][]string, error) {
	channel := "stable"
	if isExperimental {
		channel = "experimental"
	}

	trendline, err := fetchCSV(ctx, fmt.Sprintf(featureInteropTrendlineURLFormat, channel))
	if err != nil {
		return nil, nil, err
	}

	date, err := NewestScoredDate(trendline)
	if err != nil {
		return nil, nil, err
	}

	// The trendline and the breakdown are separate files, so a partial deploy can
	// publish one without the other. Failing here would lose the graph over a
	// missing table, and the caching handler stores only 200s, so every request
	// would refetch both CSVs until the breakdown appeared.
	breakdown, err := fetchCSV(ctx, fmt.Sprintf(featureInteropBreakdownURLFormat, channel, date))
	if err != nil {
		GetLogger(ctx).Warningf("Failed to fetch the %s feature interop breakdown for %s. %s", channel, date, err.Error())
	}

	return trendline, breakdown, nil
}

// fetchCSV reads the CSV published at `url` into a table with its fields at the
// 0th index. A browser version can hold a comma, so the fields are parsed
// rather than split on.
func fetchCSV(ctx context.Context, url string) ([][]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("feature interop request for %q failed: %w", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("feature interop fetch from %q failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK HTTP status code of %d from %q", resp.StatusCode, url)
	}

	data, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("feature interop CSV parse from %q failed: %w", url, err)
	}

	return data, nil
}

// NewFetchFeatureInterop returns an instance of FetchFeatureInterop for
// apiFeatureInteropHandler.
func NewFetchFeatureInterop() FetchFeatureInterop {
	return fetchFeatureInterop{}
}
