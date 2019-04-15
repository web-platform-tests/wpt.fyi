// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package screenshot

import "github.com/web-platform-tests/wpt.fyi/shared"

// RegisterRoutes adds all the screenshot route handlers.
func RegisterRoutes() {
	// API endpoint for getting a list of recent screenshot hashes.
	shared.AddRoute("/api/screenshots/hashes", "api-screenshots-hashes",
		shared.WrapApplicationJSON(getHashesHandler))

	// PRIVATE API endpoint for creating a screenshot.
	// Only this AppEngine project can access.
	shared.AddRoute("/api/screenshots/upload", "api-screenshots-upload", uploadScreenshotHandler)
}
