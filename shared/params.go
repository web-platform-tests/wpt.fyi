// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
)

// QueryFilter represents the ways search results can be filtered in the webapp
// search API.
type QueryFilter struct {
	RunIDs []int64
	Q      string
}

// ProductSpec is a struct representing a parsed product spec string.
type ProductSpec struct {
	ProductAtRevision

	Labels mapset.Set
}

// Matches returns whether the spec matches the given run.
func (p ProductSpec) Matches(run TestRun) bool {
	if run.BrowserName != p.BrowserName {
		return false
	}
	if !IsLatest(p.Revision) && p.Revision != run.Revision {
		return false
	}
	if p.Labels != nil && p.Labels.Cardinality() > 0 {
		runLabels := run.LabelsSet()
		if !p.Labels.IsSubset(runLabels) {
			return false
		}
	}
	if p.BrowserVersion != "" {
		// Make "6" not match "60.123" by adding trailing dots to both.
		if !strings.HasPrefix(run.BrowserVersion+".", p.BrowserVersion+".") {
			return false
		}
	}
	return true
}

// IsExperimental returns true if the product spec is restricted to experimental
// runs (i.e. has the label "experimental").
func (p ProductSpec) IsExperimental() bool {
	return p.Labels != nil && p.Labels.Contains(ExperimentalLabel)
}

// DisplayName returns a capitalized version of the product's name.
func (p ProductSpec) DisplayName() string {
	switch p.BrowserName {
	case "chrome":
		return "Chrome"
	case "edge":
		return "Edge"
	case "firefox":
		return "Firefox"
	case "safari":
		return "Safari"
	default:
		return p.BrowserName
	}
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

// OrDefault returns the current product specs, or the default if the set is empty.
func (p ProductSpecs) OrDefault() ProductSpecs {
	if len(p) < 1 {
		return GetDefaultProducts()
	}
	return p
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
		labels := make([]string, 0, p.Labels.Cardinality())
		for l := range p.Labels.Iter() {
			labels = append(labels, l.(string))
		}
		sort.Strings(labels) // Deterministic String() output.
		s += "[" + strings.Join(labels, ",") + "]"
	}
	if !IsLatest(p.Revision) {
		s += "@" + p.Revision
	}
	return s
}

func (p ProductSpecs) Len() int           { return len(p) }
func (p ProductSpecs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ProductSpecs) Less(i, j int) bool { return p[i].String() < p[j].String() }

// MaxCountMaxValue is the maximum allowed value for the max-count param.
const MaxCountMaxValue = 500

// MaxCountMinValue is the minimum allowed value for the max-count param.
const MaxCountMinValue = 1

// SHARegex is a regex for SHA[0:10] slice of a git hash.
var SHARegex = regexp.MustCompile("[0-9a-fA-F]{10,40}")

// ParseSHAParam parses and validates the 'sha' param for the request,
// cropping it to 10 chars. It returns "latest" by default. (and in error cases).
func ParseSHAParam(v url.Values) (runSHA string, err error) {
	sha, err := ParseSHAParamFull(v)
	if err != nil || !SHARegex.MatchString(sha) {
		return sha, err
	}
	return sha[:10], nil
}

// ParseSHA validates the given 'sha' value, cropping it to 10 chars.
// It returns "latest" by default (and in error cases).
func ParseSHA(sha string) (runSHA string, err error) {
	sha, err = ParseSHAFull(sha)
	if err != nil || !SHARegex.MatchString(sha) {
		return sha, err
	}
	return sha[:10], nil
}

// ParseSHAParamFull parses and validates the 'sha' param for the request.
// It returns "latest" by default (and in error cases).
func ParseSHAParamFull(v url.Values) (runSHA string, err error) {
	// Get the SHA for the run being loaded (the first part of the path.)
	return ParseSHAFull(v.Get("sha"))
}

// ParseSHAFull parses and validates the given 'sha'.
// It returns "latest" by default (and in error cases).
func ParseSHAFull(runParam string) (runSHA string, err error) {
	// Get the SHA for the run being loaded (the first part of the path.)
	runSHA = "latest"
	if runParam != "" && runParam != "latest" {
		runSHA = runParam
		if !SHARegex.MatchString(runParam) {
			return "latest", fmt.Errorf("Invalid sha param value: %s", runParam)
		}
	}
	return runSHA, err
}

// ParseProductSpecs parses multiple product specs
func ParseProductSpecs(specs ...string) (products ProductSpecs, err error) {
	products = make(ProductSpecs, len(specs))
	for i, p := range specs {
		product, err := ParseProductSpec(p)
		if err != nil {
			return nil, err
		}
		products[i] = product
	}
	return products, nil
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
		if productSpec.Revision, err = ParseSHA(atSHAPieces[1]); err != nil {
			return productSpec, errors.New(errMsg)
		}
	}
	// [foo,bar] labels syntax (optional)
	labelPieces := strings.Split(name, "[")
	if len(labelPieces) > 2 {
		return productSpec, errors.New(errMsg)
	} else if len(labelPieces) == 2 {
		name = labelPieces[0]
		labels := labelPieces[1]
		if labels == "" {
			return productSpec, errors.New(errMsg)
		}
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
	pieces := strings.Split(version, " ")
	channel := ""
	if len(pieces) > 2 {
		return nil, fmt.Errorf("Invalid version: %s", version)
	} else if len(pieces) > 1 {
		channel = " " + pieces[1]
		version = pieces[0]
	}

	// Special case ff's "a1" suffix
	ffSuffix := regexp.MustCompile(`^.*([ab]\d+)$`)
	if match := ffSuffix.FindStringSubmatch(version); match != nil {
		channel = match[1]
		version = version[:len(version)-len(channel)]
	}

	pieces = strings.Split(version, ".")
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
		Major:   numbers[0],
		Channel: channel,
	}
	if len(numbers) > 1 {
		result.Minor = &numbers[1]
	}
	if len(numbers) > 2 {
		result.Build = &numbers[2]
	}
	if len(numbers) > 3 {
		result.Revision = &numbers[3]
	}
	return result, nil
}

// ParseBrowserParam parses and validates the 'browser' param for the request.
// It returns "" by default (and in error cases).
func ParseBrowserParam(v url.Values) (product *Product, err error) {
	browser := v.Get("browser")
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
func ParseBrowsersParam(v url.Values) (browsers []string, err error) {
	browserParams := ParseRepeatedParam(v, "browser", "browsers")
	if browserParams == nil {
		return nil, nil
	}
	for _, b := range browserParams {
		if !IsBrowserName(b) {
			return nil, fmt.Errorf("Invalid browser param value %s", b)
		}
		browsers = append(browsers, b)
	}
	return browsers, nil
}

// ParseProductParam parses and validates the 'product' param for the request.
func ParseProductParam(v url.Values) (product *ProductSpec, err error) {
	productParam := v.Get("product")
	if "" == productParam {
		return nil, nil
	}
	parsed, err := ParseProductSpec(productParam)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// ParseProductsParam returns a list of product params for the request.
// It parses the 'products' parameter, split on commas, and also checks for the (repeatable)
// 'product' params.
func ParseProductsParam(v url.Values) (ProductSpecs, error) {
	repeatedParam := v["product"]
	pluralParam := v.Get("products")
	// Replace nested ',' in the label part with a placeholder
	nestedCommas := regexp.MustCompile(`(\[[^\]]*),`)
	const comma = `%COMMA%`
	for nestedCommas.MatchString(pluralParam) {
		pluralParam = nestedCommas.ReplaceAllString(pluralParam, "$1"+comma)
	}
	productParams := parseRepeatedParamValues(repeatedParam, pluralParam)
	if productParams == nil {
		return nil, nil
	}
	// Revert placeholder to ',' and parse.
	for i := range productParams {
		productParams[i] = strings.Replace(productParams[i], comma, ",", -1)
	}
	return ParseProductSpecs(productParams...)
}

// ParseProductOrBrowserParams parses the product (or, browser) params present in the given
// request.
func ParseProductOrBrowserParams(v url.Values) (products ProductSpecs, err error) {
	if products, err = ParseProductsParam(v); err != nil {
		return nil, err
	}
	// Handle legacy browser param.
	browserParams, err := ParseBrowsersParam(v)
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

// ParseMaxCountParam parses the 'max-count' parameter as an integer
func ParseMaxCountParam(v url.Values) (*int, error) {
	if maxCountParam := v.Get("max-count"); maxCountParam != "" {
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
func ParseMaxCountParamWithDefault(v url.Values, defaultValue int) (count int, err error) {
	if maxCountParam, err := ParseMaxCountParam(v); maxCountParam != nil {
		return *maxCountParam, err
	} else if err != nil {
		return defaultValue, err
	}
	return defaultValue, nil
}

// ParseDateTimeParam parses the date/time param named "name" as a timestamp.
func ParseDateTimeParam(v url.Values, name string) (*time.Time, error) {
	if fromParam := v.Get(name); fromParam != "" {
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
}

func (d DiffFilterParam) String() string {
	s := ""
	if d.Added {
		s += "A"
	}
	if d.Deleted {
		s += "D"
	}
	if d.Changed {
		s += "C"
	}
	if d.Unchanged {
		s += "U"
	}
	return s
}

// ParseDiffFilterParams collects the diff filtering params for the given request.
// It splits the filter param into the differences to include. The filter param is inspired by Git's --diff-filter flag.
// It also adds the set of test paths to include; see ParsePathsParam below.
func ParseDiffFilterParams(v url.Values) (param DiffFilterParam, paths mapset.Set, err error) {
	param = DiffFilterParam{
		Added:   true,
		Deleted: true,
		Changed: true,
	}
	if filter := v.Get("filter"); filter != "" {
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
				return param, nil, fmt.Errorf("invalid filter character %c", char)
			}
		}
	}
	return param, NewSetFromStringSlice(ParsePathsParam(v)), nil
}

// ParsePathsParam returns a set list of test paths to include, or nil if no
// filter is provided (and all tests should be included). It parses the 'paths'
// parameter, split on commas, and also checks for the (repeatable) 'path' params
func ParsePathsParam(v url.Values) []string {
	return ParseRepeatedParam(v, "path", "paths")
}

// ParseLabelsParam returns a set list of test-run labels to include, or nil if
// no labels are provided.
func ParseLabelsParam(v url.Values) []string {
	return ParseRepeatedParam(v, "label", "labels")
}

// ParseRepeatedParam parses a param that may be a plural name, with all values
// comma-separated, or a repeated singular param.
// e.g. ?label=foo&label=bar vs ?labels=foo,bar
func ParseRepeatedParam(v url.Values, singular string, plural string) (params []string) {
	repeatedParam := v[singular]
	pluralParam := v.Get(plural)
	return parseRepeatedParamValues(repeatedParam, pluralParam)
}

func parseRepeatedParamValues(repeatedParam []string, pluralParam string) (params []string) {
	if len(repeatedParam) == 0 && pluralParam == "" {
		return nil
	}
	allValues := repeatedParam
	if pluralParam != "" {
		allValues = append(allValues, strings.Split(pluralParam, ",")...)
	}

	seen := mapset.NewSet()
	for _, value := range allValues {
		if value == "" {
			continue
		}
		if !seen.Contains(value) {
			params = append(params, value)
			seen.Add(value)
		}
	}
	return params
}

// ParseIntParam parses the result of ParseParam as int64.
func ParseIntParam(v url.Values, param string) (*int, error) {
	strVal := v.Get(param)
	if strVal == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(strVal)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// ParseRepeatedInt64Param parses the result of ParseRepeatedParam as int64.
func ParseRepeatedInt64Param(v url.Values, singular, plural string) (params []int64, err error) {
	strs := ParseRepeatedParam(v, singular, plural)
	if len(strs) < 1 {
		return nil, nil
	}
	ints := make([]int64, len(strs))
	for i, idStr := range strs {
		ints[i], err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	return ints, err
}

// ParseQueryParamInt parses the URL query parameter at key. If the parameter is
// empty or missing, nil is returned.
func ParseQueryParamInt(v url.Values, key string) (*int, error) {
	value := v.Get(key)
	if value == "" {
		return nil, nil
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return &i, fmt.Errorf("Invalid %s value: %s", key, value)
	}
	return &i, err
}

// ParseAlignedParam parses the "aligned" param. See ParseBooleanParam.
func ParseAlignedParam(v url.Values) (aligned *bool, err error) {
	if aligned, err := ParseBooleanParam(v, "aligned"); aligned != nil || err != nil {
		return aligned, err
	}
	// Legacy param name: complete
	return ParseBooleanParam(v, "complete")
}

// ParseBooleanParam parses the given param name as a bool.
// Return nil if the param is missing, true if if it's present with no value,
// otherwise the parsed boolean value of the param's value.
func ParseBooleanParam(v url.Values, name string) (result *bool, err error) {
	q := v
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

// ParseRunIDsParam parses the "run_ids" parameter. If the ID is not a valid
// int64, an error will be returned.
func ParseRunIDsParam(v url.Values) (ids TestRunIDs, err error) {
	return ParseRepeatedInt64Param(v, "run_id", "run_ids")
}

// ParsePRParam parses the "pr" parameter. If it's not a valid int64, an error
// will be returned.
func ParsePRParam(v url.Values) (*int, error) {
	return ParseIntParam(v, "pr")
}

// ParseQueryFilterParams parses shared params for the search and autocomplete
// APIs.
func ParseQueryFilterParams(v url.Values) (filter QueryFilter, err error) {
	keys, err := ParseRunIDsParam(v)
	if err != nil {
		return filter, err
	}
	filter.RunIDs = keys

	filter.Q = v.Get("q")

	return filter, nil
}

// ParseTestRunFilterParams parses all of the filter params for a TestRun query.
func ParseTestRunFilterParams(v url.Values) (filter TestRunFilter, err error) {
	if page, err := ParsePageToken(v); page != nil || err != nil {
		return *page, err
	}

	runSHA, err := ParseSHAParam(v)
	if err != nil {
		return filter, err
	}
	filter.SHA = runSHA
	filter.Labels = NewSetFromStringSlice(ParseLabelsParam(v))
	if filter.Aligned, err = ParseAlignedParam(v); err != nil {
		return filter, err
	}
	if filter.Products, err = ParseProductOrBrowserParams(v); err != nil {
		return filter, err
	}
	if filter.MaxCount, err = ParseMaxCountParam(v); err != nil {
		return filter, err
	}
	if filter.From, err = ParseDateTimeParam(v, "from"); err != nil {
		return filter, err
	}
	if filter.To, err = ParseDateTimeParam(v, "to"); err != nil {
		return filter, err
	}
	return filter, nil
}

// ParseBeforeAndAfterParams parses the before and after params used when
// intending to diff two test runs. Either both or neither of the params
// must be present.
func ParseBeforeAndAfterParams(v url.Values) (ProductSpecs, error) {
	before := v.Get("before")
	after := v.Get("after")
	if before == "" && after == "" {
		return nil, nil
	}
	if before == "" {
		return nil, errors.New("after param provided, but before param missing")
	} else if after == "" {
		return nil, errors.New("before param provided, but after param missing")
	}

	specs := make(ProductSpecs, 2)
	beforeSpec, err := ParseProductSpec(before)
	if err != nil {
		return nil, fmt.Errorf("invalid before param: %s", err.Error())
	}
	specs[0] = beforeSpec

	afterSpec, err := ParseProductSpec(after)
	if err != nil {
		return nil, fmt.Errorf("invalid after param: %s", err.Error())
	}
	specs[1] = afterSpec
	return specs, nil
}

// ParsePageToken decodes a base64 encoding of a TestRunFilter struct.
func ParsePageToken(v url.Values) (*TestRunFilter, error) {
	token := v.Get("page")
	if token == "" {
		return nil, nil
	}
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var filter TestRunFilter
	if err := json.Unmarshal([]byte(decoded), &filter); err != nil {
		return nil, err
	}
	return &filter, nil
}

// ExtractRunIDsBodyParam extracts {"run_ids": <run ids>} from a request JSON
// body. Optionally replace r.Body so that it can be replayed by subsequent
// request handling code can process it.
func ExtractRunIDsBodyParam(r *http.Request, replay bool) (TestRunIDs, error) {
	raw := make([]byte, 0)
	body := r.Body
	raw, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	// If requested, allow subsequent request handling code to re-read body.
	if replay {
		r.Body = ioutil.NopCloser(bytes.NewBuffer(raw))
	}

	var data map[string]*json.RawMessage
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return nil, err
	}

	msg, ok := data["run_ids"]
	if !ok {
		return nil, fmt.Errorf(`JSON request body is missing "run_ids" key; body: %s`, string(raw))
	}
	var runIDs []int64
	err = json.Unmarshal(*msg, &runIDs)
	return TestRunIDs(runIDs), err
}
