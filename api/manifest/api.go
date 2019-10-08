// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination mock_manifest/api_mock.go github.com/web-platform-tests/wpt.fyi/api/manifest API

package manifest

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"

	"github.com/google/go-github/v28/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// API handles manifest-related fetches and caching.
type API interface {
	GetManifestForSHA(string) (string, []byte, error)
}

type apiImpl struct {
	ctx context.Context
}

// NewAPI returns an API implementation for the given context.
func NewAPI(ctx context.Context) API {
	return apiImpl{
		ctx: ctx,
	}
}

// GetManifestForSHA loads the (gzipped) contents of the manifest JSON for the release associated
// with the given SHA, if any.
func (a apiImpl) GetManifestForSHA(sha string) (fetchedSHA string, manifest []byte, err error) {
	aeAPI := shared.NewAppEngineAPI(a.ctx)
	fetchedSHA, body, err := getGitHubReleaseAssetForSHA(aeAPI, sha)
	if err != nil {
		return fetchedSHA, nil, err
	}
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return fetchedSHA, nil, err
	}
	return fetchedSHA, data, err
}

// getGitHubReleaseAssetForSHA gets the bytes for the SHA's release's manifest json gzip asset.
// This is done using a few hops on the GitHub API, so should be cached afterward.
func getGitHubReleaseAssetForSHA(aeAPI shared.AppEngineAPI, sha string) (fetchedSHA string, manifest io.Reader, err error) {
	client, err := aeAPI.GetGitHubClient()
	if err != nil {
		return "", nil, err
	}
	var release *github.RepositoryRelease
	releaseTag := "latest"
	fetchedSHA = sha
	if shared.IsLatest(sha) {
		// Use GitHub's API for latest release.
		release, _, err = client.Repositories.GetLatestRelease(aeAPI.Context(), shared.WPTRepoOwner, shared.WPTRepoName)
	} else {
		q := fmt.Sprintf("SHA:%s user:web-platform-tests repo:wpt", sha)
		issues, _, err := client.Search.Issues(aeAPI.Context(), q, nil)
		if err != nil {
			return fetchedSHA, nil, err
		}
		if issues == nil || len(issues.Issues) < 1 {
			return fetchedSHA, nil, fmt.Errorf("No search results found for SHA %s", sha)
		}

		releaseTag = fmt.Sprintf("merge_pr_%d", issues.Issues[0].GetNumber())
		release, _, err = client.Repositories.GetReleaseByTag(aeAPI.Context(), shared.WPTRepoOwner, shared.WPTRepoName, releaseTag)
	}

	if err != nil {
		return fetchedSHA, nil, err
	} else if release == nil || len(release.Assets) < 1 {
		return fetchedSHA, nil, fmt.Errorf("No assets found for %s release", releaseTag)
	}
	// Get (and unzip) the asset with name "MANIFEST-{sha}.json.gz"
	shaMatch := sha
	if sha == "" || sha == "latest" {
		shaMatch = "[0-9a-f]{40}"
	}
	assetRegex := regexp.MustCompile(fmt.Sprintf("MANIFEST-(%s).json.gz", shaMatch))
	for _, asset := range release.Assets {
		name := asset.GetName()
		var url string
		if assetRegex.MatchString(name) {
			fetchedSHA = assetRegex.FindStringSubmatch(name)[1]
			url = asset.GetBrowserDownloadURL()

			client := aeAPI.GetHTTPClient()
			resp, err := client.Get(url)
			if err != nil {
				return fetchedSHA, nil, err
			}
			return fetchedSHA, resp.Body, err
		}
	}
	return fetchedSHA, nil, fmt.Errorf("No manifest asset found for release %s", releaseTag)
}
