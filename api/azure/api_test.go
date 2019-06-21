// +build small

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package azure_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	logrustest "github.com/Hexcles/logrus/hooks/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"

	"github.com/web-platform-tests/wpt.fyi/api/azure"
	"github.com/web-platform-tests/wpt.fyi/api/azure/mock_azure"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestHandleCheckRunEvent(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha := strings.Repeat("0123456789", 4)
	detailsURL := "https://dev.azure.com/web-platform-tests/b14026b4-9423-4454-858f-bf76cf6d1faa/_build/results?buildId=123"

	id := azure.PipelinesAppID
	chrome := "chrome"
	completed := "completed"
	created := "created"
	repoName := shared.WPTRepoName
	repoOwner := shared.WPTRepoOwner
	sender := "lukebjerring"
	event := &github.CheckRunEvent{
		Action: &created,
		CheckRun: &github.CheckRun{
			App:     &github.App{ID: &id},
			Name:    &chrome,
			Status:  &completed,
			HeadSHA: &sha,
		},
		Repo: &github.Repository{
			Name:  &repoName,
			Owner: &github.User{Login: &repoOwner},
		},
		Sender: &github.User{Login: &sender},
	}

	event.CheckRun.DetailsURL = &detailsURL

	artifact := azure.BuildArtifact{Name: "results"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/upload":
			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, username, "azure")
			assert.Equal(t, password, "123")
			w.WriteHeader(200)
		case "/123/artifacts":
			artifacts := azure.BuildArtifacts{
				Count: 1,
				Value: []azure.BuildArtifact{
					artifact,
				},
			}
			bytes, _ := json.Marshal(artifacts)
			w.Write(bytes)
		default:
			assert.FailNow(t, "Invalid spoofed request "+r.URL.String())
		}
	}))
	defer server.Close()

	build := azure.Build{
		TriggerInfo: azure.BuildTriggerInfo{
			SourceBranch: "master",
		},
	}

	azureAPI := mock_azure.NewMockAPI(mockCtrl)
	serverURL, _ := url.Parse(server.URL)
	azureAPI.EXPECT().GetAzureArtifactsURL(repoOwner, repoName, int64(123)).Return(server.URL + "/123/artifacts")
	azureAPI.EXPECT().GetBuild(repoOwner, repoName, int64(123)).Return(&build)

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().GetVersionedHostname().AnyTimes().Return(serverURL.Host)
	uploadURL, _ := url.Parse(server.URL + "/upload")
	aeAPI.EXPECT().GetResultsUploadURL().Return(uploadURL)
	aeAPI.EXPECT().GetUploader("azure").Return(shared.Uploader{Username: "azure", Password: "123"}, nil)
	aeAPI.EXPECT().GetHTTPClient().AnyTimes().Return(server.Client())
	aeAPI.EXPECT().GetSlowHTTPClient(gomock.Any()).AnyTimes().Return(server.Client(), func() {})

	log, hook := logrustest.NewNullLogger()
	ctx := context.WithValue(context.Background(), shared.DefaultLoggerCtxKey(), log)
	aeAPI.EXPECT().Context().AnyTimes().Return(ctx)
	processed, err := azure.HandleCheckRunEvent(azureAPI, aeAPI, event)
	if err != nil {
		assert.FailNow(t, "Error isn't nil", err.Error())
	}
	assert.True(t, processed)
	if len(hook.Entries) < 1 {
		assert.FailNow(t, "No logging was found")
	}
	assert.Contains(t, hook.Entries[0].Message, "/123/")
}
