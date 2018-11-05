// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import "github.com/web-platform-tests/wpt.fyi/shared"

// RegisterRoutes adds route handlers for webhooks.
func RegisterRoutes() {
	// GitHub webhook for creating custom status checks.
	shared.AddRoute("/api/webhook/status", "api-webhook-check", checkWebhookHandler)
}
