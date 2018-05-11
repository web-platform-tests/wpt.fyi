// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const fullSHA = "abcdef0123456789abcdef0123456789abcdef01"

// Shorthand for arbitrary json objects.
type object map[string]interface{}

type mockGitHubClient struct {
	Responses map[string][]byte
}

func (m *mockGitHubClient) fetch(url string) ([]byte, error) {
	if _, ok := m.Responses[url]; !ok {
		return nil, fmt.Errorf("fore! oh; for: %s", url)
	}
	return m.Responses[url], nil
}

func unsafeMarshal(i interface{}) []byte {
	result, _ := json.Marshal(i)
	return result
}

func TestGetGitHubReleaseAssetForSHA_SHANotFound(t *testing.T) {
	client := mockGitHubClient{}
	_, manifest, err := getGitHubReleaseAssetForSHA(&client, fullSHA)
	assert.Nil(t, manifest)
	assert.NotNil(t, err)
}

func TestGetGitHubReleaseAssetForSHA(t *testing.T) {
	searchResults, _ := json.Marshal(
		object{
			"items": []object{
				object{
					"number": 123,
				},
			},
		},
	)
	downloadURL := "http://github.com/magic_url"

	releaseJSON := object{
		"assets": []object{
			object{
				"name":                 fmt.Sprintf("MANIFEST-%s.json.gz", fullSHA),
				"browser_download_url": downloadURL,
			},
		},
	}

	content := "magic data"
	data := getManifestPayload(content)
	client := mockGitHubClient{
		Responses: map[string][]byte{
			gitHubSHASearchURL(fullSHA):      searchResults,
			gitHubReleaseURL("merge_pr_123"): unsafeMarshal(releaseJSON),
			downloadURL:                      data,
		},
	}

	// 1) Data is unzipped.
	_, manifest, err := getGitHubReleaseAssetForSHA(&client, fullSHA)
	assert.Nil(t, err)
	assert.Equal(t, []byte(content), manifest)

	// 2) Correct asset picked when first asset is some other asset.
	releaseJSON["assets"] = []object{
		object{
			"name":                 "Some other asset.txt",
			"browser_download_url": "http://huh.com?",
		},
		releaseJSON["assets"].([]object)[0],
	}
	client.Responses[gitHubReleaseURL("merge_pr_123")] = unsafeMarshal(releaseJSON)
	_, manifest, err = getGitHubReleaseAssetForSHA(&client, fullSHA)
	assert.Nil(t, err)
	assert.Equal(t, []byte(content), manifest)

	// 3) Error when no matching asset found.
	releaseJSON["assets"] = releaseJSON["assets"].([]object)[0:1] // Just the other asset
	client.Responses[gitHubReleaseURL("merge_pr_123")] = unsafeMarshal(releaseJSON)
	_, manifest, err = getGitHubReleaseAssetForSHA(&client, fullSHA)
	assert.NotNil(t, err)
	assert.Nil(t, manifest)
}

func TestGetGitHubReleaseAssetLatest(t *testing.T) {
	downloadURL := "http://github.com/magic_url"
	releaseJSON := object{
		"assets": []object{
			object{
				"name":                 fmt.Sprintf("MANIFEST-%s.json.gz", fullSHA),
				"browser_download_url": downloadURL,
			},
		},
	}

	content := "latest data"
	data := getManifestPayload(content)
	client := mockGitHubClient{
		Responses: map[string][]byte{
			downloadURL:            data,
			gitHubLatestReleaseURL: unsafeMarshal(releaseJSON),
		},
	}

	// Release by empty SHA or "latest" match.
	sha, latestManifest, _ := getGitHubReleaseAssetForSHA(&client, "")
	assert.Equal(t, []byte(content), latestManifest)
	assert.Equal(t, fullSHA, sha)
	sha, latestManifest, _ = getGitHubReleaseAssetForSHA(&client, "latest")
	assert.Equal(t, []byte(content), latestManifest)
	assert.Equal(t, fullSHA, sha)
}

func getManifestPayload(data string) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write([]byte(data))
	zw.Close()
	return buf.Bytes()
}
