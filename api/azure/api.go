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

// Build is an Azure Pipelines build object.
type Build struct {
	SourceBranch string           `json:"sourceBranch"`
	HeadSHA      string           `json:"sourceVersion"`
	TriggerInfo  BuildTriggerInfo `json:"triggerInfo"`
}

// BuildTriggerInfo is information about what triggered the build.
type BuildTriggerInfo struct {
	SourceBranch string `json:"pr.sourceBranch"`
}

// IsMasterBranch returns whether the source branch for the build is the master branch.
func (a *Build) IsMasterBranch() bool {
	return a != nil && a.TriggerInfo.SourceBranch == "master"
}

// API is for Azure Pipelines related requests.
type API interface {
	HandleCheckRunEvent(*github.CheckRunEvent) (bool, error)
	GetBuildURL(owner, repo string, buildID int64) string
	GetAzureArtifactsURL(owner, repo string, buildID int64) string
	GetBuild(owner, repo string, buildID int64) *Build
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
	return HandleCheckRunEvent(a, shared.NewAppEngineAPI(a.ctx), checkRun)
}

func (a apiImpl) GetBuildURL(owner, repo string, buildID int64) string {
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

func (a apiImpl) GetBuild(owner, repo string, buildID int64) *Build {
	buildURL := a.GetBuildURL(owner, repo, buildID)
	client := shared.NewAppEngineAPI(a.ctx).GetHTTPClient()
	log := shared.GetLogger(a.ctx)
	resp, err := client.Get(buildURL)
	if err != nil {
		log.Errorf("Failed to fetch build: %s", err.Error())
		return nil
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read request response: %s", err.Error())
		return nil
	}
	var build Build
	if err := json.Unmarshal(data, &build); err != nil {
		log.Errorf("Failed to unmarshal request response: %s", err.Error())
		return nil
	}
	log.Debugf("Source branch: %s", build.SourceBranch)
	log.Debugf("Trigger PR branch: %s", build.TriggerInfo.SourceBranch)
	return &build
}
