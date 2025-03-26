// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	mapset "github.com/deckarep/golang-set"
)

// ExperimentalLabel is the implicit label present for runs marked 'experimental'.
const ExperimentalLabel = "experimental"

// LatestSHA is a helper for the 'latest' keyword/special case.
const LatestSHA = "latest"

// StableLabel is the implicit label present for runs marked 'stable'.
const StableLabel = "stable"

// BetaLabel is the implicit label present for runs marked 'beta'.
const BetaLabel = "beta"

// MasterLabel is the implicit label present for runs marked 'master',
// i.e. run from the master branch.
const MasterLabel = "master"

// PRBaseLabel is the implicit label for running just the affected tests on a
// PR but without the changes (i.e. against the base branch).
const PRBaseLabel = "pr_base"

// PRHeadLabel is the implicit label for running just the affected tests on the
// head of a PR (with the changes).
const PRHeadLabel = "pr_head"

// UserLabelPrefix is a prefix used to denote a label for a user's GitHub handle,
// prefixed because usernames are essentially user input.
const UserLabelPrefix = "user:"

// WPTRepoOwner is the owner (username) for the GitHub wpt repo.
const WPTRepoOwner = "web-platform-tests"

// WPTRepoName is the repo name for the GitHub wpt repo.
const WPTRepoName = "wpt"

// GetUserLabel prefixes the given username with the prefix for using as a label.
func GetUserLabel(username string) string {
	return UserLabelPrefix + username
}

// ProductChannelToLabel maps known product-specific channel names
// to the wpt.fyi model's equivalent.
func ProductChannelToLabel(channel string) string {
	switch channel {
	case "release", StableLabel:
		return StableLabel
	case BetaLabel:
		return BetaLabel
	case "dev", "nightly", "preview", ExperimentalLabel:
		return ExperimentalLabel
	}
	return ""
}

// GetDefaultProducts returns the default set of products to show on wpt.fyi
func GetDefaultProducts() ProductSpecs {
	browserNames := GetDefaultBrowserNames()
	products := make(ProductSpecs, len(browserNames))
	for i, name := range browserNames {
		products[i] = ProductSpec{}
		products[i].BrowserName = name
	}
	return products
}

// ToStringSlice converts a set to a typed string slice.
func ToStringSlice(set mapset.Set) []string {
	if set == nil {
		return nil
	}
	slice := set.ToSlice()
	result := make([]string, len(slice))
	for i, item := range slice {
		result[i] = item.(string)
	}
	return result
}

// IsLatest returns whether a SHA[0:10] is empty or "latest", both
// of which are treated as looking up the latest run for each browser.
func IsLatest(sha string) bool {
	return sha == "" || sha == "latest"
}

// NewSetFromStringSlice is a helper for the inability to cast []string to []interface{}
func NewSetFromStringSlice(items []string) mapset.Set {
	set := mapset.NewSet()
	if items == nil {
		return set
	}
	for _, i := range items {
		set.Add(i)
	}
	return set
}

// StringSliceContains returns true if the given slice contains the given string.
func StringSliceContains(ss []string, s string) bool {
	for _, i := range ss {
		if i == s {
			return true
		}
	}
	return false
}

// MapStringKeys returns the keys in the given string-keyed map.
func MapStringKeys(m interface{}) ([]string, error) {
	mapType := reflect.ValueOf(m)
	if mapType.Kind() != reflect.Map {
		return nil, errors.New("interface is not a map type")
	}
	keys := mapType.MapKeys()
	strKeys := make([]string, len(keys))
	for i, key := range keys {
		var ok bool
		if strKeys[i], ok = key.Interface().(string); !ok {
			return nil, fmt.Errorf("key %v was not a string type", key)
		}
	}
	return strKeys, nil
}

// GetResultsURL constructs the URL to the result of a single test file in the
// given run.
func GetResultsURL(run TestRun, testFile string) (resultsURL string) {
	resultsURL = run.ResultsURL
	if testFile != "" && testFile != "/" {
		// Assumes that result files are under a directory named SHA[0:10].
		resultsBase := strings.SplitAfter(resultsURL, "/"+run.Revision)[0]
		resultsPieces := strings.Split(resultsURL, "/")
		re := regexp.MustCompile("(-summary_v2)?\\.json\\.gz$")
		product := re.ReplaceAllString(resultsPieces[len(resultsPieces)-1], "")
		resultsURL = fmt.Sprintf("%s/%s/%s", resultsBase, product, testFile)
	}
	return resultsURL
}

// CropString conditionally crops a string to the given length, if it is longer.
// Returns the original string otherwise.
func CropString(s string, i int) string {
	if len(s) <= i {
		return s
	}
	return s[:i]
}

// GetSharedPath gets the longest path shared between the given paths.
func GetSharedPath(paths ...string) string {
	var parts []string
	for _, path := range paths {
		if parts == nil {
			parts = strings.Split(path, "/")
		} else {
			otherParts := strings.Split(path, "/")
			for i, part := range parts {
				if part == otherParts[i] {
					continue
				}
				// Crop to the matching parts, append empty last-part
				// so that we have a trailing slash.
				parts = append(parts[:i], "")
				break
			}
		}
	}
	return strings.Join(parts, "/")
}
