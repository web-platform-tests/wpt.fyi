// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package azure

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"regexp"
	"time"

	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// PipelinesAppID is the ID of the Azure Pipelines GitHub app.
const PipelinesAppID = int64(9426)

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

// API is for Azure Pipelines related requests.
type API interface {
	HandleCheckRunEvent(*github.CheckRunEvent) (bool, error)
	GetAzureArtifactsURL(owner, repo string, buildID int64) string
	FetchAzureArtifact(BuildArtifact, *multipart.Writer) error
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
	return handleCheckRunEvent(
		shared.GetLogger(a.ctx),
		a,
		shared.NewAppEngineAPI(a.ctx),
		checkRun)
}

func (a apiImpl) GetAzureArtifactsURL(owner, repo string, buildID int64) string {
	return fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/build/builds/%v/artifacts",
		owner,
		repo,
		buildID)
}

// FetchAzureArtifact gets the gzipped bytes of the wpt_report.json from inside
// the zip file provided by Azure, and writes them to the given writer.
func (a apiImpl) FetchAzureArtifact(artifact BuildArtifact, writer *multipart.Writer) error {
	aeAPI := shared.NewAppEngineAPI(a.ctx)
	log := shared.GetLogger(a.ctx)
	// The default timeout is 5s, not enough to download the reports.
	client, cancel := aeAPI.GetSlowHTTPClient(time.Minute)
	defer cancel()
	resp, err := client.Get(artifact.Resource.DownloadURL)
	if err != nil {
		log.Errorf("Failed to fetch %s: %s", artifact.Resource.DownloadURL, err.Error())
		return err
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Errorf("Failed to fetch %s: %s", artifact.Resource.DownloadURL, resp.Status)
		return err
	}

	// Extract the report from the artifact.
	reportPath, err := regexp.Compile(fmt.Sprintf(`%s/wpt_report.*\.json$`, artifact.Name))
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response body")
		return err
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	for _, f := range z.File {
		if reportPath.MatchString(f.Name) {
			fileName := f.Name[len(artifact.Name)+1:]
			fileField, err := writer.CreateFormFile("result_file", fileName)
			var fileData io.ReadCloser
			if fileData, err = f.Open(); err != nil {
				log.Errorf("Failed to extract %s", reportPath)
				return err
			}
			defer fileData.Close()

			gzw := gzip.NewWriter(fileField)
			if _, err := io.Copy(gzw, fileData); err != nil {
				log.Errorf("Failed to gzip file contents")
				return err
			}
			if err := gzw.Close(); err != nil {
				log.Errorf("Failed to close gzip writer")
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("File %s not found in zip", reportPath)
}

func getCheckTitle(product shared.ProductSpec) string {
	return fmt.Sprintf("wpt.fyi - %s results", product.DisplayName())
}
