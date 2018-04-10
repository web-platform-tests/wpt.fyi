// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"io/ioutil"
	"sort"

	models "github.com/w3c/wptdashboard/shared"
)

// All errors are considered fatal
var browserNames, browserNamesAlphabetical = loadBrowsers()

// GetBrowserNames returns an alphabetically-ordered array of the names
// of the browsers which are to be included (flagged as initially_loaded in the
// JSON).
func GetBrowserNames() ([]string, error) {
	return browserNamesAlphabetical, nil
}

// IsBrowserName determines whether the given name string is a valid browser name.
// Used for validating user-input params for browsers.
func IsBrowserName(name string) bool {
	_, ok := browserNames[name]
	return ok
}

// loadBrowsers loads, parses and returns the set of names of browsers which
// are to be included (flagged as initially_loaded in the JSON).
func loadBrowsers() (map[string]bool, []string) {
	var browsers map[string]models.Browser
	var err error
	var bytes []byte
	var browserNames map[string]bool
	var browserNamesAlphabetical []string

	if bytes, err = ioutil.ReadFile("browsers.json"); err != nil {
		panic(err)
	}

	if err = json.Unmarshal(bytes, &browsers); err != nil {
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}
