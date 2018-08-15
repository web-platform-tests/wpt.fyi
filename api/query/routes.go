// Copyright 218 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import "github.com/web-platform-tests/wpt.fyi/shared"

func RegisterRoutes() {
	// API endpoint for searching results over given runs.
	shared.AddRoute("/api/search", "api-search", apiSearchHandler)
	// API endpoint for search autocomplete.
	shared.AddRoute("/api/autocomplete", "api-autocomplete", apiAutocompleteHandler)
}
