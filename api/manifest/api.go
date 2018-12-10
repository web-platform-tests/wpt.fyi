// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"time"

	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/memcache"
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

// GetManifestForSHA loads the contents of the manifest JSON for the release associated with
// the given SHA, if any.
func (a apiImpl) GetManifestForSHA(sha string) (fetchedSHA string, manifest []byte, err error) {
	aeAPI := shared.NewAppEngineAPI(a.ctx)
	if sha == "" {
		sha = "latest"
	}
	fetchedSHA = sha
	cached, err := memcache.Get(a.ctx, manifestCacheKey(sha))

	var body io.Reader
	if err != nil && err != memcache.ErrCacheMiss {
		return "", nil, err
	} else if cached != nil {
		// "latest" caches which SHA is latest; Return the manifest for that SHA.
		if sha == "latest" {
			return a.GetManifestForSHA(string(cached.Value))
		}
		body = bytes.NewReader(cached.Value)
	}

	if fetchedSHA, body, err = getGitHubReleaseAssetForSHA(aeAPI, sha); err != nil {
		return fetchedSHA, nil, err
	}
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return fetchedSHA, nil, err
	}
	item := &memcache.Item{
		Key:   manifestCacheKey(fetchedSHA),
		Value: data,
	}
	memcache.Set(a.ctx, item)

	// Shorter expiry for latest SHA, to keep it current.
	if sha == "latest" {
		latestSHAItem := &memcache.Item{
			Key:        manifestCacheKey("latest"),
			Value:      []byte(fetchedSHA),
			Expiration: time.Minute * 5,
		}
		memcache.Set(a.ctx, latestSHAItem)
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fetchedSHA, nil, err
	}
	manifest, err = ioutil.ReadAll(gzReader)
	return fetchedSHA, manifest, err
}

// getGitHubReleaseAssetForSHA gets the bytes for the SHA's release's manifest json gzip asset.
// This is done using a few hops on the GitHub API, so should be cached afterward.
func getGitHubReleaseAssetForSHA(aeAPI shared.AppEngineAPI, sha string) (fetchedSHA string, manifest io.Reader, err error) {
	client, err := aeAPI.GetGitHubClient()
	if err != nil {
		return "", nil, err
	}
	var release *github.RepositoryRelease
	var releaseTag string
	fetchedSHA = sha
	if shared.IsLatest(sha) {
		// Use GitHub's API for latest release.
		releaseTag = "latest"
		release, _, err = client.Repositories.GetReleaseByTag(aeAPI.Context(), "web-platform-tests", "wpt", releaseTag)
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
		release, _, err = client.Repositories.GetReleaseByTag(aeAPI.Context(), "web-platform-tests", "wpt", releaseTag)
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

func manifestCacheKey(sha string) string {
	return fmt.Sprintf("MANIFEST-%s", sha)
}
