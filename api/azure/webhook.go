// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package azure

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"time"

	mapset "github.com/deckarep/golang-set"

	uc "github.com/web-platform-tests/wpt.fyi/api/receiver/client"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

const uploaderName = "azure"

// Labels for runs from Azure Pipelines are determined from the artifact names.
// For master runs, artifact name may be either just "results" or something
// like "safari-results".
var (
	masterRegex        = regexp.MustCompile(`\bresults$`)
	prHeadRegex        = regexp.MustCompile(`\baffected-tests$`)
	prBaseRegex        = regexp.MustCompile(`\baffected-tests-without-changes$`)
	epochBranchesRegex = regexp.MustCompile("^refs/heads/epochs/.*")
)

func processBuild(aeAPI shared.AppEngineAPI, azureAPI API, owner, repo, sender, artifactName string, buildID int64) (bool, error) {
	build := azureAPI.GetBuild(owner, repo, buildID)
	sha := ""
	if build != nil {
		sha = build.TriggerInfo.SourceSHA
	}

	// https://docs.microsoft.com/en-us/rest/api/azure/devops/build/artifacts/get?view=azure-devops-rest-4.1
	artifactsURL := azureAPI.GetAzureArtifactsURL(owner, repo, buildID)

	log := shared.GetLogger(aeAPI.Context())
	log.Infof("Fetching %s", artifactsURL)

	slowClient, cancel := aeAPI.GetSlowHTTPClient(time.Minute)
	defer cancel()
	resp, err := slowClient.Get(artifactsURL)
	if err != nil {
		log.Errorf("Failed to fetch artifacts for %s/%s build %v", owner, repo, buildID)
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

	uploadedAny := false
	errors := make(chan (error), artifacts.Count)
	for _, artifact := range artifacts.Value {
		if artifactName != "" && artifactName != artifact.Name {
			log.Infof("Skipping artifact %s (looking for %s)", artifact.Name, artifactName)
			continue
		}
		log.Infof("Uploading %s for %s/%s build %v...", artifact.Name, owner, repo, buildID)

		labels := mapset.NewSet()
		if sender != "" {
			labels.Add(shared.GetUserLabel(sender))
		}

		if masterRegex.MatchString(artifact.Name) {
			if build.IsMasterBranch() || epochBranchesRegex.MatchString(build.SourceBranch) {
				labels.Add(shared.MasterLabel)
			}
		} else if prHeadRegex.MatchString(artifact.Name) {
			labels.Add(shared.PRHeadLabel)
		} else if prBaseRegex.MatchString(artifact.Name) {
			labels.Add(shared.PRBaseLabel)
		}

		uploader, err := aeAPI.GetUploader(uploaderName)
		if err != nil {
			log.Errorf("Failed to get uploader creds from Datastore")
			return false, err
		}

		uploadClient := uc.NewClient(aeAPI)
		err = uploadClient.CreateRun(
			sha,
			uploader.Username,
			uploader.Password,
			[]string{artifact.Resource.DownloadURL},
			shared.ToStringSlice(labels))
		if err != nil {
			log.Errorf("Failed to create run: %s", err.Error())
			errors <- err
		} else {
			uploadedAny = true
		}
	}
	close(errors)
	for err := range errors {
		return uploadedAny, err
	}
	return uploadedAny, nil
}
