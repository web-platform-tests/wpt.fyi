// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// handleCheckRunEvent processes an Azure Pipelines check run "completed" event.
func handleCheckRunEvent(log shared.Logger, azureAPI API, aeAPI shared.AppEngineAPI, event *github.CheckRunEvent) (bool, error) {
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
	artifactsURL := azureAPI.GetAzureArtifactsURL(owner, repo, buildID)
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

		var labels []string
		if artifact.Name == "results" {
			labels = []string{shared.PRHeadLabel}
		} else if artifact.Name == "results-without-changes" {
			labels = []string{shared.PRBaseLabel}
		}
		err := createAzureRun(
			log,
			azureAPI,
			aeAPI,
			event.GetCheckRun().GetHeadSHA(),
			artifact,
			labels)
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
	azureAPI API,
	aeAPI shared.AppEngineAPI,
	sha string,
	artifact BuildArtifact,
	labels []string) error {
	// https://github.com/web-platform-tests/wpt.fyi/blob/master/api/README.md#url-payload
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	// Not to be confused with `revision` in the wpt.fyi TestRun model, this
	// parameter is the full revision hash.
	writer.WriteField("revision", sha)
	if len(labels) > 0 {
		writer.WriteField("labels", strings.Join(labels, ","))
	}
	// Ensure we call back to this appengine version instance.
	host := aeAPI.GetHostname()
	writer.WriteField("callback_url", fmt.Sprintf("https://%s/api/results/create", host))

	fileField, err := writer.CreateFormFile("result_file", "wpt_report.json")
	if err = azureAPI.FetchAzureArtifact(artifact, fileField); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", aeAPI.GetResultsUploadURL().String(), buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	uploader, err := aeAPI.GetUploader("azure")
	if err != nil {
		log.Errorf("Failed to load azure uploader")
		return err
	}
	req.SetBasicAuth(uploader.Username, uploader.Password)

	// The default timeout is 5s, not enough for the receiver to process the reports.
	client, cancel := aeAPI.GetSlowHTTPClient(time.Minute)
	defer cancel()
	resp, err := client.Do(req)
	if err != nil {
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
