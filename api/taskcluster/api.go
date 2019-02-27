// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination mock_taskcluster/api_mock.go github.com/web-platform-tests/wpt.fyi/api/taskcluster API

package taskcluster

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// AppID is the ID of the Taskcluster GitHub app.
const AppID = int64(1317)

// https://docs.microsoft.com/en-us/rest/api/taskcluster/devops/build/artifacts/get?view=taskcluster-devops-rest-4.1

// API is for Azure Taskcluster related requests.
type API interface {
	HandleCheckRunEvent(*github.CheckRunEvent) (bool, error)
}

type apiImpl struct {
	ctx context.Context
}

// NewAPI returns an implementation of taskcluster API
func NewAPI(ctx context.Context) API {
	return apiImpl{
		ctx: ctx,
	}
}

// HandleCheckRunEvent processes an Azure Taskcluster check run "completed" event.
func (a apiImpl) HandleCheckRunEvent(checkRun *github.CheckRunEvent) (bool, error) {
	return handleCheckRunEvent(shared.NewAppEngineAPI(a.ctx), checkRun)
}
