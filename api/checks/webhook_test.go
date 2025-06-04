//go:build small
// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package checks

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-github/v72/github"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/checks/mock_checks"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"go.uber.org/mock/gomock"
)

func TestHandleCheckRunEvent_InvalidApp(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	id := int64(123)
	chrome := "chrome"
	event := github.CheckRunEvent{
		CheckRun: &github.CheckRun{
			App: &github.App{
				ID: &id,
			},
			Name: &chrome,
		},
	}
	payload, _ := json.Marshal(event)

	api := mock_checks.NewMockAPI(mockCtrl)
	api.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())

	processed, err := handleCheckRunEvent(api, payload)
	assert.Nil(t, err)
	assert.False(t, processed)
}

func TestHandleCheckRunEvent_Created_Completed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("1234567890", 4)
	event := getCheckRunCreatedEvent("completed", "lukebjerring", sha)
	payload, _ := json.Marshal(event)

	api := mock_checks.NewMockAPI(mockCtrl)
	api.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	api.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)

	processed, err := handleCheckRunEvent(api, payload)
	assert.Nil(t, err)
	assert.False(t, processed)
}

func TestHandleCheckRunEvent_Created_Pending_ChecksNotEnabledForUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("0123456789", 4)
	event := getCheckRunCreatedEvent("pending", "user-without-checks", sha)
	payload, _ := json.Marshal(event)

	api := mock_checks.NewMockAPI(mockCtrl)
	api.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	api.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)

	processed, err := handleCheckRunEvent(api, payload)
	assert.Nil(t, err)
	assert.False(t, processed)
}

func TestHandleCheckRunEvent_Created_Pending(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("0123456789", 4)
	event := getCheckRunCreatedEvent("pending", "lukebjerring", sha)
	payload, _ := json.Marshal(event)

	api := mock_checks.NewMockAPI(mockCtrl)
	api.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	api.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
	api.EXPECT().ScheduleResultsProcessing(sha, sharedtest.SameProductSpec("chrome")).Return(nil)

	processed, err := handleCheckRunEvent(api, payload)
	assert.Nil(t, err)
	assert.True(t, processed)
}

func TestHandleCheckRunEvent_ActionRequested_Ignore(t *testing.T) {
	for _, prefix := range []string{"staging.wpt.fyi - ", "wpt.fyi - ", ""} {
		t.Run(prefix, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			id := int64(wptfyiStagingCheckAppID)
			sha := strings.Repeat("0123456789", 4)
			name := prefix + "chrome"
			requestedAction := "requested_action"
			pending := "pending"
			username := "lukebjerring"
			owner := shared.WPTRepoOwner
			repo := shared.WPTRepoName
			appID := int64(wptfyiStagingCheckAppID)
			event := github.CheckRunEvent{
				Action: &requestedAction,
				CheckRun: &github.CheckRun{
					App:     &github.App{ID: &id},
					Name:    &name,
					Status:  &pending,
					HeadSHA: &sha,
				},
				Repo: &github.Repository{
					Owner: &github.User{Login: &owner},
					Name:  &repo,
				},
				RequestedAction: &github.RequestedAction{Identifier: "ignore"},
				Installation:    &github.Installation{AppID: &appID},
				Sender:          &github.User{Login: &username},
			}
			payload, _ := json.Marshal(event)

			api := mock_checks.NewMockAPI(mockCtrl)
			api.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
			api.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
			api.EXPECT().IgnoreFailure(username, owner, repo, event.GetCheckRun(), event.GetInstallation())

			processed, err := handleCheckRunEvent(api, payload)
			assert.Nil(t, err)
			assert.True(t, processed)
		})
	}
}

func TestHandleCheckRunEvent_ActionRequested_Recompute(t *testing.T) {
	for _, prefix := range []string{"staging.wpt.fyi - ", "wpt.fyi - ", ""} {
		t.Run(prefix, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			id := int64(wptfyiStagingCheckAppID)
			sha := strings.Repeat("0123456789", 4)
			name := prefix + "chrome[experimental]"
			requestedAction := "requested_action"
			pending := "pending"
			username := "lukebjerring"
			owner := shared.WPTRepoOwner
			repo := shared.WPTRepoName
			appID := int64(wptfyiStagingCheckAppID)
			event := github.CheckRunEvent{
				Action: &requestedAction,
				CheckRun: &github.CheckRun{
					App:     &github.App{ID: &id},
					Name:    &name,
					Status:  &pending,
					HeadSHA: &sha,
				},
				Repo: &github.Repository{
					Owner: &github.User{Login: &owner},
					Name:  &repo,
				},
				RequestedAction: &github.RequestedAction{Identifier: "recompute"},
				Installation:    &github.Installation{AppID: &appID},
				Sender:          &github.User{Login: &username},
			}
			payload, _ := json.Marshal(event)

			api := mock_checks.NewMockAPI(mockCtrl)
			api.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
			api.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
			api.EXPECT().ScheduleResultsProcessing(sha, sharedtest.SameProductSpec("chrome[experimental]")).Return(nil)

			processed, err := handleCheckRunEvent(api, payload)
			assert.Nil(t, err)
			assert.True(t, processed)
		})
	}
}

func TestHandleCheckRunEvent_ActionRequested_Cancel(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("0123456789", 4)
	username := "lukebjerring"
	event := getCheckRunCreatedEvent("completed", username, sha)
	requestedAction := "requested_action"
	event.Action = &requestedAction
	event.RequestedAction = &github.RequestedAction{Identifier: "cancel"}
	payload, _ := json.Marshal(event)

	api := mock_checks.NewMockAPI(mockCtrl)
	api.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	api.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
	api.EXPECT().CancelRun(username, shared.WPTRepoOwner, shared.WPTRepoName, event.GetCheckRun(), event.GetInstallation())

	processed, err := handleCheckRunEvent(api, payload)
	assert.Nil(t, err)
	assert.True(t, processed)
}

func getCheckRunCreatedEvent(status, sender, sha string) github.CheckRunEvent {
	id := int64(wptfyiStagingCheckAppID)
	chrome := "chrome"
	created := "created"
	repoName := shared.WPTRepoName
	repoOwner := shared.WPTRepoOwner
	return github.CheckRunEvent{
		Action: &created,
		CheckRun: &github.CheckRun{
			App:     &github.App{ID: &id},
			Name:    &chrome,
			Status:  &status,
			HeadSHA: &sha,
		},
		Repo: &github.Repository{
			Name:  &repoName,
			Owner: &github.User{Login: &repoOwner},
		},
		Sender: &github.User{Login: &sender},
	}
}

func TestHandlePullRequestEvent_ChecksNotEnabledForUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("1234567890", 4)
	event := getOpenedPREvent("user-without-checks", sha)
	payload, _ := json.Marshal(event)

	api := mock_checks.NewMockAPI(mockCtrl)
	api.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	api.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)

	processed, err := handlePullRequestEvent(api, payload)
	assert.Nil(t, err)
	assert.False(t, processed)
}

func TestHandlePullRequestEvent_ChecksEnabledForUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("1234567890", 4)
	event := getOpenedPREvent("lukebjerring", sha)
	payload, _ := json.Marshal(event)

	api := mock_checks.NewMockAPI(mockCtrl)
	api.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	api.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
	api.EXPECT().GetWPTRepoAppInstallationIDs().Return(wptfyiStagingCheckAppID, wptRepoStagingInstallationID)
	api.EXPECT().CreateWPTCheckSuite(wptfyiStagingCheckAppID, wptRepoStagingInstallationID, sha, 123).Return(true, nil)

	processed, err := handlePullRequestEvent(api, payload)
	assert.Nil(t, err)
	assert.True(t, processed)
}

func getOpenedPREvent(user, sha string) github.PullRequestEvent {
	opened := "opened"
	// handlePullRequestEvent only operates on pull requests from forks, so
	// the head repo must be different from the base.
	headRepoID := wptRepoID - 1
	baseRepoID := wptRepoID
	number := 123
	return github.PullRequestEvent{
		Number: &number,
		PullRequest: &github.PullRequest{
			User: &github.User{Login: &user},
			Head: &github.PullRequestBranch{
				SHA:  &sha,
				Repo: &github.Repository{ID: &headRepoID},
			},
			Base: &github.PullRequestBranch{
				Repo: &github.Repository{ID: &baseRepoID},
			},
		},
		Action: &opened,
	}
}
