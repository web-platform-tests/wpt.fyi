// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	"google.golang.org/appengine/datastore"
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

// AddRoute registers a handler for an http path (route).
// Note that it adds an HSTS header to the response.
func AddRoute(route string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(route, wrapHSTS(handler))
}

func wrapHSTS(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := "max-age=31536000; preload"
		w.Header().Add("Strict-Transport-Security", value)
		h(w, r)
	})
}

// QueryPrefix returns the given query with a prefix filter on the given
// field name, using the >= and < filters.
func QueryPrefix(query *datastore.Query, fieldName, prefix string, desc bool) *datastore.Query {
	order := fieldName
	if desc {
		order = "-" + order
	}
	return query.
		Order(order).
		Filter(fieldName+" >=", prefix).
		Filter(fieldName+" <=", fmt.Sprintf("%s%c", prefix, utf8.MaxRune))
}
