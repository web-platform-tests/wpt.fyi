// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSHAParam(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/", nil)
	runSHA, err := ParseSHAParam(r)
	assert.Nil(t, err)
	assert.Equal(t, "latest", runSHA)
}

func TestParseSHAParam_Latest(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?sha=latest", nil)
	runSHA, err := ParseSHAParam(r)
	assert.Nil(t, err)
	assert.Equal(t, "latest", runSHA)
}

func TestParseSHAParam_ShortSHA(t *testing.T) {
	sha := "0123456789"
	r := httptest.NewRequest("GET", "http://wpt.fyi/?sha="+sha, nil)
	runSHA, err := ParseSHAParam(r)
	assert.Nil(t, err)
	assert.Equal(t, sha, runSHA)
}

func TestParseSHAParam_FullSHA(t *testing.T) {
	sha := "0123456789aaaaabbbbbcccccdddddeeeeefffff"
	r := httptest.NewRequest("GET", "http://wpt.fyi/?sha="+sha, nil)
	runSHA, err := ParseSHAParam(r)
	assert.Nil(t, err)
	assert.Equal(t, sha[:10], runSHA)
}

func TestParseSHAParam_BadRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?sha=%zz", nil)
	runSHA, err := ParseSHAParam(r)
	assert.NotNil(t, err)
	assert.Equal(t, "latest", runSHA)
}

func TestParseSHAParam_NonSHA(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?sha=123", nil)
	_, err := ParseSHAParam(r)
	assert.NotNil(t, err)
}

func TestParseSHAParam_NonSHA_2(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?sha=zapper0123", nil)
	_, err := ParseSHAParam(r)
	assert.NotNil(t, err)
}

func TestParseBrowserParam(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/", nil)
	browser, err := ParseBrowserParam(r)
	assert.Nil(t, err)
	assert.Equal(t, "", browser)
}

func TestParseBrowserParam_Chrome(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browser=chrome", nil)
	browser, err := ParseBrowserParam(r)
	assert.Nil(t, err)
	assert.Equal(t, "chrome", browser)
}

func TestParseBrowserParam_Invalid(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browser=invalid", nil)
	browser, err := ParseBrowserParam(r)
	assert.NotNil(t, err)
	assert.Equal(t, "", browser)
}

func TestGetBrowsersForRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/", nil)
	browsers, err := GetBrowsersForRequest(r)
	assert.Nil(t, err)
	defaultBrowsers, err := GetBrowserNames()
	assert.Equal(t, defaultBrowsers, browsers)
}

func TestGetBrowsersForRequest_ChromeSafari(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browsers=chrome,safari", nil)
	browsers, err := GetBrowsersForRequest(r)
	assert.Nil(t, err)
	assert.Equal(t, []string{"chrome", "safari"}, browsers)
}

func TestGetBrowsersForRequest_ChromeInvalid(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browsers=chrome,invalid", nil)
	_, err := GetBrowsersForRequest(r)
	assert.NotNil(t, err)
}

func TestGetBrowsersForRequest_EmptyCommas(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browsers=,chrome,,,,safari,,", nil)
	browsers, err := GetBrowsersForRequest(r)
	assert.Nil(t, err)
	assert.Equal(t, []string{"chrome", "safari"}, browsers)
}

func TestGetBrowsersForRequest_SafariChrome(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browsers=safari,chrome", nil)
	browsers, err := GetBrowsersForRequest(r)
	assert.Nil(t, err)
	assert.Equal(t, []string{"chrome", "safari"}, browsers) // Alphabetical
}

func TestGetBrowsersForRequest_MultiBrowserParam_SafariChrome(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browser=safari&browser=chrome", nil)
	browsers, err := GetBrowsersForRequest(r)
	assert.Nil(t, err)
	assert.Equal(t, []string{"chrome", "safari"}, browsers)
}

func TestGetBrowsersForRequest_MultiBrowserParam_SafariInvalid(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browser=safari&browser=invalid", nil)
	_, err := GetBrowsersForRequest(r)
	assert.NotNil(t, err)
}

func TestGetBrowsersForRequest_ChromeLabel(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?label=chrome", nil)
	browsers, err := GetBrowsersForRequest(r)
	assert.Nil(t, err)
	assert.Equal(t, []string{"chrome", "chrome-experimental"}, browsers)
}

func TestGetBrowsersForRequest_ExperimentalLabel(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?labels=experimental", nil)
	browsers, err := GetBrowsersForRequest(r)
	assert.Nil(t, err)
	names, _ := GetBrowserNames()
	for i := range names {
		names[i] = names[i] + "-" + ExperimentalLabel
	}
	assert.Equal(t, names, browsers)
}

func TestGetBrowsersForRequest_ChromeAndExperimentalLabel(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?labels=chrome,experimental", nil)
	browsers, err := GetBrowsersForRequest(r)
	assert.Nil(t, err)
	assert.Equal(t, []string{"chrome-experimental"}, browsers)
}

func TestParseMaxCountParam_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/", nil)
	count, err := ParseMaxCountParam(r)
	assert.Nil(t, err)
	assert.Equal(t, MaxCountDefaultValue, count)
}

func TestParseMaxCountParam_TooSmall(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?max-count=0", nil)
	count, err := ParseMaxCountParam(r)
	assert.Nil(t, err)
	assert.Equal(t, MaxCountMinValue, count)

	r = httptest.NewRequest("GET", "http://wpt.fyi/?max-count=-1", nil)
	count, err = ParseMaxCountParam(r)
	assert.Nil(t, err)
	assert.Equal(t, MaxCountMinValue, count)
}

func TestParseMaxCountParam_TooLarge(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?max-count=123456789", nil)
	count, err := ParseMaxCountParam(r)
	assert.Nil(t, err)
	assert.Equal(t, MaxCountMaxValue, count)

	r = httptest.NewRequest("GET", "http://wpt.fyi/?max-count=100000000", nil)
	count, err = ParseMaxCountParam(r)
	assert.Nil(t, err)
	assert.Equal(t, MaxCountMaxValue, count)
}

func TestParseMaxCountParam(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?max-count=2", nil)
	count, err := ParseMaxCountParam(r)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)
}

func TestParsePathsParam_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff", nil)
	paths := ParsePathsParam(r)
	assert.Nil(t, paths)
}

func TestParsePathsParam_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?path=", nil)
	paths := ParsePathsParam(r)
	assert.Nil(t, paths)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?paths=", nil)
	paths = ParsePathsParam(r)
	assert.Nil(t, paths)
}

func TestParsePathsParam_Path_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?path=/css&path=/css", nil)
	paths := ParsePathsParam(r)
	assert.Equal(t, 1, paths.Cardinality())
}

func TestParsePathsParam_Paths_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?paths=/css,/css", nil)
	paths := ParsePathsParam(r)
	assert.Equal(t, 1, paths.Cardinality())
}

func TestParsePathsParam_PathsAndPath_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?paths=/css&path=/css", nil)
	paths := ParsePathsParam(r)
	assert.Equal(t, 1, paths.Cardinality())
}

func TestParsePathsParam_Paths_DiffFilter(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?paths=/css&path=/css", nil)
	filter, err := ParseDiffFilterParams(r)
	assert.Nil(t, err)
	assert.Equal(t, 1, filter.Paths.Cardinality())
}

func TestParseDiffFilterParam(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=A", nil)
	filter, _ := ParseDiffFilterParams(r)
	assert.Equal(t, DiffFilterParam{Added: true, Deleted: false, Changed: false}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=D", nil)
	filter, _ = ParseDiffFilterParams(r)
	assert.Equal(t, DiffFilterParam{Added: false, Deleted: true, Changed: false}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=C", nil)
	filter, _ = ParseDiffFilterParams(r)
	assert.Equal(t, DiffFilterParam{Added: false, Deleted: false, Changed: true}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=CAD", nil)
	filter, _ = ParseDiffFilterParams(r)
	assert.Equal(t, DiffFilterParam{Added: true, Deleted: true, Changed: true}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=CD", nil)
	filter, _ = ParseDiffFilterParams(r)
	assert.Equal(t, DiffFilterParam{Added: false, Deleted: true, Changed: true}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=CACA", nil)
	filter, _ = ParseDiffFilterParams(r)
	assert.Equal(t, DiffFilterParam{Added: true, Deleted: false, Changed: true}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=U", nil)
	filter, _ = ParseDiffFilterParams(r)
	assert.Equal(t, DiffFilterParam{Unchanged: true}, filter)
}

func TestParseDiffFilterParam_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff", nil)
	filter, err := ParseDiffFilterParams(r)
	assert.Nil(t, err)
	assert.Equal(t, DiffFilterParam{Added: true, Deleted: true, Changed: true, Unchanged: false}, filter)
}

func TestParseDiffFilterParam_Invalid(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=Z", nil)
	_, err := ParseDiffFilterParams(r)
	assert.NotNil(t, err)
}

func TestParseLabelsParam_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs", nil)
	labels := ParseLabelsParam(r)
	assert.Nil(t, labels)
}

func TestParseLabelsParam_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs?label=", nil)
	labels := ParseLabelsParam(r)
	assert.Nil(t, labels)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/runs?labels=", nil)
	labels = ParseLabelsParam(r)
	assert.Nil(t, labels)
}

func TestParseLabelsParam_Label_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs?label=experimental&label=experimental", nil)
	labels := ParseLabelsParam(r)
	assert.Equal(t, 1, labels.Cardinality())
}

func TestParseLabelsParam_Labels_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs?labels=experimental,experimental", nil)
	labels := ParseLabelsParam(r)
	assert.Equal(t, 1, labels.Cardinality())
}

func TestParseLabelsParam_LabelsAndLabel_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs?labels=experimental&label=experimental", nil)
	labels := ParseLabelsParam(r)
	assert.Equal(t, 1, labels.Cardinality())
}
