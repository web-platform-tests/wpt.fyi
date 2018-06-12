// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/deckarep/golang-set"
)

// All errors are considered fatal
var browserNames, browserNamesAlphabetical = loadBrowsers()

// ExperimentalLabel is the implicit label present for runs marked 'experimental'.
const ExperimentalLabel = `experimental`

// GetBrowserNames returns an alphabetically-ordered array of the names
// of the browsers which are to be included (flagged as initially_loaded in the
// JSON).
func GetBrowserNames() ([]string, error) {
	// Slice to make source immutable
	tmp := make([]string, len(browserNamesAlphabetical))
	copy(tmp, browserNamesAlphabetical)
	return tmp, nil
}

// IsBrowserName determines whether the given name string is a valid browser name.
// Used for validating user-input params for browsers.
func IsBrowserName(name string) bool {
	if strings.HasSuffix(name, "-"+ExperimentalLabel) {
		name = name[0 : len(name)-1-len(ExperimentalLabel)] // Trim suffix.
	}
	_, ok := browserNames[name]
	return ok
}

// loadBrowsers loads, parses and returns the set of names of browsers which
// are to be included (flagged as initially_loaded in the JSON).
func loadBrowsers() (map[string]bool, []string) {
	var browsers map[string]Browser
	var err error
	var browserNames map[string]bool
	var browserNamesAlphabetical []string

	if err = json.Unmarshal([]byte(browsersJSON), &browsers); err != nil {
		panic(err)
	}

	browserNames = make(map[string]bool)
	for _, browser := range browsers {
		if browser.InitiallyLoaded {
			browserNamesAlphabetical = append(browserNamesAlphabetical, browser.BrowserName)
			browserNames[browser.BrowserName] = true
		}
	}
	sort.Strings(browserNamesAlphabetical)

	return browserNames, browserNamesAlphabetical
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
