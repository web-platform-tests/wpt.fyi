// Copyright 218 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import "github.com/web-platform-tests/wpt.fyi/shared"

// RegisterRoutes binds query API handlers to URL routes.
func RegisterRoutes() {
	// API endpoint for searching results over given runs.
	shared.AddRoute(
		"/api/search",
		"api-search",
		shared.WrapPermissiveCORS(apiSearchHandler))
	// API endpoint for search autocomplete.
	shared.AddRoute("/api/autocomplete", "api-autocomplete", apiAutocompleteHandler)
}
