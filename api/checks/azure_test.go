// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestHandleAzurePipelinesEvent(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("0123456789", 4)
	detailsURL := "https://dev.azure.com/web-platform-tests/b14026b4-9423-4454-858f-bf76cf6d1faa/_build/results?buildId=123"
	event := getCheckRunCreatedEvent("completed", "lukebjerring", sha)
	event.CheckRun.DetailsURL = &detailsURL

	log, hook := logrustest.NewNullLogger()
	processed, err := handleAzurePipelinesEvent(log, &http.Client{}, event)
	assert.Nil(t, err)
	assert.False(t, processed)
	assert.Len(t, hook.Entries, 1)
	assert.Contains(t, hook.Entries[0].Message, "/123/")
}

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
}

func TestDownloadURL(t *testing.T) {
	var artifacts BuildArtifacts
	err := json.Unmarshal([]byte(artifactsJSON), &artifacts)
	assert.Nil(t, err)
	org := "lukebjerring"
	url, err := artifacts.Value[1].GetReportURL(org)
	assert.Nil(t, err)
	assert.Equal(
		t,
		"https://dev.azure.com/lukebjerring/_apis/resources/Containers/1714875?itemPath=results%2Fwpt_report.json",
		url.String())
}
