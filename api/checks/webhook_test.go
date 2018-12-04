// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package checks

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
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

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
	checksAPI := NewMockAPI(mockCtrl)

	processed, err := handleCheckRunEvent(aeAPI, checksAPI, payload)
	assert.Nil(t, err)
	assert.False(t, processed)
}

func TestHandleCheckRunEvent_Created_Completed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("1234567890", 4)
	event := getCheckRunCreatedEvent("completed", "lukebjerring", sha)
	payload, _ := json.Marshal(event)

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
	aeAPI.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
	checksAPI := NewMockAPI(mockCtrl)

	processed, err := handleCheckRunEvent(aeAPI, checksAPI, payload)
	assert.Nil(t, err)
	assert.False(t, processed)
}

func TestHandleCheckRunEvent_Created_Pending_UserNotWhitelisted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("0123456789", 4)
	event := getCheckRunCreatedEvent("pending", "user-not-whitelisted", sha)
	payload, _ := json.Marshal(event)

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
	aeAPI.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
	checksAPI := NewMockAPI(mockCtrl)

	processed, err := handleCheckRunEvent(aeAPI, checksAPI, payload)
	assert.Nil(t, err)
	assert.False(t, processed)
}

func TestHandleCheckRunEvent_Created_Pending(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("0123456789", 4)
	event := getCheckRunCreatedEvent("pending", "lukebjerring", sha)
	payload, _ := json.Marshal(event)

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
	aeAPI.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
	checksAPI := NewMockAPI(mockCtrl)
	checksAPI.EXPECT().ScheduleResultsProcessing(sha, sharedtest.SameProductSpec("chrome"))

	processed, err := handleCheckRunEvent(aeAPI, checksAPI, payload)
	assert.Nil(t, err)
	assert.True(t, processed)
}

func TestHandleCheckRunEvent_ActionRequested_Ignore(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	id := int64(wptfyiCheckAppID)
	sha := strings.Repeat("0123456789", 4)
	chrome := "chrome"
	requestedAction := "requested_action"
	pending := "pending"
	username := "lukebjerring"
	owner := wptRepoOwner
	repo := wptRepoName
	appID := int64(wptfyiCheckAppID)
	event := github.CheckRunEvent{
		Action: &requestedAction,
		CheckRun: &github.CheckRun{
			App:     &github.App{ID: &id},
			Name:    &chrome,
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

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
	aeAPI.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
	checksAPI := NewMockAPI(mockCtrl)
	checksAPI.EXPECT().IgnoreFailure(username, owner, repo, event.GetCheckRun(), event.GetInstallation())

	processed, err := handleCheckRunEvent(aeAPI, checksAPI, payload)
	assert.Nil(t, err)
	assert.True(t, processed)
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

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
	aeAPI.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
	checksAPI := NewMockAPI(mockCtrl)
	checksAPI.EXPECT().CancelRun(username, wptRepoOwner, wptRepoName, event.GetCheckRun(), event.GetInstallation())

	processed, err := handleCheckRunEvent(aeAPI, checksAPI, payload)
	assert.Nil(t, err)
	assert.True(t, processed)
}

func getCheckRunCreatedEvent(status, sender, sha string) github.CheckRunEvent {
	id := int64(wptfyiCheckAppID)
	chrome := "chrome"
	created := "created"
	repoName := wptRepoName
	repoOwner := wptRepoOwner
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

func TestHandlePullRequestEvent_UserNotWhitelisted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("1234567890", 4)
	event := getOpenedPREvent("user-not-whitelisted", sha)
	payload, _ := json.Marshal(event)

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
	aeAPI.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
	checksAPI := NewMockAPI(mockCtrl)

	processed, err := handlePullRequestEvent(aeAPI, checksAPI, payload)
	assert.Nil(t, err)
	assert.False(t, processed)
}

func TestHandlePullRequestEvent_UserWhitelisted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("1234567890", 4)
	event := getOpenedPREvent("lukebjerring", sha)
	payload, _ := json.Marshal(event)

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	aeAPI.EXPECT().IsFeatureEnabled(checksForAllUsersFeature).Return(false)
	checksAPI := NewMockAPI(mockCtrl)
	checksAPI.EXPECT().CreateWPTCheckSuite(wptfyiCheckAppID, wptRepoInstallationID, sha).Return(true, nil)

	processed, err := handlePullRequestEvent(aeAPI, checksAPI, payload)
	assert.Nil(t, err)
	assert.True(t, processed)
}

func getOpenedPREvent(user, sha string) github.PullRequestEvent {
	opened := "opened"
	repoID := wptRepoID
	return github.PullRequestEvent{
		PullRequest: &github.PullRequest{
			User: &github.User{Login: &user},
			Head: &github.PullRequestBranch{SHA: &sha},
			Base: &github.PullRequestBranch{
				Repo: &github.Repository{ID: &repoID},
			},
		},
		Action: &opened,
	}
}
