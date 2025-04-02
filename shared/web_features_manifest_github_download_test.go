// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build small

package shared

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v70/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var compressedWebFeaturesManifestFilePath = filepath.Join("web_features_manifest_testdata", "WEB_FEATURES_MANIFEST.json.gz")

func createWebFeaturesTestdata() {
	v1Manifest := struct {
		Version int                 `json:"version,omitempty"`
		Data    map[string][]string `json:"data,omitempty"`
	}{
		Version: 1,
		Data: map[string][]string{
			"grid":    {"test1.js", "test2.js"},
			"subgrid": {"test3.js", "test4.js"},
		},
	}
	jsonData, err := json.Marshal(v1Manifest)
	if err != nil {
		panic(err)
	}

	// Create a buffer for compressing the JSON.
	var buf bytes.Buffer

	// Create a gzip writer and write the JSON to it.
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}

	// Write the compressed data to a file.
	if err := os.WriteFile(compressedWebFeaturesManifestFilePath, buf.Bytes(), 0644); err != nil {
		panic(err)
	}
}

func TestResponseBodyTransformer_Success(t *testing.T) {
	updateGolden := false // Switch this when we want to update the golden file.
	if updateGolden {
		createWebFeaturesTestdata()
	}
	f, err := os.Open(compressedWebFeaturesManifestFilePath)
	defer f.Close()
	require.NoError(t, err)

	transformer := gzipBodyTransformer{}
	reader, err := transformer.Transform(f)
	defer reader.Close()
	require.NoError(t, err)

	rawBytes, err := io.ReadAll(reader)
	require.NoError(t, err)

	assert.Equal(t, `{"version":1,"data":{"grid":["test1.js","test2.js"],"subgrid":["test3.js","test4.js"]}}`, string(rawBytes))
}

type RoundTripFunc struct {
	function func(req *http.Request) *http.Response
	err      error
}

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.function(req), f.err
}

type mockBodyTransformerInput struct {
	expectedBody string
	output       io.ReadCloser
	err          error
}

type mockBodyTransformer struct {
	t *testing.T
	mockBodyTransformerInput
}

func (tr mockBodyTransformer) Transform(body io.Reader) (io.ReadCloser, error) {
	bodyBytes, err := io.ReadAll(body)
	require.NoError(tr.t, err)
	assert.Equal(tr.t, tr.expectedBody, string(bodyBytes))
	return tr.output, tr.err
}

type mockRepositoryReleaseGetter struct {
	t *testing.T
	mockRepositoryReleaseGetterInput
}

type mockRepositoryReleaseGetterInput struct {
	expectedOwner string
	expectedRepo  string
	repoRelease   *github.RepositoryRelease
	resp          *github.Response
	err           error
}

func (g mockRepositoryReleaseGetter) GetLatestRelease(
	ctx context.Context,
	owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
	require.Equal(g.t, g.expectedOwner, owner)
	require.Equal(g.t, g.expectedRepo, repo)
	return g.repoRelease, g.resp, g.err
}

/*
Truncated example output from GitHub API.
Useful for building the returned responses in TestGitHubWebFeaturesManifestDownloader_Download.

gh release view --repo web-platform-tests/wpt --json assets
{
  "assets": [
    {
      "apiUrl": "https://api.github.com/repos/web-platform-tests/wpt/releases/assets/147533430",
      "contentType": "application/octet-stream",
      "createdAt": "2024-01-24T14:40:18Z",
      "downloadCount": 0,
      "id": "RA_kwDOADc1Vc4Iyy52",
      "label": "WEB_FEATURES_MANIFEST.json.gz",
      "name": "WEB_FEATURES_MANIFEST-f8871bc568c2cf86b38cb70f28a9d5f707e19259.json.gz",
      "size": 38815,
      "state": "uploaded",
      "updatedAt": "2024-01-24T14:40:18Z",
      "url": "https://github.com/web-platform-tests/wpt/releases/download/merge_pr_41522/WEB_FEATURES_MANIFEST-f8871bc568c2cf86b38cb70f28a9d5f707e19259.json.gz"
    }
  ]
}
*/

func TestGitHubWebFeaturesManifestDownloader_Download(t *testing.T) {
	// Test cases for Download
	tests := []struct {
		name               string
		releaseGetterInput mockRepositoryReleaseGetterInput
		roundTrip          RoundTripFunc
		transformer        mockBodyTransformerInput
		expectedBody       []byte
		expectedError      error
	}{
		{
			name: "successful download",
			releaseGetterInput: mockRepositoryReleaseGetterInput{
				expectedOwner: "web-platform-tests",
				expectedRepo:  "wpt",
				repoRelease: &github.RepositoryRelease{
					Assets: []*github.ReleaseAsset{
						{
							Label:              github.String("WEB_FEATURES_MANIFEST.json.gz"),
							BrowserDownloadURL: github.String("https://example.com/WEB_FEATURES_MANIFEST.json.gz"),
						},
					},
				},
				resp: &github.Response{},
				err:  nil,
			},
			roundTrip: RoundTripFunc{function: func(req *http.Request) *http.Response {
				assert.Equal(t, "https://example.com/WEB_FEATURES_MANIFEST.json.gz", req.URL.String())
				return &http.Response{
					StatusCode:    http.StatusOK,
					ContentLength: int64(len("raw data")),
					Body:          io.NopCloser(bytes.NewBufferString("raw data")),
				}
			}, err: nil},
			transformer: mockBodyTransformerInput{
				expectedBody: "raw data",
				output:       io.NopCloser(bytes.NewBufferString("transformed data")),
				err:          nil,
			},
			expectedBody:  []byte("transformed data"),
			expectedError: nil,
		},
		{
			name: "error getting latest release",
			releaseGetterInput: mockRepositoryReleaseGetterInput{
				expectedOwner: "web-platform-tests",
				expectedRepo:  "wpt",
				repoRelease:   nil,
				resp:          nil,
				err:           errors.New("fake GitHub client error"),
			},
			expectedBody:  nil,
			expectedError: errUnableToRetrieveGitHubRelease,
		},
		{
			name: "manifest file not found",
			releaseGetterInput: mockRepositoryReleaseGetterInput{
				expectedOwner: "web-platform-tests",
				expectedRepo:  "wpt",
				repoRelease: &github.RepositoryRelease{
					Assets: []*github.ReleaseAsset{},
				},
				resp: &github.Response{},
				err:  nil,
			},
			expectedBody:  nil,
			expectedError: errNoWebFeaturesManifestFileFound,
		},
		{
			name: "error downloading asset",
			releaseGetterInput: mockRepositoryReleaseGetterInput{
				expectedOwner: "web-platform-tests",
				expectedRepo:  "wpt",
				repoRelease: &github.RepositoryRelease{
					Assets: []*github.ReleaseAsset{
						{
							Label:              github.String("WEB_FEATURES_MANIFEST.json.gz"),
							BrowserDownloadURL: github.String("https://example.com/WEB_FEATURES_MANIFEST.json.gz"),
						},
					},
				},
				resp: &github.Response{},
				err:  nil,
			},
			roundTrip: RoundTripFunc{function: func(req *http.Request) *http.Response {
				assert.Equal(t, "https://example.com/WEB_FEATURES_MANIFEST.json.gz", req.URL.String())
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
				}
			}, err: errors.New("simulated network error")},
			expectedBody:  nil,
			expectedError: errGitHubAssetDownloadFailedToComplete,
		},
		{
			name: "empty response body",
			releaseGetterInput: mockRepositoryReleaseGetterInput{
				expectedOwner: "web-platform-tests",
				expectedRepo:  "wpt",
				repoRelease: &github.RepositoryRelease{
					Assets: []*github.ReleaseAsset{
						{
							Label:              github.String("WEB_FEATURES_MANIFEST.json.gz"),
							BrowserDownloadURL: github.String("https://example.com/WEB_FEATURES_MANIFEST.json.gz"),
						},
					},
				},
				resp: &github.Response{},
				err:  nil,
			},
			roundTrip: RoundTripFunc{function: func(req *http.Request) *http.Response {
				assert.Equal(t, "https://example.com/WEB_FEATURES_MANIFEST.json.gz", req.URL.String())
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       nil,
				}
			}, err: nil},
			expectedBody:  nil,
			expectedError: errMissingBodyDuringWebFeaturesManifestDownload,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			getter := mockRepositoryReleaseGetter{t, tc.releaseGetterInput}
			httpClient := &http.Client{
				Transport: tc.roundTrip,
			}
			downloader := newGitHubWebFeaturesManifestDownloader(httpClient, getter)
			downloader.bodyTransformer = mockBodyTransformer{t, tc.transformer}
			body, err := downloader.Download(context.Background())
			if !errors.Is(err, tc.expectedError) {
				t.Errorf("Download() returned unexpected error: (%v). expected error: (%v).", err, tc.expectedError)
			}

			// No need to compare the body if there's an error.
			if err != nil {
				return
			}

			bodyBytes, err := io.ReadAll(body)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedBody, bodyBytes)
		})
	}
}
