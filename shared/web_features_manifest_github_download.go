// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-github/v47/github"
)

// WebFeatureManifestRepo is where the web feature manifest is published
const WebFeatureManifestRepo = "wpt"

// webFeaturesManifestFilename is the name of the manifest file in a given release
const webFeaturesManifestFilename = "WEB_FEATURES_MANIFEST.json.gz"

// ErrNoWebFeaturesManifestFileFound when a given GitHub release does not contain a manifest file to download.
var ErrNoWebFeaturesManifestFileFound = errors.New("web features manifest not found in release")

// ErrMissingBodyDuringWebFeaturesManifestDownload when a http call to download the file contains no body.
// This should not happen often.
var ErrMissingBodyDuringWebFeaturesManifestDownload = errors.New("empty body when downloading web features manifest")

// ErrUnableToRetrieveGitHubRelease indicates the request to retrieve the latest release failed
var ErrUnableToRetrieveGitHubRelease = errors.New("failed to retrieve latest github release")

// ErrGitHubAssetDownloadFailedToComplete indicates the download request failed for a given release asset
var ErrGitHubAssetDownloadFailedToComplete = errors.New("request to download github asset failed")

// gzipBodyTransformer extracts the g-zipped body into a raw file stream.
type gzipBodyTransformer struct{}

func (t gzipBodyTransformer) Transform(body io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(body)
}

// ResponseBodyTransformer provides a interface to transform the incoming response body into a
// final expected body format.
type ResponseBodyTransformer interface {
	Transform(io.Reader) (io.ReadCloser, error)
}

// NewGitHubWebFeaturesManifestDownloader creates a downloader which will examine
// a GitHub release, find the manifest file, and download it.
func NewGitHubWebFeaturesManifestDownloader(httpClient *http.Client, gitHubClient *github.Client) *GitHubWebFeaturesManifestDownloader {
	return &GitHubWebFeaturesManifestDownloader{
		httpClient:      httpClient,
		gitHubClient:    gitHubClient,
		bodyTransformer: gzipBodyTransformer{},
	}
}

// GitHubWebFeaturesManifestDownloader is a downloader that will examine
// a GitHub release, find the manifest file, and download it.
// Use NewGitHubWebFeaturesManifestDownloader to create an instance.
type GitHubWebFeaturesManifestDownloader struct {
	httpClient      *http.Client
	gitHubClient    *github.Client
	bodyTransformer ResponseBodyTransformer
}

// Download attempts to download the manifest file from the latest release.
func (d GitHubWebFeaturesManifestDownloader) Download(ctx context.Context) (io.ReadCloser, error) {
	release, _, err := d.gitHubClient.Repositories.GetLatestRelease(
		ctx,
		"jcscottiii", // REPLACE WITH SourceOwner BEFORE MERGING,
		WebFeatureManifestRepo)
	if err != nil {
		return nil, errors.Join(ErrUnableToRetrieveGitHubRelease, err)
	}

	// Find the asset URL for the manifest file
	assetURL := ""
	for _, asset := range release.Assets {
		if asset != nil && asset.Name != nil && strings.EqualFold(webFeaturesManifestFilename, *asset.Name) && asset.BrowserDownloadURL != nil {
			assetURL = *asset.BrowserDownloadURL

			break
		}
	}
	if assetURL == "" {
		return nil, ErrNoWebFeaturesManifestFileFound
	}

	// Download the file
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, ErrGitHubAssetDownloadFailedToComplete
	}
	if resp.Body == nil || resp.ContentLength == 0 {
		return nil, ErrMissingBodyDuringWebFeaturesManifestDownload
	}

	// Perform any necessary extractions / transformations
	decompressedBody, err := d.bodyTransformer.Transform(resp.Body)
	if err != nil {
		// Transformation did not happen. Clean up
		resp.Body.Close()

		return nil, err
	}

	return &gitHubDownloadStream{resp.Body, decompressedBody}, nil
}

type gitHubDownloadStream struct {
	originalBody    io.ReadCloser
	transformedBody io.ReadCloser
}

func (s *gitHubDownloadStream) Read(p []byte) (int, error) {
	return s.transformedBody.Read(p)
}

func (s *gitHubDownloadStream) Close() error {
	transformedBodyErr := s.transformedBody.Close()
	originalBodyErr := s.originalBody.Close()

	return errors.Join(transformedBodyErr, originalBodyErr)
}
