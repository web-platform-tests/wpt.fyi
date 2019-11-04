// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package taskcluster

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v28/github"
	"github.com/stretchr/testify/assert"
	"github.com/taskcluster/taskcluster/clients/client-go/v20/tcqueue"
	uc "github.com/web-platform-tests/wpt.fyi/api/receiver/client"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func strPtr(s string) *string {
	return &s
}

func TestShouldProcessStatus_states(t *testing.T) {
	status := statusEventPayload{}
	status.State = strPtr("success")
	status.Context = strPtr("Taskcluster")
	status.Branches = branchInfos{&github.Branch{Name: strPtr(shared.MasterLabel)}}
	assert.True(t, shouldProcessStatus(shared.NewNilLogger(), false, &status))

	status.State = strPtr("failure")
	assert.True(t, shouldProcessStatus(shared.NewNilLogger(), false, &status))

	status.State = strPtr("error")
	assert.False(t, shouldProcessStatus(shared.NewNilLogger(), false, &status))

	status.State = strPtr("pending")
	assert.False(t, shouldProcessStatus(shared.NewNilLogger(), false, &status))
}

func TestShouldProcessStatus_notTaskcluster(t *testing.T) {
	status := statusEventPayload{}
	status.State = strPtr("success")
	status.Context = strPtr("Travis")
	status.Branches = branchInfos{&github.Branch{Name: strPtr(shared.MasterLabel)}}
	assert.False(t, shouldProcessStatus(shared.NewNilLogger(), false, &status))
}

func TestShouldProcessStatus_notOnMaster(t *testing.T) {
	status := statusEventPayload{}
	status.State = strPtr("success")
	status.Context = strPtr("Taskcluster")
	status.Branches = branchInfos{&github.Branch{Name: strPtr("gh-pages")}}
	assert.False(t, shouldProcessStatus(shared.NewNilLogger(), false, &status))
	assert.True(t, shouldProcessStatus(shared.NewNilLogger(), true, &status))
}

func TestIsOnMaster(t *testing.T) {
	status := statusEventPayload{}
	status.SHA = strPtr("a10867b14bb761a232cd80139fbd4c0d33264240")
	status.State = strPtr("success")
	status.Context = strPtr("Taskcluster")
	status.Branches = branchInfos{
		&github.Branch{
			Name:   strPtr(shared.MasterLabel),
			Commit: &github.RepositoryCommit{SHA: strPtr("a10867b14bb761a232cd80139fbd4c0d33264240")},
		},
		&github.Branch{
			Name:   strPtr("changes"),
			Commit: &github.RepositoryCommit{SHA: strPtr("34c5c7793cb3b279e22454cb6750c80560547b3a")},
		},
		&github.Branch{
			Name:   strPtr("gh-pages"),
			Commit: &github.RepositoryCommit{SHA: strPtr("fd353d4ae7c19d2268397459524f849c129944a7")},
		},
	}
	assert.Equal(t, []string{shared.MasterLabel}, status.HeadingBranches().GetNames())
	assert.True(t, status.IsOnMaster())

	status.Branches = status.Branches[1:]
	assert.False(t, status.IsOnMaster())
}

func TestParseTaskclusterURL(t *testing.T) {
	t.Run("Status", func(t *testing.T) {
		root, group, task := parseTaskclusterURL("https://tc.example.com/task-group-inspector/#/Y4rnZeqDRXGiRNiqxT5Qeg")
		assert.Equal(t, "https://tc.example.com", root)
		assert.Equal(t, "Y4rnZeqDRXGiRNiqxT5Qeg", group)
		assert.Equal(t, "", task)
	})
	t.Run("CheckRun with task", func(t *testing.T) {
		root, group, task := parseTaskclusterURL("https://tc.example.com/groups/IWlO7NuxRnO0_8PKMuHFkw/tasks/NOToWHr0T-u62B9yGQnD5w/details")
		assert.Equal(t, "https://tc.example.com", root)
		assert.Equal(t, "IWlO7NuxRnO0_8PKMuHFkw", group)
		assert.Equal(t, "NOToWHr0T-u62B9yGQnD5w", task)
	})
	t.Run("CheckRun without task", func(t *testing.T) {
		root, group, task := parseTaskclusterURL("https://tc.other-example.com/groups/IWlO7NuxRnO0_8PKMuHFkw")
		assert.Equal(t, "https://tc.other-example.com", root)
		assert.Equal(t, "IWlO7NuxRnO0_8PKMuHFkw", group)
		assert.Equal(t, "", task)
	})
}

func TestExtractArtifactURLs_all_success_master(t *testing.T) {
	group := &taskGroupInfo{Tasks: make([]tcqueue.TaskDefinitionAndStatus, 4)}
	group.Tasks[0].Task.Metadata.Name = "wpt-firefox-nightly-testharness-1"
	group.Tasks[1].Task.Metadata.Name = "wpt-firefox-nightly-testharness-2"
	group.Tasks[2].Task.Metadata.Name = "wpt-chrome-dev-testharness-1"
	group.Tasks[3].Task.Metadata.Name = "wpt-chrome-dev-reftest-1"
	for i := 0; i < len(group.Tasks); i++ {
		group.Tasks[i].Status.State = "completed"
		group.Tasks[i].Status.TaskID = fmt.Sprint(i)
	}

	t.Run("All", func(t *testing.T) {
		urls, err := extractArtifactURLs("https://tc.example.com", shared.NewNilLogger(), group, "")
		assert.Nil(t, err)
		assert.Equal(t, map[string]artifactURLs{
			"firefox-nightly": {
				Results: []string{
					"https://tc.example.com/api/queue/v1/task/0/artifacts/public/results/wpt_report.json.gz",
					"https://tc.example.com/api/queue/v1/task/1/artifacts/public/results/wpt_report.json.gz",
				},
				Screenshots: []string{
					"https://tc.example.com/api/queue/v1/task/0/artifacts/public/results/wpt_screenshot.txt.gz",
					"https://tc.example.com/api/queue/v1/task/1/artifacts/public/results/wpt_screenshot.txt.gz",
				},
			},
			"chrome-dev": {
				Results: []string{
					"https://tc.example.com/api/queue/v1/task/2/artifacts/public/results/wpt_report.json.gz",
					"https://tc.example.com/api/queue/v1/task/3/artifacts/public/results/wpt_report.json.gz",
				},
				Screenshots: []string{
					"https://tc.example.com/api/queue/v1/task/2/artifacts/public/results/wpt_screenshot.txt.gz",
					"https://tc.example.com/api/queue/v1/task/3/artifacts/public/results/wpt_screenshot.txt.gz",
				},
			},
		}, urls)
	})

	t.Run("Filtered", func(t *testing.T) {
		urls, err := extractArtifactURLs("https://tc.example.com", shared.NewNilLogger(), group, "0")
		assert.Nil(t, err)
		assert.Equal(t, map[string]artifactURLs{
			"firefox-nightly": {
				Results: []string{
					"https://tc.example.com/api/queue/v1/task/0/artifacts/public/results/wpt_report.json.gz",
				},
				Screenshots: []string{
					"https://tc.example.com/api/queue/v1/task/0/artifacts/public/results/wpt_screenshot.txt.gz",
				},
			},
		}, urls)
	})
}

func TestExtractArtifactURLs_all_success_pr(t *testing.T) {
	group := &taskGroupInfo{Tasks: make([]tcqueue.TaskDefinitionAndStatus, 3)}
	group.Tasks[0].Task.Metadata.Name = "wpt-chrome-dev-results"
	group.Tasks[1].Task.Metadata.Name = "wpt-chrome-dev-stability"
	group.Tasks[2].Task.Metadata.Name = "wpt-chrome-dev-results-without-changes"
	for i := 0; i < len(group.Tasks); i++ {
		group.Tasks[i].Status.State = "completed"
		group.Tasks[i].Status.TaskID = fmt.Sprint(i)
	}

	t.Run("All", func(t *testing.T) {
		urls, err := extractArtifactURLs("https://tc.example.com", shared.NewNilLogger(), group, "")
		assert.Nil(t, err)
		assert.Equal(t, map[string]artifactURLs{
			"chrome-dev-pr_head": {
				Results: []string{
					"https://tc.example.com/api/queue/v1/task/0/artifacts/public/results/wpt_report.json.gz",
				},
				Screenshots: []string{
					"https://tc.example.com/api/queue/v1/task/0/artifacts/public/results/wpt_screenshot.txt.gz",
				},
			},
			"chrome-dev-pr_base": {
				Results: []string{
					"https://tc.example.com/api/queue/v1/task/2/artifacts/public/results/wpt_report.json.gz",
				},
				Screenshots: []string{
					"https://tc.example.com/api/queue/v1/task/2/artifacts/public/results/wpt_screenshot.txt.gz",
				},
			},
		}, urls)
	})

	t.Run("Filtered", func(t *testing.T) {
		urls, err := extractArtifactURLs("https://tc.example.com", shared.NewNilLogger(), group, "2")
		assert.Nil(t, err)
		assert.Equal(t, map[string]artifactURLs{
			"chrome-dev-pr_base": {
				Results: []string{
					"https://tc.example.com/api/queue/v1/task/2/artifacts/public/results/wpt_report.json.gz",
				},
				Screenshots: []string{
					"https://tc.example.com/api/queue/v1/task/2/artifacts/public/results/wpt_screenshot.txt.gz",
				},
			},
		}, urls)
	})
}

func TestExtractArtifactURLs_with_failures(t *testing.T) {
	group := &taskGroupInfo{Tasks: make([]tcqueue.TaskDefinitionAndStatus, 3)}
	group.Tasks[0].Status.State = "failed"
	group.Tasks[0].Status.TaskID = "foo"
	group.Tasks[0].Task.Metadata.Name = "wpt-firefox-nightly-testharness-1"
	group.Tasks[1].Status.State = "completed"
	group.Tasks[1].Status.TaskID = "bar"
	group.Tasks[1].Task.Metadata.Name = "wpt-firefox-nightly-testharness-2"
	group.Tasks[2].Status.State = "completed"
	group.Tasks[2].Status.TaskID = "baz"
	group.Tasks[2].Task.Metadata.Name = "wpt-chrome-dev-testharness-1"

	urls, err := extractArtifactURLs("https://tc.example.com", shared.NewNilLogger(), group, "")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(urls))
	assert.Contains(t, urls, "chrome-dev")
}

func TestCreateAllRuns_success(t *testing.T) {
	var requested uint32
	requested = 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&requested, 1)
		w.Write([]byte("OK"))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()
	serverURL, _ := url.Parse(server.URL)
	sha := "abcdef1234abcdef1234abcdef1234abcdef1234"

	mockC := gomock.NewController(t)
	defer mockC.Finish()
	aeAPI := sharedtest.NewMockAppEngineAPI(mockC)
	aeAPI.EXPECT().GetVersionedHostname().AnyTimes().Return("localhost:8080")
	aeAPI.EXPECT().GetSlowHTTPClient(uc.UploadTimeout).AnyTimes().Return(&http.Client{}, func() {})
	aeAPI.EXPECT().GetResultsUploadURL().AnyTimes().Return(serverURL)

	t.Run("master", func(t *testing.T) {
		err := createAllRuns(
			shared.NewNilLogger(),
			aeAPI,
			sha,
			"username",
			"password",
			map[string]artifactURLs{
				"safari-preview": {Results: []string{"1"}},
				"chrome-dev":     {Results: []string{"1"}},
				"firefox-stable": {Results: []string{"1", "2"}},
			},
			[]string{shared.MasterLabel, "user:person"},
		)
		assert.Nil(t, err)
		assert.Equal(t, uint32(3), requested)
	})

	requested = 0
	t.Run("PR", func(t *testing.T) {
		err := createAllRuns(
			shared.NewNilLogger(),
			aeAPI,
			sha,
			"username",
			"password",
			map[string]artifactURLs{
				"chrome-dev-pr_head":     {Results: []string{"1"}},
				"chrome-dev-pr_base":     {Results: []string{"1"}},
				"firefox-stable-pr_head": {Results: []string{"1"}},
				"firefox-stable-pr_base": {Results: []string{"1"}},
			},
			nil,
		)
		assert.Nil(t, err)
		assert.Equal(t, uint32(4), requested)
	})
}

func TestCreateAllRuns_one_error(t *testing.T) {
	var requested uint32
	requested = 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		if atomic.CompareAndSwapUint32(&requested, 0, 1) {
			w.Write([]byte("OK"))
		} else if atomic.CompareAndSwapUint32(&requested, 1, 2) {
			http.Error(w, "Not found", http.StatusNotFound)
		} else {
			assert.FailNow(t, "requested != 0 && requested != 1")
		}
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()
	mockC := gomock.NewController(t)
	defer mockC.Finish()

	sha := "abcdef1234abcdef1234abcdef1234abcdef1234"

	aeAPI := sharedtest.NewMockAppEngineAPI(mockC)
	aeAPI.EXPECT().GetVersionedHostname().MinTimes(1).Return("localhost:8080")
	aeAPI.EXPECT().GetSlowHTTPClient(uc.UploadTimeout).Times(2).Return(&http.Client{}, func() {})
	serverURL, _ := url.Parse(server.URL)
	aeAPI.EXPECT().GetResultsUploadURL().AnyTimes().Return(serverURL)

	err := createAllRuns(
		shared.NewNilLogger(),
		aeAPI,
		sha,
		"username",
		"password",
		map[string]artifactURLs{
			"chrome":  {Results: []string{"1"}},
			"firefox": {Results: []string{"1", "2"}},
		},
		[]string{shared.MasterLabel},
	)
	assert.NotNil(t, err)
	assert.Equal(t, uint32(2), requested)
	assert.Contains(t, err.Error(), "API error:")
	assert.Contains(t, err.Error(), "404")
}

func TestCreateAllRuns_all_errors(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 2)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()
	mockC := gomock.NewController(t)
	defer mockC.Finish()

	sha := "abcdef1234abcdef1234abcdef1234abcdef1234"

	aeAPI := sharedtest.NewMockAppEngineAPI(mockC)
	aeAPI.EXPECT().GetVersionedHostname().MinTimes(1).Return("localhost:8080")
	// Give a very short timeout (instead of the asked 1min) to make tests faster.
	aeAPI.EXPECT().GetSlowHTTPClient(uc.UploadTimeout).MinTimes(1).Return(&http.Client{Timeout: time.Microsecond}, func() {})
	serverURL, _ := url.Parse(server.URL)
	aeAPI.EXPECT().GetResultsUploadURL().AnyTimes().Return(serverURL)

	err := createAllRuns(
		shared.NewNilLogger(),
		aeAPI,
		sha,
		"username",
		"password",
		map[string]artifactURLs{
			"chrome":  {Results: []string{"1"}},
			"firefox": {Results: []string{"1", "2"}},
		},
		[]string{shared.MasterLabel},
	)
	assert.NotNil(t, err)
	assert.Equal(t, 2, strings.Count(err.Error(), "Client.Timeout"))
}

func TestTaskNameRegex(t *testing.T) {
	assert.Equal(t, []string{"chrome-dev", "results"}, taskNameRegex.FindStringSubmatch("wpt-chrome-dev-results")[1:])
	assert.Equal(t, []string{"chrome-stable", "reftest"}, taskNameRegex.FindStringSubmatch("wpt-chrome-stable-reftest-1")[1:])
	assert.Equal(t, []string{"firefox-nightly", "testharness"}, taskNameRegex.FindStringSubmatch("wpt-firefox-nightly-testharness-5")[1:])
	assert.Equal(t, []string{"firefox-stable", "wdspec"}, taskNameRegex.FindStringSubmatch("wpt-firefox-stable-wdspec-1")[1:])
	assert.Equal(t, []string{"chrome-dev", "results-without-changes"}, taskNameRegex.FindStringSubmatch("wpt-chrome-dev-results-without-changes")[1:])
	assert.Nil(t, taskNameRegex.FindStringSubmatch("wpt-chrome-dev-stability"))
}
