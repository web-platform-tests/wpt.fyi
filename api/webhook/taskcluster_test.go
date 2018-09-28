// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webhook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldProcessStatus_ok(t *testing.T) {
	status := statusEventPayload{
		State:    "success",
		Context:  "Taskcluster",
		Branches: []branchInfo{branchInfo{Name: "master"}},
	}
	assert.True(t, shouldProcessStatus(&status))
}

func TestShouldProcessStatus_unsuccessful(t *testing.T) {
	status := statusEventPayload{
		State:    "error",
		Context:  "Taskcluster",
		Branches: []branchInfo{branchInfo{Name: "master"}},
	}
	assert.False(t, shouldProcessStatus(&status))
}

func TestShouldProcessStatus_notTaskcluster(t *testing.T) {
	status := statusEventPayload{
		State:    "success",
		Context:  "Travis",
		Branches: []branchInfo{branchInfo{Name: "master"}},
	}
	assert.False(t, shouldProcessStatus(&status))
}

func TestShouldProcessStatus_notOnMaster(t *testing.T) {
	status := statusEventPayload{
		State:    "success",
		Context:  "Taskcluster",
		Branches: []branchInfo{branchInfo{Name: "gh-pages"}},
	}
	assert.False(t, shouldProcessStatus(&status))
}

func TestExtractTaskGroupID(t *testing.T) {
	assert.Equal(t, "Y4rnZeqDRXGiRNiqxT5Qeg",
		extractTaskGroupID("https://tools.taskcluster.net/task-group-inspector/#/Y4rnZeqDRXGiRNiqxT5Qeg"))
}

func TestExtractResultURLs(t *testing.T) {
	group := &taskGroupInfo{Tasks: make([]taskInfo, 3)}
	group.Tasks[0].Status.State = "completed"
	group.Tasks[0].Status.TaskID = "foo"
	group.Tasks[0].Task.Metadata.Name = "wpt-firefox-nightly-testharness-1"
	group.Tasks[1].Status.State = "completed"
	group.Tasks[1].Status.TaskID = "bar"
	group.Tasks[1].Task.Metadata.Name = "wpt-firefox-nightly-testharness-2"
	group.Tasks[2].Status.State = "completed"
	group.Tasks[2].Status.TaskID = "baz"
	group.Tasks[2].Task.Metadata.Name = "wpt-chrome-dev-testharness-1"

	urls, err := extractResultURLs(group)
	assert.Nil(t, err)
	assert.Equal(t, map[string][]string{
		"firefox-nightly": {
			"https://queue.taskcluster.net/v1/task/foo/artifacts/public/results/wpt_report.json.gz",
			"https://queue.taskcluster.net/v1/task/bar/artifacts/public/results/wpt_report.json.gz",
		},
		"chrome-dev": {
			"https://queue.taskcluster.net/v1/task/baz/artifacts/public/results/wpt_report.json.gz",
		},
	}, urls)
}

func TestVerifySignature(t *testing.T) {
	message := []byte("test")
	signature := "a053ee211b4693456ef071e336f74ab699250318"
	secret := "95bfab9afa719185ee7c3658356b166b7f45349a"
	assert.True(t, verifySignature(
		message, signature, secret))
	assert.False(t, verifySignature(
		[]byte("foobar"), signature, secret))
	assert.False(t, verifySignature(
		message, "875a5feef4cde4265d6d5d21c304d755903ccb60", secret))
	assert.False(t, verifySignature(
		message, signature, "875a5feef4cde4265d6d5d21c304d755903ccb60"))
	// Test an ill-formed (odd-length) signature.
	assert.False(t, verifySignature(
		message, "875a5feef4cde4265d6d5d21c304d755903ccb6", secret))
}
