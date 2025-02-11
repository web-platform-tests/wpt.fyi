//go:build small
// +build small

// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ghactions

import (
	"testing"

	"github.com/google/go-github/v69/github"
	"github.com/stretchr/testify/assert"
)

func PointerTo[T any](v T) *T {
	return &v
}

func TestArtifactRegexes(t *testing.T) {
	assert.True(t, prHeadRegex.MatchString("safari-preview-1-affected-tests"))
	assert.True(t, prBaseRegex.MatchString("safari-preview-1-affected-tests-without-changes"))

	// Base and Head could be confused with substring matching
	assert.False(t, prBaseRegex.MatchString("safari-preview-1-affected-tests"))
	assert.False(t, prHeadRegex.MatchString("safari-preview-1-affected-tests-without-changes"))
}

func TestEpochBranchesRegex(t *testing.T) {
	assert.True(t, epochBranchesRegex.MatchString("epochs/twelve_hourly"))
	assert.True(t, epochBranchesRegex.MatchString("epochs/six_hourly"))
	assert.True(t, epochBranchesRegex.MatchString("epochs/weekly"))
	assert.True(t, epochBranchesRegex.MatchString("epochs/daily"))

	assert.False(t, epochBranchesRegex.MatchString("weekly"))
	assert.False(t, epochBranchesRegex.MatchString("a/epochs/weekly"))
}

func TestChooseLabels(t *testing.T) {
	wptOrgUser := github.User{
		Login: PointerTo("web-platform-tests"),
	}

	otherUser := github.User{
		Login: PointerTo("xxx"),
	}

	wptRepo := github.Repository{
		Name:     PointerTo("wpt"),
		FullName: PointerTo("web-platform-tests/wpt"),
		Owner:    &wptOrgUser,
	}

	otherRepo := github.Repository{
		Name:     PointerTo("wpt"),
		FullName: PointerTo("xxx/wpt"),
		Owner:    &otherUser,
	}

	masterWorkflowRun := github.WorkflowRun{
		HeadBranch:     PointerTo("master"),
		Event:          PointerTo("push"),
		Status:         PointerTo("completed"),
		Conclusion:     PointerTo("success"),
		HeadSHA:        PointerTo("74dc6f6f5b2ba16940e6b6075f0faf311361dbb2"),
		Repository:     &wptRepo,
		HeadRepository: &wptRepo,
	}

	masterOtherWorkflowRun := github.WorkflowRun{
		HeadBranch:     PointerTo("master"),
		Event:          PointerTo("push"),
		Status:         PointerTo("completed"),
		Conclusion:     PointerTo("success"),
		HeadSHA:        PointerTo("74dc6f6f5b2ba16940e6b6075f0faf311361dbb2"),
		Repository:     &otherRepo,
		HeadRepository: &otherRepo,
	}

	prWorkflowRun := github.WorkflowRun{
		HeadBranch:     PointerTo("new-branch"),
		Event:          PointerTo("pull_request"),
		Status:         PointerTo("completed"),
		Conclusion:     PointerTo("success"),
		HeadSHA:        PointerTo("74dc6f6f5b2ba16940e6b6075f0faf311361dbb2"),
		Repository:     &wptRepo,
		HeadRepository: &wptRepo,
	}

	prOtherWorkflowRun := github.WorkflowRun{
		HeadBranch:     PointerTo("master"),
		Event:          PointerTo("pull_request"),
		Status:         PointerTo("completed"),
		Conclusion:     PointerTo("success"),
		HeadSHA:        PointerTo("74dc6f6f5b2ba16940e6b6075f0faf311361dbb2"),
		Repository:     &wptRepo,
		HeadRepository: &otherRepo,
	}

	assert.ElementsMatch(
		t,
		chooseLabels(&masterWorkflowRun, "results-safari-1", "web-platform-tests", "wpt").ToSlice(),
		[]string{"master"},
	)

	assert.ElementsMatch(t,
		chooseLabels(&masterOtherWorkflowRun, "results-safari-1", "web-platform-tests", "wpt").ToSlice(),
		[]string{},
	)

	assert.ElementsMatch(t,
		chooseLabels(&prWorkflowRun, "results-safari-1", "web-platform-tests", "wpt").ToSlice(),
		[]string{},
	)

	assert.ElementsMatch(t,
		chooseLabels(&prOtherWorkflowRun, "results-safari-1", "web-platform-tests", "wpt").ToSlice(),
		[]string{},
	)

	assert.ElementsMatch(
		t,
		chooseLabels(&prWorkflowRun, "results-safari-1-affected-tests", "web-platform-tests", "wpt").ToSlice(),
		[]string{"pr_head"},
	)

	assert.ElementsMatch(t,
		chooseLabels(&prOtherWorkflowRun, "results-safari-1-affected-tests-without-changes", "web-platform-tests", "wpt").ToSlice(),
		[]string{"pr_base"},
	)
}
