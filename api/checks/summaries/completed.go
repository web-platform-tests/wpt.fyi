// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import (
	"github.com/google/go-github/v70/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// ResultsComparison is all the fields shared across summaries that
// involve a result comparison.
type ResultsComparison struct {
	BaseRun       shared.TestRun
	HeadRun       shared.TestRun
	MasterDiffURL string
	DiffURL       string // URL for the diff-view of the results
	HostURL       string // Host environment URL, e.g. "https://wpt.fyi"
}

// Completed is the struct for completed.md.
type Completed struct {
	CheckState
	ResultsComparison

	Results BeforeAndAfter
	More    int
}

// GetCheckState returns the info needed to update a check.
func (c Completed) GetCheckState() CheckState {
	return c.CheckState
}

// GetSummary executes the template for the data.
func (c Completed) GetSummary() (string, error) {
	return compile(&c, "completed.md")
}

// GetActions returns the actions that can be taken by the user.
func (c Completed) GetActions() []*github.CheckRunAction {
	return []*github.CheckRunAction{
		RecomputeAction(),
	}
}
