// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import "strings"

// AutocompleteResult contains a single autocomplete suggestion.
type AutocompleteResult struct {
	// String represents the most basic form of an autocomplete result. It is the
	// complete string being recommended. The client is responsibile for fitting
	// it to any substring(s) already appearing in a search UI.
	String string `json:"string"`
}

// AutocompleteResponse contains a response to autocmplete API calls.
type AutocompleteResponse struct {
	// Results is the collection of test results, grouped by test file name.
	Results []SearchResult `json:"results"`
}

type byQueryIndex struct {
	q  string
	rs []AutocompleteResult
}

func (r byQueryIndex) Len() int      { return len(r.rs) }
func (r byQueryIndex) Swap(i, j int) { r.rs[i], r.rs[j] = r.rs[j], r.rs[i] }
func (r byQueryIndex) Less(i, j int) bool {
	return strings.Index(r.rs[i].String, r.q) < strings.Index(r.rs[j].String, r.q)
}
