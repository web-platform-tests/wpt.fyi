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
	"github.com/lukebjerring/go-github/github"
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

func TestHandleCheckRunEvent_Created_Complete(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	id := int64(wptfyiCheckAppID)
	chrome := "chrome"
	created := "created"
	completed := "completed"
	event := github.CheckRunEvent{
		Action: &created,
		CheckRun: &github.CheckRun{
			App: &github.App{
				ID: &id,
			},
			Name:   &chrome,
			Status: &completed,
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

func TestHandleCheckRunEvent_Created_Pending(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	id := int64(wptfyiCheckAppID)
	sha := strings.Repeat("0123456789", 4)
	chrome := "chrome"
	created := "created"
	pending := "pending"
	event := github.CheckRunEvent{
		Action: &created,
		CheckRun: &github.CheckRun{
			App: &github.App{
				ID: &id,
			},
			Name:    &chrome,
			Status:  &pending,
			HeadSHA: &sha,
		},
	}
	payload, _ := json.Marshal(event)

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
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
	username := "username"
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
	checksAPI := NewMockAPI(mockCtrl)
	checksAPI.EXPECT().IgnoreFailure(username, owner, repo, event.GetCheckRun(), event.GetInstallation())

	processed, err := handleCheckRunEvent(aeAPI, checksAPI, payload)
	assert.Nil(t, err)
	assert.True(t, processed)
}
