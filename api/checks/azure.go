// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

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

// GetReportURL produces a URL for fetching the report file, if applicable.
func (b BuildArtifact) GetReportURL(organization string) (*url.URL, error) {
	containerID := b.Resource.GetContainerID()
	if containerID == "" {
		return nil, errors.New("Failed to extract container ID")
	}
	result, err := url.Parse(fmt.Sprintf(
		"https://dev.azure.com/%s/_apis/resources/Containers/%v",
		organization,
		containerID))
	if err != nil {
		return nil, err
	}
	q := result.Query()
	q.Set("itemPath", fmt.Sprintf("%s/wpt_report.json", b.Name))
	result.RawQuery = q.Encode()
	return result, err
}

// ArtifactResource is a resource for an artifact.
type ArtifactResource struct {
	Data        string `json:"data"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	URL         string `json:"url"`
}

// GetContainerID extracts the container ID from the artifact resource.
func (a ArtifactResource) GetContainerID() string {
	match := regexp.MustCompile(`#/(\d+)/\w+`).FindStringSubmatch(a.Data)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func handleAzurePipelinesEvent(log shared.Logger, client *http.Client, event github.CheckRunEvent) (bool, error) {
	status := event.GetCheckRun().GetStatus()
	if status != "completed" {
		log.Infof("Ignoring non-completed status %s", status)
		return false, nil
	}
	detailsURL := event.GetCheckRun().GetDetailsURL()
	buildID := extractAzureBuildID(detailsURL)
	if buildID == "" {
		log.Errorf("Failed to extract build ID from details_url \"%s\"", detailsURL)
		return false, nil
	}
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	artifactsURL := fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/build/builds/%s/artifacts",
		owner,
		repo,
		buildID)
	log.Infof("Fetching %s", artifactsURL)

	resp, err := client.Get(artifactsURL)
	if err != nil {
		log.Errorf("Failed to fetch artifacts for %s/%s build %s", owner, repo, buildID)
		return false, err
	}

	var artifacts BuildArtifacts
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		log.Errorf("Failed to read response body")
		return false, err
	} else if err = json.Unmarshal(body, &artifacts); err != nil {
		log.Errorf("Failed to unmarshal JSON")
		return false, err
	}

	for _, artifact := range artifacts.Value {
		reportURL, err := artifact.GetReportURL(owner)
		if err != nil {
			log.Errorf("Failed to extract report URL: %s", err.Error())
			continue
		}
		log.Infof("Uploading %s for %s/%s build %s...", artifact.Name, owner, repo, buildID)
		log.Warningf("TODO: Not actually uploading %s", reportURL)
	}
	return false, nil
}
