// +build small

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseSHAParam(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/", nil)
	runSHA, err := ParseSHAParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Equal(t, "latest", runSHA)
}

func TestParseSHAParam_Latest(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?sha=latest", nil)
	runSHA, err := ParseSHAParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Equal(t, "latest", runSHA)
}

func TestParseSHAParam_ShortSHA(t *testing.T) {
	sha := "0123456789"
	r := httptest.NewRequest("GET", "http://wpt.fyi/?sha="+sha, nil)
	runSHA, err := ParseSHAParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Equal(t, sha, runSHA)
}

func TestParseSHAParam_FullSHA(t *testing.T) {
	sha := "0123456789aaaaabbbbbcccccdddddeeeeefffff"
	r := httptest.NewRequest("GET", "http://wpt.fyi/?sha="+sha, nil)
	runSHA, err := ParseSHAParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Equal(t, sha, runSHA)
}

func TestParseSHAParam_NonSHA(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?sha=123", nil)
	_, err := ParseSHAParam(r.URL.Query())
	assert.NotNil(t, err)
}

func TestParseSHAParam_NonSHA_2(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?sha=zapper0123", nil)
	_, err := ParseSHAParam(r.URL.Query())
	assert.NotNil(t, err)
}

func TestParseBrowserParam(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/", nil)
	browser, err := ParseBrowserParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Nil(t, browser)
}

func TestParseBrowserParam_Chrome(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browser=chrome", nil)
	browser, err := ParseBrowserParam(r.URL.Query())
	assert.Nil(t, err)
	assert.NotNil(t, browser)
	assert.Equal(t, "chrome", browser.BrowserName)
}

func TestParseBrowserParam_Invalid(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browser=invalid", nil)
	browser, err := ParseBrowserParam(r.URL.Query())
	assert.NotNil(t, err)
	assert.Nil(t, browser)
}

func TestGetProductsOrDefault_Default(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/", nil)
	filters, err := ParseTestRunFilterParams(r.URL.Query())
	products := filters.GetProductsOrDefault()
	assert.Nil(t, err)
	defaultBrowsers := GetDefaultBrowserNames()
	assert.Equal(t, len(defaultBrowsers), len(products))
	for i := range defaultBrowsers {
		assert.Equal(t, defaultBrowsers[i], products[i].BrowserName)
	}
}

func TestGetProductsOrDefault_BrowserParam_ChromeSafari(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browsers=chrome,safari", nil)
	filter, err := ParseTestRunFilterParams(r.URL.Query())
	browsers := filter.GetProductsOrDefault()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(browsers))
	assert.Equal(t, "chrome", browsers[0].BrowserName)
	assert.Equal(t, "safari", browsers[1].BrowserName)
}

func TestGetProductsOrDefault_BrowserParam_ChromeInvalid(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browsers=chrome,invalid", nil)
	_, err := ParseTestRunFilterParams(r.URL.Query())
	assert.NotNil(t, err)
}

func TestGetProductsOrDefault_BrowserParam_EmptyCommas(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browsers=,edge,,,,chrome,,", nil)
	filters, err := ParseTestRunFilterParams(r.URL.Query())
	products := filters.GetProductsOrDefault()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(products))
	assert.Equal(t, "edge", products[0].BrowserName)
	assert.Equal(t, "chrome", products[1].BrowserName)
}

func TestGetProductsOrDefault_BrowserParam_SafariChrome(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browsers=safari,chrome", nil)
	filters, err := ParseTestRunFilterParams(r.URL.Query())
	products := filters.GetProductsOrDefault()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(products))
	assert.Equal(t, "safari", products[0].BrowserName)
	assert.Equal(t, "chrome", products[1].BrowserName)
}

func TestGetProductsOrDefault_BrowserParam_MultiBrowserParam_SafariChrome(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browser=safari&browser=chrome", nil)
	filters, err := ParseTestRunFilterParams(r.URL.Query())
	products := filters.GetProductsOrDefault()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(products))
	assert.Equal(t, "safari", products[0].BrowserName)
	assert.Equal(t, "chrome", products[1].BrowserName)
}

func TestGetProductsOrDefault_BrowserParam_MultiBrowserParam_SafariInvalid(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?browser=safari&browser=invalid", nil)
	_, err := ParseTestRunFilterParams(r.URL.Query())
	assert.NotNil(t, err)
}

func TestGetProductsOrDefault_BrowserAndProductParam(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?product=edge-16&browser=chrome", nil)
	filters, err := ParseTestRunFilterParams(r.URL.Query())
	products := filters.GetProductsOrDefault()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(products))
	assert.Equal(t, "edge", products[0].BrowserName)
	assert.Equal(t, "16", products[0].BrowserVersion)
	assert.Equal(t, "chrome", products[1].BrowserName)
}

func TestGetProductsOrDefault_BrowsersAndProductsParam(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?products=edge-16,safari&browsers=chrome,firefox", nil)
	filters, err := ParseTestRunFilterParams(r.URL.Query())
	products := filters.GetProductsOrDefault()
	assert.Nil(t, err)
	assert.Equal(t, 4, len(products))
	assert.Equal(t, "edge", products[0].BrowserName)
	assert.Equal(t, "16", products[0].BrowserVersion)
	assert.Equal(t, "safari", products[1].BrowserName)
	assert.Equal(t, "chrome", products[2].BrowserName)
	assert.Equal(t, "firefox", products[3].BrowserName)
}

func TestParseMaxCountParam_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/", nil)
	count, err := ParseMaxCountParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Nil(t, count)

	d, err := ParseMaxCountParamWithDefault(r.URL.Query(), 5)
	assert.Nil(t, err)
	assert.Equal(t, 5, d)
}

func TestParseMaxCountParam_TooSmall(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?max-count=0", nil)
	count, err := ParseMaxCountParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Equal(t, MaxCountMinValue, *count)

	r = httptest.NewRequest("GET", "http://wpt.fyi/?max-count=-1", nil)
	count, err = ParseMaxCountParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Equal(t, MaxCountMinValue, *count)
}

func TestParseMaxCountParam_TooLarge(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?max-count=123456789", nil)
	count, err := ParseMaxCountParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Equal(t, MaxCountMaxValue, *count)

	r = httptest.NewRequest("GET", "http://wpt.fyi/?max-count=100000000", nil)
	count, err = ParseMaxCountParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Equal(t, MaxCountMaxValue, *count)
}

func TestParseMaxCountParam(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?max-count=2", nil)
	count, err := ParseMaxCountParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Equal(t, 2, *count)
}

func TestParsePathsParam_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff", nil)
	paths := ParsePathsParam(r.URL.Query())
	assert.Nil(t, paths)
}

func TestParsePathsParam_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?path=", nil)
	paths := ParsePathsParam(r.URL.Query())
	assert.Nil(t, paths)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?paths=", nil)
	paths = ParsePathsParam(r.URL.Query())
	assert.Nil(t, paths)
}

func TestParsePathsParam_Path_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?path=/css&path=/css", nil)
	paths := ParsePathsParam(r.URL.Query())
	assert.Len(t, paths, 1)
}

func TestParsePathsParam_Paths_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?paths=/css,/css", nil)
	paths := ParsePathsParam(r.URL.Query())
	assert.Len(t, paths, 1)
}

func TestParsePathsParam_PathsAndPath_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?paths=/css&path=/css", nil)
	paths := ParsePathsParam(r.URL.Query())
	assert.Len(t, paths, 1)
}

func TestParsePathsParam_Paths_DiffFilter(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?paths=/css&path=/css", nil)
	_, paths, err := ParseDiffFilterParams(r.URL.Query())
	assert.Nil(t, err)
	assert.Equal(t, 1, paths.Cardinality())
}

func TestParseDiffFilterParam(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=A", nil)
	filter, _, _ := ParseDiffFilterParams(r.URL.Query())
	assert.Equal(t, DiffFilterParam{Added: true, Deleted: false, Changed: false}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=D", nil)
	filter, _, _ = ParseDiffFilterParams(r.URL.Query())
	assert.Equal(t, DiffFilterParam{Added: false, Deleted: true, Changed: false}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=C", nil)
	filter, _, _ = ParseDiffFilterParams(r.URL.Query())
	assert.Equal(t, DiffFilterParam{Added: false, Deleted: false, Changed: true}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=CAD", nil)
	filter, _, _ = ParseDiffFilterParams(r.URL.Query())
	assert.Equal(t, DiffFilterParam{Added: true, Deleted: true, Changed: true}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=CD", nil)
	filter, _, _ = ParseDiffFilterParams(r.URL.Query())
	assert.Equal(t, DiffFilterParam{Added: false, Deleted: true, Changed: true}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=CACA", nil)
	filter, _, _ = ParseDiffFilterParams(r.URL.Query())
	assert.Equal(t, DiffFilterParam{Added: true, Deleted: false, Changed: true}, filter)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=U", nil)
	filter, _, _ = ParseDiffFilterParams(r.URL.Query())
	assert.Equal(t, DiffFilterParam{Unchanged: true}, filter)
}

func TestParseDiffFilterParam_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff", nil)
	filter, _, err := ParseDiffFilterParams(r.URL.Query())
	assert.Nil(t, err)
	assert.Equal(t, DiffFilterParam{Added: true, Deleted: true, Changed: true, Unchanged: false}, filter)
}

func TestParseDiffFilterParam_Invalid(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/diff?filter=Z", nil)
	_, _, err := ParseDiffFilterParams(r.URL.Query())
	assert.NotNil(t, err)
}

func TestParseLabelsParam_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs", nil)
	labels := ParseLabelsParam(r.URL.Query())
	assert.Nil(t, labels)
}

func TestParseLabelsParam_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs?label=", nil)
	labels := ParseLabelsParam(r.URL.Query())
	assert.Nil(t, labels)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/runs?labels=", nil)
	labels = ParseLabelsParam(r.URL.Query())
	assert.Nil(t, labels)
}

func TestParseLabelsParam_Label_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs?label=experimental&label=experimental", nil)
	labels := ParseLabelsParam(r.URL.Query())
	assert.Len(t, labels, 1)
}

func TestParseLabelsParam_Labels_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs?labels=experimental,experimental", nil)
	labels := ParseLabelsParam(r.URL.Query())
	assert.Len(t, labels, 1)
}

func TestParseLabelsParam_LabelsAndLabel_Duplicate(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs?labels=experimental&label=experimental", nil)
	labels := ParseLabelsParam(r.URL.Query())
	assert.Len(t, labels, 1)
}

func TestParseVersion(t *testing.T) {
	v, err := ParseVersion("63.0")
	assert.Nil(t, err)
	assert.Equal(t, 63, v.Major)
	assert.Equal(t, 0, *v.Minor)
	assert.Nil(t, v.Build)
	assert.Nil(t, v.Revision)
	assert.Empty(t, v.Channel)

	// FF
	v, err = ParseVersion("65.0a1")
	assert.Nil(t, err)
	assert.Equal(t, 65, v.Major)
	assert.Equal(t, 0, *v.Minor)
	assert.Nil(t, v.Build)
	assert.Nil(t, v.Revision)
	assert.Equal(t, "a1", v.Channel)

	// Chrome
	v, err = ParseVersion("71.0.3578.20 dev")
	assert.Nil(t, err)
	assert.Equal(t, 71, v.Major)
	assert.Equal(t, 0, *v.Minor)
	assert.Equal(t, 3578, *v.Build)
	assert.Equal(t, 20, *v.Revision)
	assert.Equal(t, " dev", v.Channel)
}

func TestParseProductSpec(t *testing.T) {
	productSpec, err := ParseProductSpec("chrome@latest")
	assert.Nil(t, err)
	assert.Equal(t, "chrome", productSpec.BrowserName)
	assert.Equal(t, "latest", productSpec.Revision)

	productSpec, err = ParseProductSpec("edge")
	assert.Nil(t, err)
	assert.Equal(t, "edge", productSpec.BrowserName)
}

func TestParseProductSpec_FullSHA(t *testing.T) {
	sha := "0123456789aaaaabbbbbcccccdddddeeeeefffff"
	r := httptest.NewRequest("GET", "http://wpt.fyi/?product=chrome@"+sha, nil)
	filters, err := ParseTestRunFilterParams(r.URL.Query())
	assert.Nil(t, err)
	products := filters.GetProductsOrDefault()
	assert.Len(t, products, 1)
	if len(products) > 0 {
		assert.Equal(t, sha, products[0].Revision)
	}
}

func TestParseProductSpec_BrowserVersion(t *testing.T) {
	productSpec, err := ParseProductSpec("chrome-63.0@latest")
	assert.Nil(t, err)
	assert.Equal(t, "chrome", productSpec.BrowserName)
	assert.Equal(t, "63.0", productSpec.BrowserVersion)
	assert.Equal(t, "latest", productSpec.Revision)
}

func TestParseProductSpec_OS(t *testing.T) {
	productSpec, err := ParseProductSpec("chrome-63.0-linux@latest")
	assert.Nil(t, err)
	assert.Equal(t, "chrome", productSpec.BrowserName)
	assert.Equal(t, "63.0", productSpec.BrowserVersion)
	assert.Equal(t, "linux", productSpec.OSName)
	assert.Equal(t, "latest", productSpec.Revision)
}

func TestParseProductSpec_OSVersion(t *testing.T) {
	productSpec, err := ParseProductSpec("chrome-63.0-linux-4.4@latest")
	assert.Nil(t, err)
	assert.Equal(t, "chrome", productSpec.BrowserName)
	assert.Equal(t, "63.0", productSpec.BrowserVersion)
	assert.Equal(t, "linux", productSpec.OSName)
	assert.Equal(t, "4.4", productSpec.OSVersion)
	assert.Equal(t, "latest", productSpec.Revision)
}

func TestParseProductSpec_Labels(t *testing.T) {
	productSpec, err := ParseProductSpec("chrome[foo,bar]")
	assert.Nil(t, err)
	assert.Equal(t, "chrome", productSpec.BrowserName)
	assert.True(t, productSpec.Labels.Contains("foo"))
	assert.True(t, productSpec.Labels.Contains("bar"))

	productSpec, err = ParseProductSpec("chrome[foo]@1234512345")
	assert.Nil(t, err)
	assert.Equal(t, "chrome", productSpec.BrowserName)
	assert.True(t, productSpec.Labels.Contains("foo"))
	assert.Equal(t, "1234512345", productSpec.Revision)

	productSpec, err = ParseProductSpec("chrome[]")
	assert.Nil(t, err)
	assert.Equal(t, "chrome", productSpec.BrowserName)
	assert.Equal(t, 0, productSpec.Labels.Cardinality())

	_, err = ParseProductSpec("chrome[foo")
	assert.NotNil(t, err)
	_, err = ParseProductSpec("chrome[foo][bar]")
	assert.NotNil(t, err)
	_, err = ParseProductSpec("[foo]")
	assert.NotNil(t, err)
	_, err = ParseProductSpec("chrome[")
	assert.NotNil(t, err)
	_, err = ParseProductSpec("chrome[foo")
	assert.NotNil(t, err)
}

func TestParseProductSpec_String(t *testing.T) {
	productSpec, err := ParseProductSpec("chrome-64[foo,bar]@1234512345")
	assert.Nil(t, err)
	assert.Equal(t, "chrome-64[bar,foo]@1234512345", productSpec.String())
}

func TestParseProductSpec_Plural(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs?products=chrome[stable],chrome[experimental]", nil)
	products, err := ParseProductsParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, "chrome[stable]", products[0].String())
	assert.Equal(t, "chrome[experimental]", products[1].String())

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/runs?products=chrome[foo,bar,baz],chrome[qux]", nil)
	products, err = ParseProductsParam(r.URL.Query())
	assert.Nil(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, "chrome[bar,baz,foo]", products[0].String()) // Labels alphabeticized.
	assert.Equal(t, "chrome[qux]", products[1].String())
}

func TestParseAligned(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/runs", nil)
	aligned, _ := ParseAlignedParam(r.URL.Query())
	assert.Nil(t, aligned)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/runs?aligned", nil)
	aligned, _ = ParseAlignedParam(r.URL.Query())
	assert.True(t, *aligned)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/runs?aligned=true", nil)
	aligned, _ = ParseAlignedParam(r.URL.Query())
	assert.True(t, *aligned)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/runs?aligned=false", nil)
	aligned, _ = ParseAlignedParam(r.URL.Query())
	assert.False(t, *aligned)
}

func TestParseRunIDsParam_nil(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/search", nil)
	runIDs, err := ParseRunIDsParam(r.URL.Query())
	assert.Nil(t, runIDs)
	assert.Nil(t, err)
}

func TestParseRunIDsParam_ok(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/search?run_ids=1,2,3", nil)
	runIDs, err := ParseRunIDsParam(r.URL.Query())
	assert.Equal(t, []int64{1, 2, 3}, []int64(runIDs))
	assert.Nil(t, err)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/search?run_id=1&run_id=2&run_id=3", nil)
	runIDs, err = ParseRunIDsParam(r.URL.Query())
	assert.Equal(t, []int64{1, 2, 3}, []int64(runIDs))
	assert.Nil(t, err)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/search?run_id=1&run_id=2&run_id=3", nil)
	runIDs, err = ParseRunIDsParam(r.URL.Query())
	assert.Equal(t, []int64{1, 2, 3}, []int64(runIDs))
	assert.Nil(t, err)
}

func TestParseRunIDsParam_err(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/search?run_ids=1,notanumber,3", nil)
	runIDs, err := ParseRunIDsParam(r.URL.Query())
	assert.Nil(t, runIDs)
	assert.NotNil(t, err)

	r = httptest.NewRequest("GET", "http://wpt.fyi/api/search?run_id=1&run_id=notanumber&run_id=3", nil)
	runIDs, err = ParseRunIDsParam(r.URL.Query())
	assert.Nil(t, runIDs)
	assert.NotNil(t, err)
}

func TestParseQueryFilterParams_nil(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/search", nil)
	filter, err := ParseQueryFilterParams(r.URL.Query())
	assert.Equal(t, QueryFilter{}, filter)
	assert.Nil(t, err)
}

func TestParseQueryFilterParams_runIDs(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/search?run_ids=1,2,3", nil)
	filter, err := ParseQueryFilterParams(r.URL.Query())
	assert.Equal(t, QueryFilter{
		RunIDs: []int64{1, 2, 3},
	}, filter)
	assert.Nil(t, err)
}

func TestParseQueryFilterParams_q(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/search?q=abcd", nil)
	filter, err := ParseQueryFilterParams(r.URL.Query())
	assert.Equal(t, QueryFilter{
		Q: "abcd",
	}, filter)
	assert.Nil(t, err)
}

func TestParseQueryFilterParams_aligned(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/search?run_ids=1,2,3&q=abcd", nil)
	filter, err := ParseQueryFilterParams(r.URL.Query())
	assert.Equal(t, QueryFilter{
		RunIDs: []int64{1, 2, 3},
		Q:      "abcd",
	}, filter)
	assert.Nil(t, err)
}

func TestParseQueryFilterParams_err(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/api/search?run_ids=1,notanumber,3&q=abcd", nil)
	_, err := ParseQueryFilterParams(r.URL.Query())
	assert.NotNil(t, err)
}

func TestParseTestRunFilterParams(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/", nil)
	filter, _ := ParseTestRunFilterParams(r.URL.Query())
	assert.Nil(t, filter.Aligned)
	assert.Equal(t, "aligned=true&label=stable", filter.OrDefault().ToQuery().Encode())
	assert.Equal(t, "", filter.ToQuery().Encode())

	r = httptest.NewRequest("GET", "http://wpt.fyi/?label=stable", nil)
	filter, _ = ParseTestRunFilterParams(r.URL.Query())
	assert.Equal(t, "label=stable", filter.OrDefault().ToQuery().Encode())
	assert.Equal(t, "label=stable", filter.ToQuery().Encode())

	r = httptest.NewRequest("GET", "http://wpt.fyi/?from=2018-01-01T00%3A00%3A00Z", nil)
	filter, _ = ParseTestRunFilterParams(r.URL.Query())
	assert.Equal(t, "from=2018-01-01T00%3A00%3A00Z", filter.ToQuery().Encode())
}

func TestParseTestRunFilterParams_Invalid(t *testing.T) {
	r := httptest.NewRequest("GET", "http://wpt.fyi/?product=chrome%5B", nil)
	_, err := ParseTestRunFilterParams(r.URL.Query())
	assert.NotNil(t, err)
}

func TestProductSpecMatches(t *testing.T) {
	chrome, err := ParseProductSpec("chrome")
	assert.Nil(t, err)

	chromeRun := TestRun{}
	chromeRun.BrowserName = "chrome"
	chromeRun.BrowserVersion = "63.123"
	assert.True(t, chrome.Matches(chromeRun))

	chrome6, err := ParseProductSpec("chrome-6")
	assert.False(t, chrome6.Matches(chromeRun))
	chrome63, err := ParseProductSpec("chrome-63")
	assert.True(t, chrome63.Matches(chromeRun))

	safariRun := TestRun{}
	safariRun.BrowserName = "safari"
	assert.False(t, chrome.Matches(safariRun))
}

func TestProductSpecMatches_Labels(t *testing.T) {
	chrome, err := ParseProductSpec("chrome[foo]")
	assert.Nil(t, err)

	chromeRun := TestRun{}
	chromeRun.BrowserName = "chrome"
	assert.False(t, chrome.Matches(chromeRun))
	chromeRun.Labels = []string{"bar", "foo"}
	assert.True(t, chrome.Matches(chromeRun))
}

func TestProductSpecMatches_Revision(t *testing.T) {
	revision := "abcdef0123"
	version := "69.1.1.1"
	chrome, err := ParseProductSpec(fmt.Sprintf("chrome-%s@%s", version, revision))
	assert.Nil(t, err)

	chromeRun := TestRun{}
	chromeRun.BrowserName = "chrome"
	chromeRun.BrowserVersion = "69.1.1.0"
	chromeRun.Revision = "1234567890"
	assert.False(t, chrome.Matches(chromeRun))
	chromeRun.Revision = revision
	assert.False(t, chrome.Matches(chromeRun)) // Still wrong version
	chromeRun.BrowserVersion = version
	assert.True(t, chrome.Matches(chromeRun))
}

func TestParsePageToken(t *testing.T) {
	filter := TestRunFilter{}
	now := time.Now().Truncate(time.Second)
	filter.To = &now

	token, err := filter.Token()
	assert.Nil(t, err)
	r := httptest.NewRequest("GET", "/?page="+token, nil)

	parsed, err := ParsePageToken(r.URL.Query())
	assert.Nil(t, err)
	if parsed == nil {
		assert.FailNow(t, "Parsed page token was nil")
	} else if parsed.To == nil {
		assert.FailNow(t, "Parsed page token has no 'to' param")
	}
	assert.True(t, filter.To.Equal(*parsed.To))
}

func TestExtractRunIDsBodyParam_ParseError(t *testing.T) {
	payload := []byte("}{")
	_, err := ExtractRunIDsBodyParam(httptest.NewRequest("POST", "https://wpt.fyi/api/search", bytes.NewBuffer(payload)), false)
	assert.NotNil(t, err)
	_, err = ExtractRunIDsBodyParam(httptest.NewRequest("POST", "https://wpt.fyi/api/search", bytes.NewBuffer(payload)), true)
	assert.NotNil(t, err)
}

func TestExtractRunIDsBodyParam_MissingRunIDs(t *testing.T) {
	payload := []byte("{}")
	_, err := ExtractRunIDsBodyParam(httptest.NewRequest("POST", "https://wpt.fyi/api/search", bytes.NewBuffer(payload)), false)
	assert.NotNil(t, err)
	_, err = ExtractRunIDsBodyParam(httptest.NewRequest("POST", "https://wpt.fyi/api/search", bytes.NewBuffer(payload)), true)
	assert.NotNil(t, err)
}

func TestExtractRunIDsBodyParam_NonInt(t *testing.T) {
	arrayContents := []string{
		`"42"`,
		`true`,
		`[]`,
	}
	for _, contents := range arrayContents {
		payload := []byte(fmt.Sprintf(`{"run_ids":[%s]}`, contents))
		_, err := ExtractRunIDsBodyParam(httptest.NewRequest("POST", "https://wpt.fyi/api/search", bytes.NewBuffer(payload)), false)
		assert.NotNil(t, err)
		_, err = ExtractRunIDsBodyParam(httptest.NewRequest("POST", "https://wpt.fyi/api/search", bytes.NewBuffer(payload)), true)
		assert.NotNil(t, err)
	}
}

func TestExtractRunIDsBodyParam_OK(t *testing.T) {
	payload := []byte(`{"run_ids":[1,2,-3]}`)
	expected := TestRunIDs([]int64{1, 2, -3})
	runIDs, err := ExtractRunIDsBodyParam(httptest.NewRequest("POST", "https://wpt.fyi/api/search", bytes.NewBuffer(payload)), false)
	assert.Nil(t, err)
	assert.Equal(t, expected, runIDs)
	_, err = ExtractRunIDsBodyParam(httptest.NewRequest("POST", "https://wpt.fyi/api/search", bytes.NewBuffer(payload)), true)
	assert.Nil(t, err)
	assert.Equal(t, expected, runIDs)
}

func TestExtractRunIDsBodyParam_Replayable(t *testing.T) {
	payload := []byte(`{"run_ids":[1,2,-3]}`)
	req := httptest.NewRequest("POST", "https://wpt.fyi/api/search", bytes.NewBuffer(payload))
	ExtractRunIDsBodyParam(req, true)
	replayed, err := ioutil.ReadAll(req.Body)
	assert.Nil(t, err)
	assert.Equal(t, payload, replayed)
}
