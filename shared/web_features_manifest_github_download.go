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
	"time"

	"github.com/google/go-github/v73/github"
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

// newGitHubWebFeaturesManifestDownloader creates a downloader which will examine
// a GitHub release, find the manifest file, and download it.
func newGitHubWebFeaturesManifestDownloader(
	httpClient *http.Client,
	repoReleaseGetter repositoryReleaseGetter) *gitHubWebFeaturesManifestDownloader {
	return &gitHubWebFeaturesManifestDownloader{
		httpClient:        httpClient,
		repoReleaseGetter: repoReleaseGetter,
		bodyTransformer:   gzipBodyTransformer{},
	}
}

// repositoryReleaseGetter provides an interface to retrieve releases for a given repository.
type repositoryReleaseGetter interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
}

// gitHubWebFeaturesManifestDownloader is a downloader that will examine
// a GitHub release, find the manifest file, and download it.
// Use newGitHubWebFeaturesManifestDownloader to create an instance.
type gitHubWebFeaturesManifestDownloader struct {
	httpClient        *http.Client
	repoReleaseGetter repositoryReleaseGetter
	bodyTransformer   responseBodyTransformer
}

// Download attempts to download the manifest file from the latest release.
func (d gitHubWebFeaturesManifestDownloader) Download(ctx context.Context) (io.ReadCloser, error) {
	release, _, err := d.repoReleaseGetter.GetLatestRelease(
		ctx,
		SourceOwner,
		WebFeatureManifestRepo)
	if err != nil {
		return nil, errors.Join(errUnableToRetrieveGitHubRelease, err)
	}

	// Find the asset URL for the manifest file.
	assetURL := ""
	for _, asset := range release.Assets {
		if asset != nil && asset.Label != nil &&
			strings.EqualFold(webFeaturesManifestFilename, *asset.Label) &&
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

// GitHubWebFeaturesClient is the entrypoint to retrieving web features from GitHub.
type GitHubWebFeaturesClient struct {
	downloader webFeaturesManifestDownloader
	parser     webFeatureManifestParser
}

// gitHubWebFeaturesClientOptions contains all the non-required options that
// can be used to configure an instance of GitHubWebFeaturesClient.
type gitHubWebFeaturesClientOptions struct {
	netClient *http.Client
}

// A GitHubWebFeaturesClientOption configures GitHubWebFeaturesClient.
type GitHubWebFeaturesClientOption func(*gitHubWebFeaturesClientOptions)

// SetHTTPClientForGitHubWebFeatures overrides the http client used to download
// the found asset from the release.
func SetHTTPClientForGitHubWebFeatures(netClient *http.Client) GitHubWebFeaturesClientOption {
	return func(opts *gitHubWebFeaturesClientOptions) {
		opts.netClient = netClient
	}
}

// nolint:gochecknoglobals // non exported variable that contains default values.
var defaultGitHubWebFeaturesClientOptions = []GitHubWebFeaturesClientOption{
	SetHTTPClientForGitHubWebFeatures(&http.Client{
		Timeout: time.Second * 5,
	}),
}

// NewGitHubWebFeaturesClient constructs an instance of GitHubWebFeaturesClient
// with default values.
func NewGitHubWebFeaturesClient(ghClient *github.Client) *GitHubWebFeaturesClient {
	var options gitHubWebFeaturesClientOptions
	for _, opt := range defaultGitHubWebFeaturesClientOptions {
		opt(&options)
	}
	// For now, use the default options. In the future, we could pass in
	// variadic options and override them here.

	downloader := newGitHubWebFeaturesManifestDownloader(options.netClient, ghClient.Repositories)

	return &GitHubWebFeaturesClient{downloader: downloader, parser: webFeaturesManifestJSONParser{}}
}

// Get returns the latest web features data from GitHub.
func (c GitHubWebFeaturesClient) Get(ctx context.Context) (WebFeaturesData, error) {
	data, err := getWPTWebFeaturesManifest(ctx, c.downloader, c.parser)
	if err != nil {
		GetLogger(ctx).Errorf("unable to fetch web features manifest during query. %s", err.Error())

		return nil, err
	}

	return data, nil
}
