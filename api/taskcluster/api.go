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
	HandleCheckSuiteEvent(*github.CheckSuiteEvent) (bool, error)
}

type apiImpl struct {
	ctx   context.Context
	aeAPI shared.AppEngineAPI
}

// NewAPI returns an implementation of taskcluster API
func NewAPI(ctx context.Context) API {
	return apiImpl{
		ctx:   ctx,
		aeAPI: shared.NewAppEngineAPI(ctx),
	}
}

// HandleCheckSuiteEvent processes a Taskcluster check suite "completed" event.
func (a apiImpl) HandleCheckSuiteEvent(event *github.CheckSuiteEvent) (bool, error) {
	log := shared.GetLogger(a.ctx)
	status := event.GetCheckSuite().GetStatus()
	if status != "completed" {
		log.Infof("Ignoring non-completed status %s", status)
		return false, nil
	}
	sha := event.GetCheckSuite().GetHeadSHA()

	labels := mapset.NewSet()
	sender := event.GetSender().GetLogin()
	if sender != "" {
		labels.Add(shared.GetUserLabel(sender))
	}
	if event.GetCheckSuite().GetHeadBranch() == shared.MasterLabel {
		labels.Add(shared.MasterLabel)
	}

	ghClient, err := a.aeAPI.GetGitHubClient()
	if err != nil {
		log.Errorf("Failed to get GitHub client: %s", err.Error())
		return false, err
	}
	runs, _, err := ghClient.Checks.ListCheckRunsCheckSuite(a.ctx, shared.WPTRepoOwner, shared.WPTRepoName, event.GetCheckSuite().GetID(), nil)
	if err != nil {
		log.Errorf("Failed to fetch check runs for suite %v: %s", event.GetCheckSuite().GetID(), err.Error())
		return false, err
	}
	byGroup := mapset.NewSet()
	for _, run := range runs.CheckRuns {
		taskGroupID, _ := extractTaskGroupID(run.GetDetailsURL())
		if taskGroupID == "" {
			return false, fmt.Errorf("unrecognized target_url: %s", run.GetDetailsURL())
		}
		byGroup.Add(taskGroupID)
	}

	if byGroup.Cardinality() > 1 {
		log.Errorf("Encountered multiple (%v) TaskGroup IDs", byGroup.Cardinality())
	}

	processedSomething := false
	for _, groupID := range shared.ToStringSlice(byGroup) {
		processed, err := processTaskclusterBuild(a.aeAPI, groupID, "", sha, shared.ToStringSlice(labels)...)
		if err != nil {
			log.Errorf("Failed to process %s: %s", groupID, err.Error())
			continue
		}
		processedSomething = processedSomething || processed
	}
	return processedSomething, nil
}
