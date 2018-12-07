// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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

// ArtifactResource is a resource for an artifact.
type ArtifactResource struct {
	Data        string `json:"data"`
	DownloadURL string `json:"downloadUrl"`
	Type        string `json:"type"`
	URL         string `json:"url"`
}

func handleAzurePipelinesEvent(log shared.Logger, checksAPI API, aeAPI shared.AppEngineAPI, event github.CheckRunEvent) (bool, error) {
	status := event.GetCheckRun().GetStatus()
	if status != "completed" {
		log.Infof("Ignoring non-completed status %s", status)
		return false, nil
	}
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	detailsURL := event.GetCheckRun().GetDetailsURL()
	buildID := extractAzureBuildID(detailsURL)
	if buildID == 0 {
		log.Errorf("Failed to extract build ID from details_url \"%s\"", detailsURL)
		return false, nil
	}

	// https://docs.microsoft.com/en-us/rest/api/azure/devops/build/artifacts/get?view=azure-devops-rest-4.1
	artifactsURL := checksAPI.GetAzureArtifactsURL(owner, repo, buildID)
	log.Infof("Fetching %s", artifactsURL)

	client := aeAPI.GetHTTPClient()
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

	uploadedAny := false
	errors := make(chan (error), artifacts.Count)
	for _, artifact := range artifacts.Value {
		if err != nil {
			log.Errorf("Failed to extract report URL: %s", err.Error())
			continue
		}
		log.Infof("Uploading %s for %s/%s build %v...", artifact.Name, owner, repo, buildID)

		err := createAzureRun(
			log,
			aeAPI,
			event.GetCheckRun().GetHeadSHA(),
			artifact,
			nil)
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

func extractAzureBuildID(detailsURL string) int64 {
	parsedURL, err := url.Parse(detailsURL)
	if err != nil {
		return 0
	}
	id := parsedURL.Query().Get("buildId")
	parsedID, _ := strconv.ParseInt(id, 0, 0)
	return parsedID
}

func createAzureRun(
	log shared.Logger,
	aeAPI shared.AppEngineAPI,
	sha string,
	artifact BuildArtifact,
	labels []string) error {
	// https://github.com/web-platform-tests/wpt.fyi/blob/master/api/README.md#url-payload
	payload := make(url.Values)
	// Not to be confused with `revision` in the wpt.fyi TestRun model, this
	// parameter is the full revision hash.
	payload.Add("revision", sha)
	if len(labels) > 0 {
		payload.Add("labels", strings.Join(labels, ","))
	}
	// Ensure we call back to this appengine version instance.
	host := aeAPI.GetHostname()
	payload.Add("callback_url", fmt.Sprintf("https://%s/api/results/create", host))

	// The default timeout is 5s, not enough for the receiver to process the reports.
	client, cancel := aeAPI.GetSlowHTTPClient(time.Minute)
	defer cancel()
	resp, err := client.Get(artifact.Resource.DownloadURL)
	if err != nil {
		log.Errorf("Failed to fetch %s: %s", artifact.Resource.DownloadURL, err.Error())
		return err
	}

	// Extract the report from the artifact.
	reportPath := fmt.Sprintf("%s/wpt_report.json", artifact.Name)
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response body", err.Error())
		return err
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	for _, f := range z.File {
		if f.Name == reportPath {
			fileData, err := f.Open()
			if err != nil {
				log.Errorf("Failed to extract %s", reportPath)
				return err
			}
			var buf bytes.Buffer
			gzw := gzip.NewWriter(&buf)
			fileContents, err := ioutil.ReadAll(fileData)
			if err != nil {
				log.Errorf("Failed to read zip file")
				return err
			}
			if _, err := gzw.Write(fileContents); err != nil {
				log.Errorf("Failed to gzip file contents")
				return err
			}
			payload.Add("result_file", buf.String())
		}
	}

	req, err := http.NewRequest(
		"POST",
		aeAPI.GetResultsUploadURL().String(),
		strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}
	uploader, err := aeAPI.GetUploader("azure")
	if err != nil {
		log.Errorf("Failed to load azure uploader")
		return err
	}
	req.SetBasicAuth(uploader.Username, uploader.Password)
	if resp, err = client.Do(req); err != nil {
		log.Errorf("Failed to send upload request")
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Errorf("Failed to read response")
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", string(respBody))
	}

	return nil
}
