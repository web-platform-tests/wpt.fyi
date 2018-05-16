// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/deckarep/golang-set"
)

// MaxCountDefaultValue is the default value returned by ParseMaxCountParam for the max-count param.
const MaxCountDefaultValue = 1

// MaxCountMaxValue is the maximum allowed value for the max-count param.
const MaxCountMaxValue = 500

// MaxCountMinValue is the minimum allowed value for the max-count param.
const MaxCountMinValue = 1

// SHARegex is a regex for SHA[0:10] slice of a git hash.
var SHARegex = regexp.MustCompile("[0-9a-fA-F]{10,40}")

// ParseSHAParam parses and validates the 'sha' param for the request,
// cropping it to 10 chars. It returns "latest" by default. (and in error cases).
func ParseSHAParam(r *http.Request) (runSHA string, err error) {
	sha, err := ParseSHAParamFull(r)
	if err != nil || !SHARegex.MatchString(sha) {
		return sha, err
	}
	return sha[:10], nil
}

// ParseSHAParamFull parses and validates the 'sha' param for the request.
// It returns "latest" by default (and in error cases).
func ParseSHAParamFull(r *http.Request) (runSHA string, err error) {
	// Get the SHA for the run being loaded (the first part of the path.)
	runSHA = "latest"
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return runSHA, err
	}

	runParam := params.Get("sha")
	if runParam != "" && runParam != "latest" {
		runSHA = runParam
		if !SHARegex.MatchString(runParam) {
			return "latest", fmt.Errorf("Invalid sha param value: %s", runParam)
		}
	}
	return runSHA, err
}

// ParseProductAtRevision parses a test-run spec into a ProductAtRevision struct.
func ParseProductAtRevision(spec string) (productAtRevision ProductAtRevision, err error) {
	pieces := strings.Split(spec, "@")
	if len(pieces) > 2 {
		return productAtRevision, errors.New("invalid product@revision spec: " + spec)
	}
	if productAtRevision.Product, err = ParseProduct(pieces[0]); err != nil {
		return productAtRevision, err
	}
	if len(pieces) < 2 {
		// No @ is assumed to be the product only.
		productAtRevision.Revision = "latest"
	} else {
		productAtRevision.Revision = pieces[1]
	}
	return productAtRevision, nil
}

// ParseProduct parses the `browser-version-os-version` input as a Product struct.
func ParseProduct(product string) (result Product, err error) {
	pieces := strings.Split(product, "-")
	if len(pieces) > 4 {
		return result, fmt.Errorf("invalid product: %s", product)
	}
	result = Product{
		BrowserName: pieces[0],
	}
	if !IsBrowserName(result.BrowserName) {
		return result, fmt.Errorf("invalid browser name: %s", result.BrowserName)
	}
	if len(pieces) > 1 {
		if _, err := ParseVersion(pieces[1]); err != nil {
			return result, fmt.Errorf("invalid browser version: %s", pieces[1])
		}
		result.BrowserVersion = pieces[1]
	}
	if len(pieces) > 2 {
		result.OSName = pieces[2]
	}
	if len(pieces) > 3 {
		if _, err := ParseVersion(pieces[3]); err != nil {
			return result, fmt.Errorf("invalid OS version: %s", pieces[3])
		}
		result.OSVersion = pieces[3]
	}
	return result, nil
}

// ParseVersion parses the given version as a semantically versioned string.
func ParseVersion(version string) (result *Version, err error) {
	pieces := strings.Split(version, ".")
	for i, piece := range pieces {
		if _, err := strconv.ParseInt(piece, 10, 0); i > 3 || err != nil {
			return nil, fmt.Errorf("Invalid version: %s", version)
		}
	}
	result = &Version{
		Major: pieces[0],
	}
	if len(pieces) > 1 {
		result.Minor = pieces[1]
	}
	if len(pieces) > 2 {
		result.Revision = pieces[2]
	}
	return result, nil
}

// ParseProductParam parses and validates the 'product' param for the request.
func ParseProductParam(r *http.Request) (product *Product, err error) {
	productParam := r.URL.Query().Get("product")
	if "" == productParam {
		return nil, nil
	}
	parsed, err := ParseProduct(productParam)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// ParseBrowserParam parses and validates the 'browser' param for the request.
// It returns "" by default (and in error cases).
func ParseBrowserParam(r *http.Request) (product *Product, err error) {
	browser := r.URL.Query().Get("browser")
	if "" == browser {
		return nil, nil
	}
	if IsBrowserName(browser) {
		return &Product{
			BrowserName: browser,
		}, nil
	}
	return nil, fmt.Errorf("Invalid browser param value: %s", browser)
}

// ParseBrowsersParam returns a list of browser params for the request.
// It parses the 'browsers' parameter, split on commas, and also checks for the (repeatable)
// 'browser' params.
func ParseBrowsersParam(r *http.Request) (browsers []string, err error) {
	browsers = r.URL.Query()["browser"]
	if browsersParam := r.URL.Query().Get("browsers"); browsersParam != "" {
		browsers = append(browsers, strings.Split(browsersParam, ",")...)
	}
	// Validate browser names.
	for i := 0; i < len(browsers); {
		if !IsBrowserName(browsers[i]) {
			if browsers[i] == "" {
				// 'Remove' empty browser by switching to end and cropping.
				browsers[len(browsers)-1], browsers[i] = browsers[i], browsers[len(browsers)-1]
				browsers = browsers[:len(browsers)-1]
				continue
			} else {
				return nil, fmt.Errorf("Invalid browser param value %s", browsers[i])
			}
		}
		i++
	}
	sort.Strings(browsers)
	return browsers, nil
}

// ParseProductsParam returns a list of product params for the request.
// It parses the 'products' parameter, split on commas, and also checks for the (repeatable)
// 'product' params.
func ParseProductsParam(r *http.Request) (products []Product, err error) {
	productParams := ParseRepeatedParam(r, "product", "products")
	if productParams == nil {
		return nil, nil
	}
	for p := range productParams.Iter() {
		product, err := ParseProduct(p.(string))
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

// GetProductsForRequest parses the 'products' (and legacy 'browsers') params, returning
// the list of products to include, or a default list.
func GetProductsForRequest(r *http.Request) (products []Product, err error) {
	if products, err = ParseProductsParam(r); err != nil {
		return nil, err
	}
	// Handle legacy browser param.
	browserParams, err := ParseBrowsersParam(r)
	if err != nil {
		return nil, err
	}
	for _, browser := range browserParams {
		products = append(products, Product{
			BrowserName: browser,
		})
	}

	browserNames, err := GetBrowserNames()
	// Fall back to default browser set.
	if products == nil && browserParams == nil {
		for _, name := range browserNames {
			products = append(products, Product{
				BrowserName: name,
			})
		}
	}

	labels := ParseLabelsParam(r)
	if labels != nil {
		experimental := labels.Contains(ExperimentalLabel)
		if err != nil {
			return nil, err
		}

		browserLabel := ""
		for _, name := range browserNames {
			if !labels.Contains(name) {
				continue
			}
			// If we already encountered a browser name, nothing is two browsers (return empty set).
			if browserLabel != "" {
				products = nil
				break
			}
			browserLabel = name
			products = []Product{
				Product{
					BrowserName: name,
				},
			}
			// For a browser label (e.g. "chrome"), we also include experimental, unless we explicitly only
			// want experimental, which is handled below.
			if !experimental {
				products = append(products, Product{
					BrowserName: name + "-" + ExperimentalLabel,
				})
			}
		}

		if experimental {
			for i := range products {
				products[i].BrowserName = products[i].BrowserName + "-" + ExperimentalLabel
			}
		}
	}

	return products, nil
}

// ParseMaxCountParam parses the 'max-count' parameter as an integer, or returns 1 if no param
// is present, or on error.
func ParseMaxCountParam(r *http.Request) (count int, err error) {
	return ParseMaxCountParamWithDefault(r, MaxCountDefaultValue)
}

// ParseMaxCountParamWithDefault parses the 'max-count' parameter as an integer, or returns the
// default when no param is present, or on error.
func ParseMaxCountParamWithDefault(r *http.Request, defaultValue int) (count int, err error) {
	count = defaultValue
	if maxCountParam := r.URL.Query().Get("max-count"); maxCountParam != "" {
		if count, err = strconv.Atoi(maxCountParam); err != nil {
			return defaultValue, err
		}
		if count < MaxCountMinValue {
			count = MaxCountMinValue
		}
		if count > MaxCountMaxValue {
			count = MaxCountMaxValue
		}
	}
	return count, err
}

// DiffFilterParam represents the types of changed test paths to include.
type DiffFilterParam struct {
	// Added tests are present in the 'after' state of the diff, but not present
	// in the 'before' state of the diff.
	Added bool

	// Deleted tests are present in the 'before' state of the diff, but not present
	// in the 'after' state of the diff.
	Deleted bool

	// Changed tests are present in both the 'before' and 'after' states of the diff,
	// but the number of passes, failures, or total tests has changed.
	Changed bool

	// Unchanged tests are present in both the 'before' and 'after' states of the diff,
	// and the number of passes, failures, or total tests is unchanged.
	Unchanged bool

	// Set of test paths to include, or include all tests if nil.
	Paths mapset.Set
}

// ParseDiffFilterParams collects the diff filtering params for the given request.
// It splits the filter param into the differences to include. The filter param is inspired by Git's --diff-filter flag.
// It also adds the set of test paths to include; see ParsePathsParam below.
func ParseDiffFilterParams(r *http.Request) (param DiffFilterParam, err error) {
	param = DiffFilterParam{
		Added:   true,
		Deleted: true,
		Changed: true,
	}
	if filter := r.URL.Query().Get("filter"); filter != "" {
		param = DiffFilterParam{}
		for _, char := range filter {
			switch char {
			case 'A':
				param.Added = true
			case 'D':
				param.Deleted = true
			case 'C':
				param.Changed = true
			case 'U':
				param.Unchanged = true
			default:
				return param, fmt.Errorf("invalid filter character %c", char)
			}
		}
	}
	param.Paths = ParsePathsParam(r)
	return param, nil
}

// ParsePathsParam returns a set list of test paths to include, or nil if no
// filter is provided (and all tests should be included). It parses the 'paths'
// parameter, split on commas, and also checks for the (repeatable) 'path' params
func ParsePathsParam(r *http.Request) (paths mapset.Set) {
	return ParseRepeatedParam(r, "path", "paths")
}

// ParseLabelsParam returns a set list of test-run labels to include, or nil if
// no labels are provided.
func ParseLabelsParam(r *http.Request) (labels mapset.Set) {
	return ParseRepeatedParam(r, "label", "labels")
}

// ParseRepeatedParam parses a param that may be a plural name, with all values
// comma-separated, or a repeated singular param.
// e.g. ?label=foo&label=bar vs ?labels=foo,bar
func ParseRepeatedParam(r *http.Request, singular string, plural string) (params mapset.Set) {
	repeatedParam := r.URL.Query()[singular]
	pluralParam := r.URL.Query().Get(plural)
	if len(repeatedParam) == 0 && pluralParam == "" {
		return nil
	}

	params = mapset.NewSet()
	for _, label := range repeatedParam {
		params.Add(label)
	}
	if pluralParam != "" {
		for _, label := range strings.Split(pluralParam, ",") {
			params.Add(label)
		}
	}
	if params.Contains("") {
		params.Remove("")
	}
	if params.Cardinality() == 0 {
		return nil
	}
	return params
}
