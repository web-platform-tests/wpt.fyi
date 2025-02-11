// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import "github.com/google/go-github/v69/github"

// RecomputeAction is an action that can be taken to
// trigger a recompute of the diff, against the latest
// master run's results.
func RecomputeAction() *github.CheckRunAction {
	return &github.CheckRunAction{
		Identifier:  "recompute",
		Label:       "Recompute",
		Description: "Recompute with the latest available runs",
	}
}

// IgnoreAction is an action that can be taken to ignore a fail
// outcome, marking it as passing.
func IgnoreAction() *github.CheckRunAction {
	return &github.CheckRunAction{
		Identifier:  "ignore",
		Label:       "Ignore",
		Description: "Mark results as expected (passing)",
	}
}

// CancelAction is an action that can be taken to cancel a pending check run.
func CancelAction() *github.CheckRunAction {
	return &github.CheckRunAction{
		Identifier:  "cancel",
		Label:       "Cancel",
		Description: "Cancel this pending check run",
	}
}
