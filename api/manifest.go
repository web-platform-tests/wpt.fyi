// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

func apiManifestHandler(w http.ResponseWriter, r *http.Request) {
	sha, err := shared.ParseSHAParamFull(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(r)
	if manifest, err := getManifestForSHA(ctx, sha); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else {
		w.Header().Add("content-type", "application/json")
		w.Write(manifest)
	}
}

func gitHubSHASearchURL(sha string) string {
	return fmt.Sprintf(`https://api.github.com/search/issues?q=SHA:%s+user:w3c+repo:web-platform-tests`, sha)
}

func gitHubReleaseURL(tag string) string {
	return fmt.Sprintf(`https://api.github.com/repos/w3c/web-platform-tests/releases/tags/%s`, tag)
}

const gitHubLatestReleaseURL = `https://api.github.com/repos/w3c/web-platform-tests/releases/latest`

type gitHubClient interface {
	fetch(url string) ([]byte, error)
}

// getManifestForSHA loads the contents of the manifest JSON for the release associated with
// the given SHA, if any.
func getManifestForSHA(ctx context.Context, sha string) (manifest []byte, err error) {
	// Fetch shared.Token entity for GitHub API Token.
	tokenKey := datastore.NewKey(ctx, "Token", "github-api-token", 0, nil)
	var token shared.Token
	datastore.Get(ctx, tokenKey, &token)

	client := gitHubClientImpl{
		Token:   &token,
		Context: ctx,
	}
	return loadOrFetchManifestForSHA(ctx, &client, sha)
}

// loadOrFetchManifestForSHA gets the bytes for the SHA's release's manifest json asset (unzipped).
// The gzipped Value is stored in / loaded from memcache, to avoid unnecessary round-trips.
func loadOrFetchManifestForSHA(ctx context.Context, client gitHubClient, sha string) (manifest []byte, err error) {
	var body []byte
	cached, err := memcache.Get(ctx, sha)
	if err != nil && err != memcache.ErrCacheMiss {
		return nil, err
	} else if cached != nil {
		body = cached.Value
	} else {
		if body, err = getGitHubReleaseAssetForSHA(client, sha); err != nil {
			return nil, err
		}
		item := &memcache.Item{
			Key:   sha,
			Value: body,
		}
		// Shorter expiry for latest SHA, to keep it current.
		if sha == "latest" || sha == "" {
			item.Expiration = time.Minute * 20
		}
		memcache.Set(ctx, item)
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(gzReader)
}

// getGitHubReleaseAssetForSHA gets the bytes for the SHA's release's manifest json gzip asset.
// This is done using a few hops on the GitHub API, so should be cached afterward.
func getGitHubReleaseAssetForSHA(client gitHubClient, sha string) (manifest []byte, err error) {
	var releaseBody []byte
	var releaseTag string
	if sha == "" || sha == "latest" {
		// Use GitHub's API for latest release.
		releaseTag = "lastest"
		url := gitHubLatestReleaseURL
		if releaseBody, err = client.fetch(url); err != nil {
			return nil, err
		}
	} else {
		// Search for the PR associated with the SHA.
		url := gitHubSHASearchURL(sha)
		var body []byte
		if body, err = client.fetch(url); err != nil {
			return nil, err
		}

		var queryResults map[string]*json.RawMessage
		if err = json.Unmarshal(body, &queryResults); err != nil {
			return nil, err
		}
		var issues []map[string]*json.RawMessage
		if err = json.Unmarshal(*queryResults["items"], &issues); err != nil {
			return nil, err
		}
		if len(issues) < 1 {
			return nil, fmt.Errorf("No search results found for SHA %s", sha)
		}

		// Load the release by the presumed tag name merge_pr_*
		var prNumber int
		if err = json.Unmarshal(*issues[0]["number"], &prNumber); err != nil {
			return nil, err
		}

		releaseTag = fmt.Sprintf("merge_pr_%d", prNumber)
		url = gitHubReleaseURL(releaseTag)
		if releaseBody, err = client.fetch(url); err != nil {
			return nil, err
		}
	}

	var release map[string]*json.RawMessage
	if err = json.Unmarshal(releaseBody, &release); err != nil {
		return nil, err
	}

	var assets []map[string]*json.RawMessage
	if err = json.Unmarshal(*release["assets"], &assets); err != nil {
		return nil, err
	}
	if len(assets) < 1 {
		return nil, fmt.Errorf("No assets found for %s release", releaseTag)
	}
	// Get (and unzip) the asset with name "MANIFEST-{sha}.json.gz"
	shaMatch := sha
	if sha == "" || sha == "latest" {
		shaMatch = "[0-9a-f]{40}"
	}
	assetRegex := regexp.MustCompile(fmt.Sprintf("MANIFEST-%s.json.gz", shaMatch))
	for _, asset := range assets {
		var url string
		var name string
		var body []byte
		if err = json.Unmarshal(*asset["name"], &name); err != nil {
			return nil, err
		}
		if assetRegex.MatchString(name) {
			if err = json.Unmarshal(*asset["browser_download_url"], &url); err != nil {
				return nil, err
			}

			if body, err = client.fetch(url); err != nil {
				return nil, err
			}
			return body, err
		}
	}
	return nil, fmt.Errorf("No manifest asset found for release %s", releaseTag)
}
