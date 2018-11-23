// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package checks

import (
	"context"
	"encoding/json"
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
	suitesAPI := NewMockSuitesAPI(mockCtrl)

	processed, err := handleCheckRunEvent(aeAPI, suitesAPI, payload)
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
	suitesAPI := NewMockSuitesAPI(mockCtrl)

	processed, err := handleCheckRunEvent(aeAPI, suitesAPI, payload)
	assert.Nil(t, err)
	assert.False(t, processed)
}

func TestHandleCheckRunEvent_Created_Pending(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	id := int64(wptfyiCheckAppID)
	chrome := "chrome"
	created := "created"
	pending := "pending"
	event := github.CheckRunEvent{
		Action: &created,
		CheckRun: &github.CheckRun{
			App: &github.App{
				ID: &id,
			},
			Name:   &chrome,
			Status: &pending,
		},
	}
	payload, _ := json.Marshal(event)

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
	suitesAPI := NewMockSuitesAPI(mockCtrl)
	suitesAPI.EXPECT().ScheduleResultsProcessing(gomock.Any(), gomock.Any())

	processed, err := handleCheckRunEvent(aeAPI, suitesAPI, payload)
	assert.Nil(t, err)
	assert.True(t, processed)
}
