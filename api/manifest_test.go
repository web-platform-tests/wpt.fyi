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
	const sha = "abcdef1234"
	manifest, err := getGitHubReleaseAssetForSHA(&client, sha)
	assert.Nil(t, manifest)
	assert.NotNil(t, err)
}

func TestGetGitHubReleaseAssetForSHA(t *testing.T) {
	const sha = "abcdef1234"
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

	releasesJSON := object{
		"assets": []object{
			object{
				"name":                 "MANIFEST-abcdef1234.json.gz",
				"browser_download_url": downloadURL,
			},
		},
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write([]byte("magic data"))
	zw.Close()
	data := buf.Bytes()

	client := mockGitHubClient{
		Responses: map[string][]byte{
			gitHubSHASearchURL(sha):          searchResults,
			gitHubReleaseURL("merge_pr_123"): unsafeMarshal(releasesJSON),
			downloadURL:                      data,
		},
	}

	// 1) Data is unzipped.
	manifest, err := getGitHubReleaseAssetForSHA(&client, sha)
	assert.Nil(t, err)
	assert.Equal(t, []byte("magic data"), manifest)

	// 2) Correct asset picked when first asset is some other asset.
	releasesJSON["assets"] = []object{
		object{
			"name":                 "Some other asset.txt",
			"browser_download_url": "http://huh.com?",
		},
		releasesJSON["assets"].([]object)[0],
	}
	client.Responses[gitHubReleaseURL("merge_pr_123")] = unsafeMarshal(releasesJSON)
	manifest, err = getGitHubReleaseAssetForSHA(&client, sha)
	assert.Nil(t, err)
	assert.Equal(t, []byte("magic data"), manifest)

	// 3) Error when no matching asset found.
	releasesJSON["assets"] = releasesJSON["assets"].([]object)[0:1] // Just the other asset
	client.Responses[gitHubReleaseURL("merge_pr_123")] = unsafeMarshal(releasesJSON)
	manifest, err = getGitHubReleaseAssetForSHA(&client, sha)
	assert.NotNil(t, err)
	assert.Nil(t, manifest)
}
