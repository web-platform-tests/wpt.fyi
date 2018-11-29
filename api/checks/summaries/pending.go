// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

// Pending is the struct for pending.md
type Pending struct {
	CheckState

	HostName string // Host environment name
	RunsURL  string // URL for the list of test runs
}

// GetCheckState returns the info needed to update a check.
func (c Pending) GetCheckState() CheckState {
	return c.CheckState
}

// GetSummary executes the template for the data.
func (c Pending) GetSummary() (string, error) {
	return compile(&c, "pending.md")
}
