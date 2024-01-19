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

// WebFeatureManifestRepo is where the web feature manifest is published.
const WebFeatureManifestRepo = "wpt"

// webFeaturesManifestFilename is the name of the manifest file in a given release.
const webFeaturesManifestFilename = "WEB_FEATURES_MANIFEST.json.gz"

// errNoWebFeaturesManifestFileFound when a given GitHub release does not contain a manifest file to download.
var errNoWebFeaturesManifestFileFound = errors.New("web features manifest not found in release")

// errMissingBodyDuringWebFeaturesManifestDownload when a http call to download the file contains no body.
// This should not happen often.
var errMissingBodyDuringWebFeaturesManifestDownload = errors.New("empty body when downloading web features manifest")

// errUnableToRetrieveGitHubRelease indicates the request to retrieve the latest release failed.
var errUnableToRetrieveGitHubRelease = errors.New("failed to retrieve latest GitHub release")

// errGitHubAssetDownloadFailedToComplete indicates the download request failed for a given release asset.
var errGitHubAssetDownloadFailedToComplete = errors.New("request to download GitHub asset failed")

// gzipBodyTransformer extracts the g-zipped body into a raw file stream.
type gzipBodyTransformer struct{}

func (t gzipBodyTransformer) Transform(body io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(body)
}

// responseBodyTransformer provides a interface to transform the incoming response body into a
// final expected body format.
type responseBodyTransformer interface {
	Transform(io.Reader) (io.ReadCloser, error)
}

// NewGitHubWebFeaturesManifestDownloader creates a downloader which will examine
// a GitHub release, find the manifest file, and download it.
func NewGitHubWebFeaturesManifestDownloader(
	httpClient *http.Client,
	repoReleaseGetter RepositoryReleaseGetter) *GitHubWebFeaturesManifestDownloader {
	return &GitHubWebFeaturesManifestDownloader{
		httpClient:        httpClient,
		repoReleaseGetter: repoReleaseGetter,
		bodyTransformer:   gzipBodyTransformer{},
	}
}

// RepositoryReleaseGetter provides an interface to retrieve releases for a given repository.
type RepositoryReleaseGetter interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
}

// GitHubWebFeaturesManifestDownloader is a downloader that will examine
// a GitHub release, find the manifest file, and download it.
// Use NewGitHubWebFeaturesManifestDownloader to create an instance.
type GitHubWebFeaturesManifestDownloader struct {
	httpClient        *http.Client
	repoReleaseGetter RepositoryReleaseGetter
	bodyTransformer   responseBodyTransformer
}

// Download attempts to download the manifest file from the latest release.
func (d GitHubWebFeaturesManifestDownloader) Download(ctx context.Context) (io.ReadCloser, error) {
	release, _, err := d.repoReleaseGetter.GetLatestRelease(
		ctx,
		"jcscottiii", // REPLACE WITH SourceOwner BEFORE MERGING,
		WebFeatureManifestRepo)
	if err != nil {
		return nil, errors.Join(errUnableToRetrieveGitHubRelease, err)
	}

	// Find the asset URL for the manifest file.
	assetURL := ""
	for _, asset := range release.Assets {
		if asset != nil && asset.Name != nil &&
			strings.EqualFold(webFeaturesManifestFilename, *asset.Name) &&
			asset.BrowserDownloadURL != nil {
			assetURL = *asset.BrowserDownloadURL

			break
		}
	}
	if assetURL == "" {
		return nil, errNoWebFeaturesManifestFileFound
	}

	// Download the file.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, errGitHubAssetDownloadFailedToComplete
	}
	if resp.Body == nil || resp.ContentLength == 0 {
		return nil, errMissingBodyDuringWebFeaturesManifestDownload
	}

	// Perform any necessary extractions / transformations.
	decompressedBody, err := d.bodyTransformer.Transform(resp.Body)
	if err != nil {
		// Transformation did not happen. Clean up by closing any open resources.
		resp.Body.Close()

		return nil, err
	}

	return &gitHubDownloadStream{resp.Body, decompressedBody}, nil
}

// Instead of passing copies of []byte, this struct contains the raw stream of data
// That way it can be read once.
// This struct implements io.ReadCloser so callers are responsible for calling Close().
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
