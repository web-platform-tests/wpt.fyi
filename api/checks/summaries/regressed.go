// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import (
	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// BeforeAndAfter summarizes counts for pass/total before and after, across a
// particular path (could be a folder, could be a test).
type BeforeAndAfter map[string]TestBeforeAndAfter

// Add the given before/after counts to the totals.
func (bna BeforeAndAfter) Add(p string, before, after shared.TestSummary) {
	sum := TestBeforeAndAfter{}
	if existing, ok := bna[p]; ok {
		sum = existing
	}
	if before != nil {
		sum.PassingBefore += before[0]
		sum.TotalBefore += before[1]
	}
	if after != nil {
		sum.PassingAfter += after[0]
		sum.TotalAfter += after[1]
	}
	bna[p] = sum
}

// TestBeforeAndAfter is a struct summarizing pass rates before and after in a diff.
type TestBeforeAndAfter struct {
	PassingBefore int
	PassingAfter  int
	TotalBefore   int
	TotalAfter    int
}

// Regressed is the struct for regressed.md
type Regressed struct {
	CheckState
	ResultsComparison

	Regressions BeforeAndAfter
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

// GetActions returns the actions that can be taken by the user.
func (r Regressed) GetActions() []*github.CheckRunAction {
	return []*github.CheckRunAction{
		RecomputeAction(),
		IgnoreAction(),
	}
}
