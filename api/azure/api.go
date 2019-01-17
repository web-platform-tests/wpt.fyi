// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package azure

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	multipart "mime/multipart"
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
	FetchAzureArtifact(BuildArtifact, io.Writer) error
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
func (a apiImpl) FetchAzureArtifact(artifact BuildArtifact, writer io.Writer) error {
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

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response body")
		return err
	}

	// Extract the report from the artifact.
	return extractReports(a.ctx, artifact.Name, data, writer)
}

// extractReports extracts report files from the given zip.
func extractReports(ctx context.Context, artifactName string, data []byte, writer io.Writer) error {
	log := shared.GetLogger(ctx)
	reportPath, err := regexp.Compile(fmt.Sprintf(`%s/wpt_report.*\.json$`, artifactName))
	if err != nil {
		return err
	}
	extracted := 0
	mWriter := multipart.NewWriter(writer)
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	for _, f := range z.File {
		if reportPath.MatchString(f.Name) {
			// Wrap extraction in a function, to scope the "defer fileData.Close()"
			if err := func() error {
				fileName := f.Name[len(artifactName)+1:]
				fileField, err := mWriter.CreateFormFile("result_file", fileName)
				var fileData io.ReadCloser
				if fileData, err = f.Open(); err != nil {
					log.Errorf("Failed to extract %s", f.Name)
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
			}(); err != nil {
				return err
			}
			extracted++
		}
	}
	if extracted < 1 {
		return errors.New(`No "wpt_report.*\.json" files found in zip`)
	}
	return mWriter.Close()
}

func getCheckTitle(product shared.ProductSpec) string {
	return fmt.Sprintf("wpt.fyi - %s results", product.DisplayName())
}
