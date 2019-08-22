// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package notifications

// Subscription is a regression notification subscription.
type Subscription struct {
	Email string
	Paths []string
}
