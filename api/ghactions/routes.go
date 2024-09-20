// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ghactions

import "github.com/web-platform-tests/wpt.fyi/shared"

// RegisterRoutes adds all the api route handlers.
func RegisterRoutes() {
	// notifyHandler exposes an endpoint for notifying wpt.fyi that it can collect
	// the results of a GitHub Actions workflow run.
	// The endpoint is insecure, because we'll only try to fetch (specifically) a
	// web-platform-tests/wpt build with the given ID.
	shared.AddRoute("/api/github-actions/", "github-actions-notify", notifyHandler).Methods("POST")
}
