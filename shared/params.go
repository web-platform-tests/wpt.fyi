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
	"time"

	mapset "github.com/deckarep/golang-set"
)

// TestRunFilter represents the ways TestRun entities can be filtered in
// the webapp and api.
type TestRunFilter struct {
	SHA      string
	Labels   mapset.Set
	Complete *bool
	From     *time.Time
	MaxCount *int
	Products ProductSpecs
}

// IsDefaultQuery returns whether the params are just an empty query (or,
// the equivalent defaults of an empty query).
func (f TestRunFilter) IsDefaultQuery() bool {
	return IsLatest(f.SHA) &&
		(f.Labels == nil || f.Labels.Cardinality() < 1) &&
		(f.Complete == nil || *f.Complete) &&
		(f.From == nil) &&
		(f.MaxCount == nil || *f.MaxCount == 1) &&
		(len(f.Products) < 1)
}

// IsDefaultProducts returns whether the params products are empty, or the
// equivalent of the default product set.
func (f TestRunFilter) IsDefaultProducts() bool {
	if len(f.Products) == 0 {
		return true
	}
	def := GetDefaultProducts()
	if len(f.Products) != len(def) {
		return false
	}
	for i := range def {
		if def[i] != f.Products[i] {
			return false
		}
	}
	return true
}

// ProductSpec is a struct representing a parsed product spec string.
type ProductSpec struct {
	ProductAtRevision

	Labels mapset.Set
}

// ProductSpecs is a helper type for a slice of ProductSpec structs.
type ProductSpecs []ProductSpec

// Products gets the slice of products specified in the ProductSpecs slice.
func (p ProductSpecs) Products() []Product {
	result := make([]Product, len(p))
	for i, spec := range p {
		result[i] = spec.Product
	}
	return result
}

// Strings returns the array of the ProductSpec items as their string
// representations.
func (p ProductSpecs) Strings() []string {
	result := make([]string, len(p))
	for i, spec := range p {
		result[i] = spec.String()
	}
	return result
}

func (p ProductSpec) String() string {
	s := p.Product.String()
	if p.Labels != nil && p.Labels.Cardinality() > 0 {
		s += "["
		for label := range p.Labels.Iter() {
			s += label.(string) + ","
		}
		s = s[:len(s)-1]
		s += "]"
	}
	return s
}

func (p ProductSpecs) Len() int           { return len(p) }
func (p ProductSpecs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ProductSpecs) Less(i, j int) bool { return p[i].String() < p[j].String() }

// ToQuery converts the filter set to a url.Values (set of query params).
// completeIfDefault is whether the params should fall back to a complete run
// (the default on the homepage) if no conflicting params are used.
func (filter TestRunFilter) ToQuery(completeIfDefault bool) (q url.Values) {
	u := url.URL{}
	q = u.Query()
	if !IsLatest(filter.SHA) {
		q.Set("sha", filter.SHA)
	}
	if filter.Labels != nil && filter.Labels.Cardinality() > 0 {
		for label := range filter.Labels.Iter() {
			q.Add("label", label.(string))
		}
	}
	if len(filter.Products) > 0 {
		for _, p := range filter.Products {
			q.Add("product", p.String())
		}
	}
	if filter.Complete != nil {
		q.Set("complete", strconv.FormatBool(*filter.Complete))
	} else if completeIfDefault && len(q) == 0 {
		q.Set("complete", "true")
	}
	if filter.MaxCount != nil {
		q.Set("max-count", fmt.Sprintf("%v", *filter.MaxCount))
	}
	if filter.From != nil {
		q.Set("from", filter.From.Format(time.RFC3339))
	}
	return q
}

// MaxCountMaxValue is the maximum allowed value for the max-count param.
const MaxCountMaxValue = 500

// MaxCountMinValue is the minimum allowed value for the max-count param.
const MaxCountMinValue = 1

// SHARegex is a regex for SHA[0:10] slice of a git hash.
var SHARegex = regexp.MustCompile("[0-9a-fA-F]{10,40}")

// ErrMissing is the error returned when an expected parameter is missing.
var ErrMissing = errors.New("Missing parameter")

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

// ParseProductSpec parses a test-run spec into a ProductAtRevision struct.
func ParseProductSpec(spec string) (productSpec ProductSpec, err error) {
	errMsg := "invalid product spec: " + spec
	productSpec.Revision = "latest"
	name := spec
	// @sha (optional)
	atSHAPieces := strings.Split(spec, "@")
	if len(atSHAPieces) > 2 {
		return productSpec, errors.New(errMsg)
	} else if len(atSHAPieces) == 2 {
		name = atSHAPieces[0]
		productSpec.Revision = atSHAPieces[1]
	}
	// [foo,bar] labels syntax (optional)
	labelPieces := strings.Split(name, "[")
	if len(labelPieces) > 2 {
		return productSpec, errors.New(errMsg)
	} else if len(labelPieces) == 2 {
		name = labelPieces[0]
		labels := labelPieces[1]
		if labels[len(labels)-1:] != "]" || strings.Index(labels, "]") < len(labels)-1 {
			return productSpec, errors.New(errMsg)
		}
		labels = labels[:len(labels)-1]
		productSpec.Labels = mapset.NewSet()
		for _, label := range strings.Split(labels, ",") {
			if label != "" {
				productSpec.Labels.Add(label)
			}
		}
	}
	// Product (required)
	if productSpec.Product, err = ParseProduct(name); err != nil {
		return productSpec, err
	}
	return productSpec, nil
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
	if len(pieces) > 4 {
		return nil, fmt.Errorf("Invalid version: %s", version)
	}
	numbers := make([]int, len(pieces))
	for i, piece := range pieces {
		n, err := strconv.ParseInt(piece, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("Invalid version: %s", version)
		}
		numbers[i] = int(n)
	}
	result = &Version{
		Major: numbers[0],
	}
	if len(numbers) > 1 {
		result.Minor = numbers[1]
	}
	if len(numbers) > 2 {
		result.Build = numbers[2]
	}
	if len(numbers) > 3 {
		result.Revision = numbers[3]
	}
	return result, nil
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
	browserParams := ParseRepeatedParam(r, "browser", "browsers")
	if browserParams == nil {
		return nil, nil
	}
	for b := range browserParams.Iter() {
		if !IsBrowserName(b.(string)) {
			return nil, fmt.Errorf("Invalid browser param value %s", b.(string))
		}
		browsers = append(browsers, b.(string))
	}
	sort.Strings(browsers)
	return browsers, nil
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

// ParseProductsParam returns a list of product params for the request.
// It parses the 'products' parameter, split on commas, and also checks for the (repeatable)
// 'product' params.
func ParseProductsParam(r *http.Request) (products ProductSpecs, err error) {
	productParams := ParseRepeatedParam(r, "product", "products")
	if productParams == nil {
		return nil, nil
	}
	for p := range productParams.Iter() {
		product, err := ParseProductSpec(p.(string))
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	sort.Sort(products)
	return products, nil
}

// ParseProductOrBrowserParams parses the product (or, browser) params present in the given
// request.
func ParseProductOrBrowserParams(r *http.Request) (products ProductSpecs, err error) {
	if products, err = ParseProductsParam(r); err != nil {
		return nil, err
	}
	// Handle legacy browser param.
	browserParams, err := ParseBrowsersParam(r)
	if err != nil {
		return nil, err
	}
	for _, browser := range browserParams {
		spec := ProductSpec{}
		spec.BrowserName = browser
		products = append(products, spec)
	}
	return products, nil
}

// GetProductsOrDefault parses the 'products' (and legacy 'browsers') params, returning
// the sorted list of products to include, or a default list.
func (filter TestRunFilter) GetProductsOrDefault() (products ProductSpecs) {
	products = filter.Products
	// Fall back to default browser set.
	if products == nil {
		products = GetDefaultProducts()
	}
	sort.Sort(products)
	return products
}

// ParseMaxCountParam parses the 'max-count' parameter as an integer
func ParseMaxCountParam(r *http.Request) (*int, error) {
	if maxCountParam := r.URL.Query().Get("max-count"); maxCountParam != "" {
		count, err := strconv.Atoi(maxCountParam)
		if err != nil {
			return nil, err
		}
		if count < MaxCountMinValue {
			count = MaxCountMinValue
		}
		if count > MaxCountMaxValue {
			count = MaxCountMaxValue
		}
		return &count, nil
	}
	return nil, nil
}

// ParseMaxCountParamWithDefault parses the 'max-count' parameter as an integer, or returns the
// default when no param is present, or on error.
func ParseMaxCountParamWithDefault(r *http.Request, defaultValue int) (count int, err error) {
	if maxCountParam, err := ParseMaxCountParam(r); maxCountParam != nil {
		return *maxCountParam, err
	} else if err != nil {
		return defaultValue, err
	}
	return defaultValue, nil
}

// ParseFromParam parses the "from" param as a timestamp.
func ParseFromParam(r *http.Request) (*time.Time, error) {
	if fromParam := r.URL.Query().Get("from"); fromParam != "" {
		parsed, err := time.Parse(time.RFC3339, fromParam)
		if err != nil {
			return nil, err
		}
		return &parsed, nil
	}
	return nil, nil
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

// ParseQueryParamInt parses the URL query parameter at key. If the parameter is
// empty or missing, ErrMissing is returned.
func ParseQueryParamInt(r *http.Request, key string) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return 0, ErrMissing
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return i, fmt.Errorf("Invalid %s value: %s", key, value)
	}
	return i, err
}

// ParseCompleteParam parses the "complete" param. See ParseBooleanParam.
func ParseCompleteParam(r *http.Request) (complete *bool, err error) {
	return ParseBooleanParam(r, "complete")
}

// ParseBooleanParam parses the given param name as a bool.
// Return nil if the param is missing, true if if it's present with no value,
// otherwise the parsed boolean value of the param's value.
func ParseBooleanParam(r *http.Request, name string) (result *bool, err error) {
	q := r.URL.Query()
	b := false
	if _, ok := q[name]; !ok {
		return nil, nil
	} else if val := q.Get(name); val == "" {
		b = true
	} else {
		b, err = strconv.ParseBool(val)
	}
	return &b, err
}

// ParseTestRunFilterParams parses all of the filter params for a TestRun query.
func ParseTestRunFilterParams(r *http.Request) (filter TestRunFilter, err error) {
	runSHA, err := ParseSHAParam(r)
	if err != nil {
		return filter, err
	}
	filter.SHA = runSHA
	filter.Labels = ParseLabelsParam(r)
	if filter.Complete, err = ParseCompleteParam(r); err != nil {
		return filter, err
	}
	if filter.Products, err = ParseProductOrBrowserParams(r); err != nil {
		return filter, err
	}
	if filter.MaxCount, err = ParseMaxCountParam(r); err != nil {
		return filter, err
	}
	if filter.From, err = ParseFromParam(r); err != nil {
		return filter, err
	}
	return filter, nil
}
