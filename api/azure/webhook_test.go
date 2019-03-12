// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package azure

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const artifactsJSON = `{
	"count": 2,
	"value": [{
		"id": 1,
		"name": "results-without-patch",
		"resource": {
			"type": "Container",
			"data": "#/1714875/results-without-patch",
			"properties": {
				"localpath": "/Users/vsts/agent/2.142.1/work/1/a/wpt_report.json"
			},
			"url": "https://dev.azure.com/lukebjerring/92272aaf-ee0f-48f4-8c22-c1fa6648843c/_apis/build/builds/4/artifacts?artifactName=results-without-patch&api-version=5.0",
			"downloadUrl": "https://dev.azure.com/lukebjerring/92272aaf-ee0f-48f4-8c22-c1fa6648843c/_apis/build/builds/4/artifacts?artifactName=results-without-patch&api-version=5.0&%24format=zip"
		}
	}, {
		"id": 2,
		"name": "results",
		"resource": {
			"type": "Container",
			"data": "#/1714875/results",
			"properties": {
				"localpath": "/Users/vsts/agent/2.142.1/work/1/a/wpt_report.json"
			},
			"url": "https://dev.azure.com/lukebjerring/92272aaf-ee0f-48f4-8c22-c1fa6648843c/_apis/build/builds/4/artifacts?artifactName=results&api-version=5.0",
			"downloadUrl": "https://dev.azure.com/lukebjerring/92272aaf-ee0f-48f4-8c22-c1fa6648843c/_apis/build/builds/4/artifacts?artifactName=results&api-version=5.0&%24format=zip"
		}
	}]
}`

func TestParses(t *testing.T) {
	var artifacts BuildArtifacts
	err := json.Unmarshal([]byte(artifactsJSON), &artifacts)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), artifacts.Count)
	assert.Len(t, artifacts.Value, 2)
	for _, artifact := range artifacts.Value {
		assert.NotEmpty(t, artifact.Resource.DownloadURL)
	}
}

func TestArtifactRegexes(t *testing.T) {
	// Names before https://github.com/web-platform-tests/wpt/pull/15110
	assert.True(t, masterRegex.MatchString("results"))
	assert.True(t, prHeadRegex.MatchString("affected-tests"))
	assert.True(t, prBaseRegex.MatchString("affected-tests-without-changes"))

	// Names after https://github.com/web-platform-tests/wpt/pull/15110
	assert.True(t, masterRegex.MatchString("edge-results"))
	assert.True(t, prHeadRegex.MatchString("safari-preview-affected-tests"))
	assert.True(t, prBaseRegex.MatchString("safari-preview-affected-tests-without-changes"))

	// Don't accept the other order
	assert.False(t, masterRegex.MatchString("results-edge"))

	// Don't accept any string ending with the right pattern
	assert.False(t, masterRegex.MatchString("nodashresults"))

	// Base and Head could be confused with substring matching
	assert.False(t, prBaseRegex.MatchString("affected-tests"))
	assert.False(t, prHeadRegex.MatchString("affected-tests-without-changes"))
}

func TestEpochBranchesRegex(t *testing.T) {
	assert.True(t, epochBranchesRegex.MatchString("refs/heads/epochs/twelve_hourly"))
	assert.True(t, epochBranchesRegex.MatchString("refs/heads/epochs/six_hourly"))
	assert.True(t, epochBranchesRegex.MatchString("refs/heads/epochs/weekly"))
	assert.True(t, epochBranchesRegex.MatchString("refs/heads/epochs/daily"))

	assert.False(t, epochBranchesRegex.MatchString("refs/heads/weekly"))
}
