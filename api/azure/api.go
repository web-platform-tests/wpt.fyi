// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination mock_azure/api_mock.go github.com/web-platform-tests/wpt.fyi/api/azure API

package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// PipelinesAppID is the ID of the Azure Pipelines GitHub app.
const PipelinesAppID = int64(9426)

var epochBranchesRegex = regexp.MustCompile("^refs/heads/epochs/.*")

// https://docs.microsoft.com/en-us/rest/api/azure/devops/build/artifacts/get?view=azure-devops-rest-4.1

// BuildArtifacts is a wrapper for multiple BuildArtifact results.
type BuildArtifacts struct {
	Count int64           `json:"count"`
	Value []BuildArtifact `json:"value"`
}

// BuildArtifact is an artifact published by a build.
type BuildArtifact struct {
	ID       int64            `json:"id"`
	Name     string           `json:"name"`
	Resource ArtifactResource `json:"resource"`
}

// ArtifactResource is a resource for an artifact.
type ArtifactResource struct {
	Data        string `json:"data"`
	DownloadURL string `json:"downloadUrl"`
	Type        string `json:"type"`
	URL         string `json:"url"`
}

type azureBuild struct {
	SourceBranch string                `json:"sourceBranch"`
	TriggerInfo  azureBuildTriggerInfo `json:"triggerInfo"`
}

type azureBuildTriggerInfo struct {
	SourceBranch string `json:"pr.sourceBranch"`
}

// API is for Azure Pipelines related requests.
type API interface {
	HandleCheckRunEvent(*github.CheckRunEvent) (bool, error)
	GetAzureBuildURL(owner, repo string, buildID int64) string
	GetAzureArtifactsURL(owner, repo string, buildID int64) string
	IsMasterBranch(owner, repo string, buildID int64) bool
}

type apiImpl struct {
	ctx context.Context
}

// NewAPI returns an implementation of azure API
func NewAPI(ctx context.Context) API {
	return apiImpl{
		ctx: ctx,
	}
}

// HandleCheckRunEvent processes an Azure Pipelines check run "completed" event.
func (a apiImpl) HandleCheckRunEvent(checkRun *github.CheckRunEvent) (bool, error) {
	return handleCheckRunEvent(a, shared.NewAppEngineAPI(a.ctx), checkRun)
}

func (a apiImpl) GetAzureBuildURL(owner, repo string, buildID int64) string {
	// https://docs.microsoft.com/en-us/rest/api/azure/devops/build/builds/get?view=azure-devops-rest-4.1#build
	return fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/build/builds/%v", owner, repo, buildID)
}

func (a apiImpl) GetAzureArtifactsURL(owner, repo string, buildID int64) string {
	return fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/build/builds/%v/artifacts",
		owner,
		repo,
		buildID)
}

func (a apiImpl) IsMasterBranch(owner, repo string, buildID int64) bool {
	buildURL := a.GetAzureBuildURL(owner, repo, buildID)
	client := shared.NewAppEngineAPI(a.ctx).GetHTTPClient()
	log := shared.GetLogger(a.ctx)
	resp, err := client.Get(buildURL)
	if err != nil {
		log.Errorf("Failed to fetch build: %s", err.Error())
		return false
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read request response: %s", err.Error())
		return false
	}
	var build azureBuild
	if err := json.Unmarshal(data, &build); err != nil {
		log.Errorf("Failed to unmarshal request response: %s", err.Error())
		return false
	}
	log.Debugf("Source branch: %s", build.SourceBranch)
	log.Debugf("Trigger PR branch: %s", build.TriggerInfo.SourceBranch)
	return epochBranchesRegex.MatchString(build.SourceBranch) ||
		build.TriggerInfo.SourceBranch == "master"
}
