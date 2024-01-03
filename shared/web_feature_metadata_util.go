// Copyright 2023 The WPT Dashboard Project. All rights reserved.
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

const webFeaturesManifestFilename = "WEB_FEATURES_MANIFEST.json.gz"

// ErrNoWebFeaturesManifestFileFound when a given GitHub release does not contain a manifest file to download.
var ErrNoWebFeaturesManifestFileFound = errors.New("Web Features Manifest not found in release")

// ErrMissingBodyDuringWebFeaturesManifestDownload when a http call to download the file contains no body.
// This should not happen often.
var ErrMissingBodyDuringWebFeaturesManifestDownload = errors.New("empty body when downloading web features manifest")

// GetWPTWebFeaturesManifest is the entrypoint for handling web features manifest downloads.
// It includes two generic steps: Downloading and Parsing.
func GetWPTWebFeaturesManifest(ctx context.Context, downloader WebFeaturesManifestDownloader, parser WebFeatureManifestParser) (WebFeaturesData, error) {
	manifest, err := downloader.Download(ctx)
	if err != nil {
		return nil, nil
	}
	data, err := parser.Parse(ctx, manifest)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// WebFeaturesManifestDownloader provides an interface for downloading a manifest file.
// Typically implementers of the interface create a struct based on the location of the file.
// For example: GitHub.
type WebFeaturesManifestDownloader interface {
	Download(context.Context) (io.ReadCloser, error)
}

// WebFeatureManifestParser provides an interface on how to parse a given manifest.
// Typically implementers of the interface create a struct per type of file.
// For example: JSON file or YAML file
type WebFeatureManifestParser interface {
	// Parse reads the stream of data and returns a map.
	// The map mirrors the structure of MetadataResults where the key is the
	// test name and the value is the actual data. In this case the "data" is a
	// collection of applicable web features
	Parse(context.Context, io.ReadCloser) (WebFeaturesData, error)
}

// NewGitHubWebFeaturesManifestDownloader creates a downloader which will examine
// a GitHub release, find the manifest file, and download it.
func NewGitHubWebFeaturesManifestDownloader(httpClient *http.Client, gitHubClient *github.Client) *GitHubWebFeaturesManifestDownloader {
	return &GitHubWebFeaturesManifestDownloader{
		httpClient:   httpClient,
		gitHubClient: gitHubClient,
	}
}

// GitHubWebFeaturesManifestDownloader is a downloader that will examine
// a GitHub release, find the manifest file, and download it.
type GitHubWebFeaturesManifestDownloader struct {
	httpClient   *http.Client
	gitHubClient *github.Client
}

// Download attempts to download the manifest file from the latest release.
func (d GitHubWebFeaturesManifestDownloader) Download(ctx context.Context) (io.ReadCloser, error) {
	release, _, err := d.gitHubClient.Repositories.GetLatestRelease(ctx, SourceOwner, WebFeatureManifestRepo)
	if err != nil {
		return nil, err
	}
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if assetURL == "" {
		return nil, err
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, ErrMissingBodyDuringWebFeaturesManifestDownload
	}
	defer resp.Body.Close()

	decompressedBody, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return decompressedBody, nil
}
