//go:build small
// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package taskcluster_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/stretchr/testify/assert"
	uc "github.com/web-platform-tests/wpt.fyi/api/receiver/client"
	tc "github.com/web-platform-tests/wpt.fyi/api/taskcluster"
	mock_tc "github.com/web-platform-tests/wpt.fyi/api/taskcluster/mock_taskcluster"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"go.uber.org/mock/gomock"
)

type branchInfos []*github.Branch

func strPtr(s string) *string {
	return &s
}

func TestShouldProcessStatus_states(t *testing.T) {
	status := tc.StatusEventPayload{}
	status.State = strPtr("success")
	status.Context = strPtr("Taskcluster")
	status.Branches = branchInfos{&github.Branch{Name: strPtr(shared.MasterLabel)}}
	assert.True(t, tc.ShouldProcessStatus(shared.NewNilLogger(), &status))

	status.Context = strPtr("Community-TC")
	assert.True(t, tc.ShouldProcessStatus(shared.NewNilLogger(), &status))

	status.State = strPtr("failure")
	assert.True(t, tc.ShouldProcessStatus(shared.NewNilLogger(), &status))

	status.State = strPtr("error")
	assert.False(t, tc.ShouldProcessStatus(shared.NewNilLogger(), &status))

	status.State = strPtr("pending")
	assert.False(t, tc.ShouldProcessStatus(shared.NewNilLogger(), &status))
}

func TestShouldProcessStatus_notTaskcluster(t *testing.T) {
	status := tc.StatusEventPayload{}
	status.State = strPtr("success")
	status.Context = strPtr("Travis")
	status.Branches = branchInfos{&github.Branch{Name: strPtr(shared.MasterLabel)}}
	assert.False(t, tc.ShouldProcessStatus(shared.NewNilLogger(), &status))
}

func TestShouldProcessStatus_notOnMaster(t *testing.T) {
	status := tc.StatusEventPayload{}
	status.State = strPtr("success")
	status.Context = strPtr("Taskcluster")
	status.Branches = branchInfos{&github.Branch{Name: strPtr("gh-pages")}}
	assert.True(t, tc.ShouldProcessStatus(shared.NewNilLogger(), &status))
}

func TestIsOnMaster(t *testing.T) {
	status := tc.StatusEventPayload{}
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
	assert.True(t, status.IsOnMaster())

	status.Branches = status.Branches[1:]
	assert.False(t, status.IsOnMaster())
}

func TestParseTaskclusterURL(t *testing.T) {
	t.Run("Status", func(t *testing.T) {
		root, group, task := tc.ParseTaskclusterURL("https://tools.taskcluster.net/task-group-inspector/#/Y4rnZeqDRXGiRNiqxT5Qeg")
		assert.Equal(t, "https://taskcluster.net", root)
		assert.Equal(t, "Y4rnZeqDRXGiRNiqxT5Qeg", group)
		assert.Equal(t, "", task)
	})
	t.Run("CheckRun with task", func(t *testing.T) {
		root, group, task := tc.ParseTaskclusterURL("https://tc.example.com/groups/IWlO7NuxRnO0_8PKMuHFkw/tasks/NOToWHr0T-u62B9yGQnD5w/details")
		assert.Equal(t, "https://tc.example.com", root)
		assert.Equal(t, "IWlO7NuxRnO0_8PKMuHFkw", group)
		assert.Equal(t, "NOToWHr0T-u62B9yGQnD5w", task)
	})
	t.Run("CheckRun without task", func(t *testing.T) {
		root, group, task := tc.ParseTaskclusterURL("https://tc.other-example.com/groups/IWlO7NuxRnO0_8PKMuHFkw")
		assert.Equal(t, "https://tc.other-example.com", root)
		assert.Equal(t, "IWlO7NuxRnO0_8PKMuHFkw", group)
		assert.Equal(t, "", task)
	})
	t.Run("CheckRun without task", func(t *testing.T) {
		root, group, task := tc.ParseTaskclusterURL("https://tc.community.com/tasks/groups/IWlO7NuxRnO0_8PKMuHFkw")
		assert.Equal(t, "https://tc.community.com", root)
		assert.Equal(t, "IWlO7NuxRnO0_8PKMuHFkw", group)
		assert.Equal(t, "", task)
	})
}

func TestExtractArtifactURLs_all_success_master(t *testing.T) {
	group := &tc.TaskGroupInfo{Tasks: make([]tc.TaskInfo, 5)}
	group.Tasks[0].Name = "wpt-firefox-nightly-testharness-1"
	group.Tasks[1].Name = "wpt-firefox-nightly-testharness-2"
	group.Tasks[2].Name = "wpt-chrome-dev-testharness-1"
	group.Tasks[3].Name = "wpt-chrome-dev-reftest-1"
	group.Tasks[4].Name = "wpt-chrome-dev-crashtest-1"
	for i := 0; i < len(group.Tasks); i++ {
		group.Tasks[i].State = "completed"
		group.Tasks[i].TaskID = fmt.Sprint(i)
	}

	t.Run("All", func(t *testing.T) {
		urls, err := tc.ExtractArtifactURLs("https://tc.example.com", shared.NewNilLogger(), group, "")
		assert.Nil(t, err)
		assert.Equal(t, map[string]tc.ArtifactURLs{
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
					"https://tc.example.com/api/queue/v1/task/4/artifacts/public/results/wpt_report.json.gz",
				},
				Screenshots: []string{
					"https://tc.example.com/api/queue/v1/task/2/artifacts/public/results/wpt_screenshot.txt.gz",
					"https://tc.example.com/api/queue/v1/task/3/artifacts/public/results/wpt_screenshot.txt.gz",
					"https://tc.example.com/api/queue/v1/task/4/artifacts/public/results/wpt_screenshot.txt.gz",
				},
			},
		}, urls)
	})

	t.Run("Filtered", func(t *testing.T) {
		urls, err := tc.ExtractArtifactURLs("https://tc.example.com", shared.NewNilLogger(), group, "0")
		assert.Nil(t, err)
		assert.Equal(t, map[string]tc.ArtifactURLs{
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
	group := &tc.TaskGroupInfo{Tasks: make([]tc.TaskInfo, 3)}
	group.Tasks[0].Name = "wpt-chrome-dev-results"
	group.Tasks[1].Name = "wpt-chrome-dev-stability" // must be skipped
	group.Tasks[2].Name = "wpt-chrome-dev-results-without-changes"
	for i := 0; i < len(group.Tasks); i++ {
		group.Tasks[i].State = "completed"
		group.Tasks[i].TaskID = fmt.Sprint(i)
	}

	t.Run("All", func(t *testing.T) {
		urls, err := tc.ExtractArtifactURLs("https://tc.example.com", shared.NewNilLogger(), group, "")
		assert.Nil(t, err)
		assert.Equal(t, map[string]tc.ArtifactURLs{
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
		urls, err := tc.ExtractArtifactURLs("https://tc.example.com", shared.NewNilLogger(), group, "2")
		assert.Nil(t, err)
		assert.Equal(t, map[string]tc.ArtifactURLs{
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
	group := &tc.TaskGroupInfo{Tasks: make([]tc.TaskInfo, 3)}
	group.Tasks[0].State = "failed"
	group.Tasks[0].TaskID = "foo"
	group.Tasks[0].Name = "wpt-firefox-nightly-testharness-1"
	group.Tasks[1].State = "completed"
	group.Tasks[1].TaskID = "bar"
	group.Tasks[1].Name = "wpt-firefox-nightly-testharness-2"
	group.Tasks[2].State = "completed"
	group.Tasks[2].TaskID = "baz"
	group.Tasks[2].Name = "wpt-chrome-dev-testharness-1"

	urls, err := tc.ExtractArtifactURLs("https://tc.example.com", shared.NewNilLogger(), group, "")
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
	aeAPI.EXPECT().GetHTTPClientWithTimeout(uc.UploadTimeout).AnyTimes().Return(server.Client())
	aeAPI.EXPECT().GetResultsUploadURL().AnyTimes().Return(serverURL)

	t.Run("master", func(t *testing.T) {
		err := tc.CreateAllRuns(
			shared.NewNilLogger(),
			aeAPI,
			sha,
			"username",
			"password",
			map[string]tc.ArtifactURLs{
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
		err := tc.CreateAllRuns(
			shared.NewNilLogger(),
			aeAPI,
			sha,
			"username",
			"password",
			map[string]tc.ArtifactURLs{
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
	aeAPI.EXPECT().GetHTTPClientWithTimeout(uc.UploadTimeout).Times(2).Return(server.Client())
	serverURL, _ := url.Parse(server.URL)
	aeAPI.EXPECT().GetResultsUploadURL().AnyTimes().Return(serverURL)

	err := tc.CreateAllRuns(
		shared.NewNilLogger(),
		aeAPI,
		sha,
		"username",
		"password",
		map[string]tc.ArtifactURLs{
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
	aeAPI.EXPECT().GetHTTPClientWithTimeout(uc.UploadTimeout).MinTimes(1).Return(&http.Client{Timeout: time.Microsecond})
	serverURL, _ := url.Parse(server.URL)
	aeAPI.EXPECT().GetResultsUploadURL().AnyTimes().Return(serverURL)

	err := tc.CreateAllRuns(
		shared.NewNilLogger(),
		aeAPI,
		sha,
		"username",
		"password",
		map[string]tc.ArtifactURLs{
			"chrome":  {Results: []string{"1"}},
			"firefox": {Results: []string{"1", "2"}},
		},
		[]string{shared.MasterLabel},
	)
	assert.NotNil(t, err)
	assert.Equal(t, 2, strings.Count(err.Error(), "Client.Timeout"))
}

func TestCreateAllRuns_pr_labels_exclude_master(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// We should not see a master label here, even though we
		// specify one in the call to tc.CreateAllRuns.
		defer r.Body.Close()
		body, _ := io.ReadAll(r.Body)
		assert.NotContains(t, string(body), "master")
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
	aeAPI.EXPECT().GetHTTPClientWithTimeout(uc.UploadTimeout).AnyTimes().Return(server.Client())
	aeAPI.EXPECT().GetResultsUploadURL().AnyTimes().Return(serverURL)

	// This test reproduces the case where Community-TC executes a pull
	// request run on a master commit (which we have historically seen).
	// When we get a master-tagged run which contains pull-request runs, we
	// should ignore the tag. This is asserted by the HTTP handler above.
	err := tc.CreateAllRuns(
		shared.NewNilLogger(),
		aeAPI,
		sha,
		"username",
		"password",
		map[string]tc.ArtifactURLs{
			"chrome-dev-pr_head":     {Results: []string{"1"}},
			"chrome-dev-pr_base":     {Results: []string{"1"}},
			"firefox-stable-pr_head": {Results: []string{"1"}},
			"firefox-stable-pr_base": {Results: []string{"1"}},
		},
		[]string{shared.MasterLabel, "user:person"},
	)
	assert.Nil(t, err)
}

func TestTaskNameRegex(t *testing.T) {
	assert.Equal(t, []string{"chrome-dev", "results"}, tc.TaskNameRegex.FindStringSubmatch("wpt-chrome-dev-results")[1:])
	assert.Equal(t, []string{"chrome-dev", "results-without-changes"}, tc.TaskNameRegex.FindStringSubmatch("wpt-chrome-dev-results-without-changes")[1:])
	assert.Equal(t, []string{"chrome-dev", "stability"}, tc.TaskNameRegex.FindStringSubmatch("wpt-chrome-dev-stability")[1:])
	assert.Equal(t, []string{"chrome-stable", "reftest"}, tc.TaskNameRegex.FindStringSubmatch("wpt-chrome-stable-reftest-1")[1:])
	assert.Equal(t, []string{"firefox-beta", "crashtest"}, tc.TaskNameRegex.FindStringSubmatch("wpt-firefox-beta-crashtest-2")[1:])
	assert.Equal(t, []string{"firefox-nightly", "testharness"}, tc.TaskNameRegex.FindStringSubmatch("wpt-firefox-nightly-testharness-5")[1:])
	assert.Equal(t, []string{"firefox-stable", "wdspec"}, tc.TaskNameRegex.FindStringSubmatch("wpt-firefox-stable-wdspec-1")[1:])
	assert.Equal(t, []string{"webkitgtk_minibrowser-nightly", "testharness"}, tc.TaskNameRegex.FindStringSubmatch("wpt-webkitgtk_minibrowser-nightly-testharness-2")[1:])
	assert.Nil(t, tc.TaskNameRegex.FindStringSubmatch("wpt-foo-bar--1"))
	assert.Nil(t, tc.TaskNameRegex.FindStringSubmatch("wpt-foo-bar-"))
}

func TestGetStatusEventInfo_target_url(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)
	api.EXPECT().GetTaskGroupInfo("https://tc.community.com", "IWlO7NuxRnO0_8PKMuHFkw").Return(nil, nil)

	status := tc.StatusEventPayload{}
	status.State = strPtr("success")
	status.TargetURL = strPtr("https://tc.community.com/tasks/groups/IWlO7NuxRnO0_8PKMuHFkw/tasks/123")
	status.Context = strPtr("Community-TC")
	status.Branches = branchInfos{&github.Branch{Name: strPtr(shared.MasterLabel)}}
	status.SHA = strPtr("abcdef123")

	// The target URL must be present, and must at least be a recognized
	// URL containing a taskGroupID. ParseTaskclusterURL is tested
	// separately, so just do a basic check here.
	event, err := tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.Equal(t, event.RootURL, "https://tc.community.com")
	assert.Equal(t, event.TaskID, "123")
	assert.Nil(t, err)

	status.TargetURL = strPtr("https://example.com/nope/not/right")
	event, err = tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.NotNil(t, err)

	status.TargetURL = nil
	event, err = tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.NotNil(t, err)
}

func TestGetStatusEventInfo_sha(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)
	api.EXPECT().GetTaskGroupInfo(gomock.Any(), gomock.Any()).Return(nil, nil)

	status := tc.StatusEventPayload{}
	status.State = strPtr("success")
	status.TargetURL = strPtr("https://tc.community.com/tasks/groups/IWlO7NuxRnO0_8PKMuHFkw/tasks/123")
	status.Context = strPtr("Community-TC")
	status.Branches = branchInfos{&github.Branch{Name: strPtr(shared.MasterLabel)}}
	status.SHA = strPtr("abcdef123")

	// We don't place requirements on the SHA other than it exists.
	event, err := tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.Equal(t, event.Sha, "abcdef123")
	assert.Nil(t, err)

	status.SHA = nil
	event, err = tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.NotNil(t, err)
}

func TestGetStatusEventInfo_master(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)
	api.EXPECT().GetTaskGroupInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	status := tc.StatusEventPayload{}
	status.State = strPtr("success")
	status.TargetURL = strPtr("https://tc.community.com/tasks/groups/IWlO7NuxRnO0_8PKMuHFkw/tasks/123")
	status.Context = strPtr("Community-TC")
	status.Branches = branchInfos{&github.Branch{Name: strPtr("mybranch")}, &github.Branch{Name: strPtr(shared.MasterLabel)}}
	status.SHA = strPtr("abcdef123")

	// We check whether an event is for master by looking at the branches
	// it is associated with.
	event, err := tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.Equal(t, event.Master, true)
	assert.Nil(t, err)

	status.Branches = branchInfos{&github.Branch{Name: strPtr("mybranch")}}
	event, err = tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.Equal(t, event.Master, false)
	assert.Nil(t, err)

	// Missing the 'branches' entry is not an error; the event just isn't
	// for master.
	status.Branches = nil
	event, err = tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.Equal(t, event.Master, false)
	assert.Nil(t, err)
}

func TestGetStatusEventInfo_sender(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)
	api.EXPECT().GetTaskGroupInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	status := tc.StatusEventPayload{}
	status.State = strPtr("success")
	status.TargetURL = strPtr("https://tc.community.com/tasks/groups/IWlO7NuxRnO0_8PKMuHFkw/tasks/123")
	status.Context = strPtr("Community-TC")
	status.Branches = branchInfos{&github.Branch{Name: strPtr(shared.MasterLabel)}}
	status.SHA = strPtr("abcdef123")

	// The sender is entirely optional.
	status.Commit = &github.RepositoryCommit{Author: &github.User{Login: strPtr("someuser")}}
	event, err := tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.Equal(t, event.Sender, "someuser")
	assert.Nil(t, err)

	status.Commit = nil
	event, err = tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.Equal(t, event.Sender, "")
	assert.Nil(t, err)
}

func TestGetStatusEventInfo_group(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)
	group := &tc.TaskGroupInfo{Tasks: make([]tc.TaskInfo, 0)}

	status := tc.StatusEventPayload{}
	status.State = strPtr("success")
	status.TargetURL = strPtr("https://tc.community.com/tasks/groups/IWlO7NuxRnO0_8PKMuHFkw/tasks/123")
	status.Context = strPtr("Community-TC")
	status.Branches = branchInfos{&github.Branch{Name: strPtr(shared.MasterLabel)}}
	status.SHA = strPtr("abcdef123")

	api.EXPECT().GetTaskGroupInfo(gomock.Any(), gomock.Any()).Return(group, nil).Times(1)
	event, err := tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.Equal(t, event.Group, group)
	assert.Nil(t, err)

	api.EXPECT().GetTaskGroupInfo(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed")).Times(1)
	event, err = tc.GetStatusEventInfo(status, shared.NewNilLogger(), api)
	assert.NotNil(t, err)
}

func TestGetCheckSuiteEventInfo_sourceRepo(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)

	runs := []*github.CheckRun{
		{
			Name:       strPtr("wpt-decision-task"),
			Status:     strPtr("completed"),
			DetailsURL: strPtr("https://community-tc.services.mozilla.com/tasks/Jq4HzLz0R2eKkJFdmf47Bg"),
		},
	}
	api.EXPECT().ListCheckRuns(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(runs, nil)

	event := github.CheckSuiteEvent{
		CheckSuite: &github.CheckSuite{
			HeadSHA: strPtr("abcdef123"),
		},
		Repo: &github.Repository{
			Owner: &github.User{
				Login: strPtr("web-platform-tests"),
			},
			Name: strPtr("wpt"),
		},
	}

	// Valid owner and name.
	_, err := tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.Nil(t, err)

	// Invalid name.
	event.Repo.Name = strPtr("not-wpt")
	_, err = tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.NotNil(t, err)

	// Invalid owner.
	event.Repo.Name = strPtr("wpt")
	event.Repo.Owner.Login = strPtr("stephenmcgruer")
	_, err = tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.NotNil(t, err)
}

func TestGetCheckSuiteEventInfo_sha(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)

	runs := []*github.CheckRun{
		{
			Name:       strPtr("wpt-decision-task"),
			Status:     strPtr("completed"),
			DetailsURL: strPtr("https://community-tc.services.mozilla.com/tasks/Jq4HzLz0R2eKkJFdmf47Bg"),
		},
	}
	api.EXPECT().ListCheckRuns(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(runs, nil)

	event := github.CheckSuiteEvent{
		CheckSuite: &github.CheckSuite{
			HeadSHA: strPtr("abcdef123"),
		},
		Repo: &github.Repository{
			Owner: &github.User{
				Login: strPtr("web-platform-tests"),
			},
			Name: strPtr("wpt"),
		},
	}

	// We don't place requirements on the SHA other than it exists.
	eventInfo, err := tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.Equal(t, "abcdef123", eventInfo.Sha)
	assert.Nil(t, err)

	event.CheckSuite.HeadSHA = nil
	eventInfo, err = tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.NotNil(t, err)
}

func TestGetCheckSuiteEventInfo_master(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)

	runs := []*github.CheckRun{
		{
			Name:       strPtr("wpt-decision-task"),
			Status:     strPtr("completed"),
			DetailsURL: strPtr("https://community-tc.services.mozilla.com/tasks/Jq4HzLz0R2eKkJFdmf47Bg"),
		},
	}
	api.EXPECT().ListCheckRuns(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(runs, nil)

	event := github.CheckSuiteEvent{
		CheckSuite: &github.CheckSuite{
			HeadBranch: strPtr("master"),
			HeadSHA:    strPtr("abcdef123"),
		},
		Repo: &github.Repository{
			Owner: &github.User{
				Login: strPtr("web-platform-tests"),
			},
			Name: strPtr("wpt"),
		},
	}

	eventInfo, err := tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.Equal(t, true, eventInfo.Master)
	assert.Nil(t, err)

	event.CheckSuite.HeadBranch = strPtr("my-branch")
	eventInfo, err = tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.Equal(t, false, eventInfo.Master)
	assert.Nil(t, err)
}

func TestGetCheckSuiteEventInfo_sender(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)

	runs := []*github.CheckRun{
		{
			Name:       strPtr("wpt-decision-task"),
			Status:     strPtr("completed"),
			DetailsURL: strPtr("https://community-tc.services.mozilla.com/tasks/Jq4HzLz0R2eKkJFdmf47Bg"),
		},
	}
	api.EXPECT().ListCheckRuns(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(runs, nil)

	event := github.CheckSuiteEvent{
		Sender: &github.User{
			Login: strPtr("myuser"),
		},
		CheckSuite: &github.CheckSuite{
			HeadSHA: strPtr("abcdef123"),
		},
		Repo: &github.Repository{
			Owner: &github.User{
				Login: strPtr("web-platform-tests"),
			},
			Name: strPtr("wpt"),
		},
	}

	eventInfo, err := tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.Equal(t, "myuser", eventInfo.Sender)
	assert.Nil(t, err)

	event.Sender.Login = nil
	eventInfo, err = tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.Equal(t, "", eventInfo.Sender)
	assert.Nil(t, err)
}

func TestGetCheckSuiteEventInfo_checkRuns(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)

	// The list of check_run events give us two main pieces of information:
	//
	//	the RootURL, which must match across runs, and
	//	the TaskGroupInfo:
	//		TaskGroupID is the wpt-decision-tasks's taskID
	//		Tasks is filled with each check_run's name, taskID, and status.
	runs := []*github.CheckRun{
		{
			Name:       strPtr("wpt-decision-task"),
			Status:     strPtr("completed"),
			Conclusion: strPtr("success"),
			DetailsURL: strPtr("https://community-tc.services.mozilla.com/tasks/Jq4HzLz0R2eKkJFdmf47Bg"),
		},
		{
			Name:       strPtr("wpt-chrome-dev-testharness-1"),
			Status:     strPtr("completed"),
			Conclusion: strPtr("failed"),
			DetailsURL: strPtr("https://community-tc.services.mozilla.com/tasks/IWlO7NuxRnO0_8PKMuHFkw"),
		},
	}
	api.EXPECT().ListCheckRuns(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(runs, nil)

	event := github.CheckSuiteEvent{
		CheckSuite: &github.CheckSuite{
			HeadSHA: strPtr("abcdef123"),
		},
		Repo: &github.Repository{
			Owner: &github.User{
				Login: strPtr("web-platform-tests"),
			},
			Name: strPtr("wpt"),
		},
	}

	eventInfo, err := tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.Equal(t, "https://community-tc.services.mozilla.com", eventInfo.RootURL)
	assert.Equal(t, "Jq4HzLz0R2eKkJFdmf47Bg", eventInfo.Group.TaskGroupID)
	assert.Equal(t, "wpt-decision-task", eventInfo.Group.Tasks[0].Name)
	assert.Equal(t, "Jq4HzLz0R2eKkJFdmf47Bg", eventInfo.Group.Tasks[0].TaskID)
	assert.Equal(t, "completed", eventInfo.Group.Tasks[0].State)
	assert.Equal(t, "wpt-chrome-dev-testharness-1", eventInfo.Group.Tasks[1].Name)
	assert.Equal(t, "IWlO7NuxRnO0_8PKMuHFkw", eventInfo.Group.Tasks[1].TaskID)
	assert.Equal(t, "failed", eventInfo.Group.Tasks[1].State)
	assert.Nil(t, err)

	// Check the case where a details URL will fail to parse.
	runs[0].DetailsURL = strPtr("https://example.com/nope/not/right")
	eventInfo, err = tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.NotNil(t, err)

	// Check the case where a details URL is missing.
	runs[0].DetailsURL = nil
	eventInfo, err = tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.NotNil(t, err)

	// Check the case where a details URL has a mismatching root URL.
	runs[0].DetailsURL = strPtr("https://tc.community.com/tasks/Jq4HzLz0R2eKkJFdmf47Bg")
	eventInfo, err = tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.NotNil(t, err)
}

func TestGetCheckSuiteEventInfo_checkRunsEmpty(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)
	api.EXPECT().ListCheckRuns(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]*github.CheckRun{}, nil)

	event := github.CheckSuiteEvent{
		CheckSuite: &github.CheckSuite{
			HeadSHA: strPtr("abcdef123"),
		},
		Repo: &github.Repository{
			Owner: &github.User{
				Login: strPtr("web-platform-tests"),
			},
			Name: strPtr("wpt"),
		},
	}

	_, err := tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.NotNil(t, err)
}

func TestGetCheckSuiteEventInfo_checkRunsFailed(t *testing.T) {
	mockC := gomock.NewController(t)
	defer mockC.Finish()
	api := mock_tc.NewMockAPI(mockC)
	api.EXPECT().ListCheckRuns(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.New("failed"))

	event := github.CheckSuiteEvent{
		CheckSuite: &github.CheckSuite{
			HeadSHA: strPtr("abcdef123"),
		},
		Repo: &github.Repository{
			Owner: &github.User{
				Login: strPtr("web-platform-tests"),
			},
			Name: strPtr("wpt"),
		},
	}

	_, err := tc.GetCheckSuiteEventInfo(event, shared.NewNilLogger(), api)
	assert.NotNil(t, err)
}
