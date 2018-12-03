// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import "github.com/lukebjerring/go-github/github"

// BeforeAndAfter is a struct summarizing pass rates before and after in a diff.
type BeforeAndAfter struct {
	PassingBefore int
	PassingAfter  int
	TotalBefore   int
	TotalAfter    int
}

// Regressed is the struct for regressed.md
type Regressed struct {
	CheckState
	ResultsComparison

	Regressions map[string]BeforeAndAfter
	More        int
}

// GetCheckState returns the info needed to update a check.
func (r Regressed) GetCheckState() CheckState {
	return r.CheckState
}

// GetSummary executes the template for the data.
func (r Regressed) GetSummary() (string, error) {
	return compile(&r, "regressed.md")
}

// RecomputeAction is an action that can be taken to
// trigger a recompute of the diff, against the latest
// master run's results.
func RecomputeAction() *github.CheckRunAction {
	return &github.CheckRunAction{
		Identifier:  "recompute",
		Label:       "Recompute",
		Description: "Recompute against the latest master run",
	}
}

// IgnoreAction is an action that can be taken to ignore a fail
// outcome, marking it as passing.
func IgnoreAction() *github.CheckRunAction {
	return &github.CheckRunAction{
		Identifier:  "ignore",
		Label:       "Ignore",
		Description: "Mark these results as expected (passing)",
	}
}

// GetActions returns the actions that can be taken by the user.
func (r Regressed) GetActions() []*github.CheckRunAction {
	return []*github.CheckRunAction{
		RecomputeAction(),
		IgnoreAction(),
	}
}
