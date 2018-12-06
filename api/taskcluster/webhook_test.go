// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package taskcluster

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/checks"
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
	status.Branches = branchInfos{&github.Branch{Name: strPtr("master")}}
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
	status.Branches = branchInfos{&github.Branch{Name: strPtr("master")}}
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
			Name:   strPtr("master"),
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
	assert.Equal(t, []string{"master"}, status.HeadingBranches().GetNames())
	assert.True(t, status.IsOnMaster())

	status.Branches = status.Branches[1:]
	assert.False(t, status.IsOnMaster())
}

func TestExtractTaskGroupID(t *testing.T) {
	assert.Equal(t, "Y4rnZeqDRXGiRNiqxT5Qeg",
		extractTaskGroupID("https://tools.taskcluster.net/task-group-inspector/#/Y4rnZeqDRXGiRNiqxT5Qeg"))
}

func TestExtractResultURLs_all_success(t *testing.T) {
	group := &taskGroupInfo{Tasks: make([]taskInfo, 3)}
	group.Tasks[0].Status.State = "completed"
	group.Tasks[0].Status.TaskID = "foo"
	group.Tasks[0].Task.Metadata.Name = "wpt-firefox-nightly-testharness-1"
	group.Tasks[1].Status.State = "completed"
	group.Tasks[1].Status.TaskID = "bar"
	group.Tasks[1].Task.Metadata.Name = "wpt-firefox-nightly-testharness-2"
	group.Tasks[2].Status.State = "completed"
	group.Tasks[2].Status.TaskID = "baz"
	group.Tasks[2].Task.Metadata.Name = "wpt-chrome-dev-testharness-1"

	urls, err := extractResultURLs(shared.NewNilLogger(), group)
	assert.Nil(t, err)
	assert.Equal(t, map[string][]string{
		"firefox-nightly": {
			"https://queue.taskcluster.net/v1/task/foo/artifacts/public/results/wpt_report.json.gz",
			"https://queue.taskcluster.net/v1/task/bar/artifacts/public/results/wpt_report.json.gz",
		},
		"chrome-dev": {
			"https://queue.taskcluster.net/v1/task/baz/artifacts/public/results/wpt_report.json.gz",
		},
	}, urls)
}

func TestExtractResultURLs_with_failures(t *testing.T) {
	group := &taskGroupInfo{Tasks: make([]taskInfo, 3)}
	group.Tasks[0].Status.State = "failed"
	group.Tasks[0].Status.TaskID = "foo"
	group.Tasks[0].Task.Metadata.Name = "wpt-firefox-nightly-testharness-1"
	group.Tasks[1].Status.State = "completed"
	group.Tasks[1].Status.TaskID = "bar"
	group.Tasks[1].Task.Metadata.Name = "wpt-firefox-nightly-testharness-2"
	group.Tasks[2].Status.State = "completed"
	group.Tasks[2].Status.TaskID = "baz"
	group.Tasks[2].Task.Metadata.Name = "wpt-chrome-dev-testharness-1"

	urls, err := extractResultURLs(shared.NewNilLogger(), group)
	assert.Nil(t, err)
	assert.Equal(t, map[string][]string{
		"chrome-dev": {
			"https://queue.taskcluster.net/v1/task/baz/artifacts/public/results/wpt_report.json.gz",
		},
	}, urls)
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
	mockC := gomock.NewController(t)
	defer mockC.Finish()

	sha := "abcdef1234abcdef1234abcdef1234abcdef1234"

	checksAPI := checks.NewMockAPI(mockC)
	suite := shared.CheckSuite{SHA: sha}
	checksAPI.EXPECT().GetSuitesForSHA(sha).Return([]shared.CheckSuite{suite}, nil)
	checksAPI.EXPECT().PendingCheckRun(suite, sharedtest.SameProductSpec("safari[experimental]"))
	checksAPI.EXPECT().PendingCheckRun(suite, sharedtest.SameProductSpec("chrome[experimental]"))
	checksAPI.EXPECT().PendingCheckRun(suite, sharedtest.SameProductSpec("firefox"))
	aeAPI := sharedtest.NewMockAppEngineAPI(mockC)
	aeAPI.EXPECT().GetHostname().MinTimes(1).Return("localhost:8080")

	err := createAllRuns(
		logrus.New(),
		&http.Client{},
		aeAPI,
		checksAPI,
		server.URL,
		sha,
		"username",
		"password",
		map[string][]string{
			"safari-preview": []string{"1"},
			"chrome-dev":     []string{"1"},
			"firefox":        []string{"1", "2"},
		},
		[]string{"master"},
	)
	assert.Nil(t, err)
	assert.Equal(t, uint32(3), requested)
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
			panic("requested != 0 && requested != 1")
		}
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()
	mockC := gomock.NewController(t)
	defer mockC.Finish()

	sha := "abcdef1234abcdef1234abcdef1234abcdef1234"

	checksAPI := checks.NewMockAPI(mockC)
	suite := shared.CheckSuite{SHA: sha}
	checksAPI.EXPECT().GetSuitesForSHA(sha).Return([]shared.CheckSuite{suite}, nil)
	checksAPI.EXPECT().PendingCheckRun(suite, gomock.Any())
	aeAPI := sharedtest.NewMockAppEngineAPI(mockC)
	aeAPI.EXPECT().GetHostname().MinTimes(1).Return("localhost:8080")

	err := createAllRuns(
		logrus.New(),
		&http.Client{},
		aeAPI,
		checksAPI,
		server.URL,
		sha,
		"username",
		"password",
		map[string][]string{"chrome": []string{"1"}, "firefox": []string{"1", "2"}},
		[]string{"master"},
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

	checksAPI := checks.NewMockAPI(mockC)
	suite := shared.CheckSuite{SHA: sha}
	checksAPI.EXPECT().GetSuitesForSHA(sha).Return([]shared.CheckSuite{suite}, nil)
	aeAPI := sharedtest.NewMockAppEngineAPI(mockC)
	aeAPI.EXPECT().GetHostname().MinTimes(1).Return("localhost:8080")

	err := createAllRuns(
		logrus.New(),
		&http.Client{Timeout: time.Second},
		aeAPI,
		checksAPI,
		server.URL,
		sha,
		"username",
		"password",
		map[string][]string{"chrome": []string{"1"}, "firefox": []string{"1", "2"}},
		[]string{"master"},
	)
	assert.NotNil(t, err)
	assert.Equal(t, 2, strings.Count(err.Error(), "Client.Timeout"))
}

func TestTaskNameRegex(t *testing.T) {
	assert.Len(t, taskNameRegex.FindStringSubmatch("wpt-chrome-dev-results"), 4)
	assert.Len(t, taskNameRegex.FindStringSubmatch("wpt-chrome-dev-reftest-1"), 4)
	assert.Len(t, taskNameRegex.FindStringSubmatch("wpt-chrome-dev-testharness-5"), 4)
	assert.Len(t, taskNameRegex.FindStringSubmatch("wpt-chrome-dev-wdspec-1"), 4)
}
