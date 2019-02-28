// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination mock_taskcluster/api_mock.go github.com/web-platform-tests/wpt.fyi/api/taskcluster API

package taskcluster

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set"
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

// HandleCheckRunEvent processes a Taskcluster check run "completed" event.
func (a apiImpl) HandleCheckRunEvent(event *github.CheckRunEvent) (bool, error) {
	log := shared.GetLogger(a.ctx)
	status := event.GetCheckRun().GetStatus()
	if status != "completed" {
		log.Infof("Ignoring non-completed status %s", status)
		return false, nil
	}
	detailsURL := event.GetCheckRun().GetDetailsURL()
	sha := event.GetCheckRun().GetHeadSHA()

	labels := mapset.NewSet()
	sender := event.GetSender().GetLogin()
	if sender != "" {
		labels.Add(shared.GetUserLabel(sender))
	}

	taskGroupID, taskID := extractTaskGroupID(detailsURL)
	if taskGroupID == "" {
		return false, fmt.Errorf("unrecognized target_url: %s", detailsURL)
	}

	return processTaskclusterBuild(shared.NewAppEngineAPI(a.ctx), taskGroupID, taskID, sha, shared.ToStringSlice(labels)...)
}
