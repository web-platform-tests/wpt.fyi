// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webhook

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string {
	return &s
}

func TestShouldProcessStatus_ok(t *testing.T) {
	status := statusEventPayload{}
	status.State = strPtr("success")
	status.Context = strPtr("Taskcluster")
	status.Branches = branchInfos{&github.Branch{Name: strPtr("master")}}
	assert.True(t, shouldProcessStatus(&status))
}

func TestShouldProcessStatus_unsuccessful(t *testing.T) {
	status := statusEventPayload{}
	status.State = strPtr("error")
	status.Context = strPtr("Taskcluster")
	status.Branches = branchInfos{&github.Branch{Name: strPtr("master")}}
	assert.False(t, shouldProcessStatus(&status))
}

func TestShouldProcessStatus_notTaskcluster(t *testing.T) {
	status := statusEventPayload{}
	status.State = strPtr("success")
	status.Context = strPtr("Travis")
	status.Branches = branchInfos{&github.Branch{Name: strPtr("master")}}
	assert.False(t, shouldProcessStatus(&status))
}

func TestShouldProcessStatus_notOnMaster(t *testing.T) {
	status := statusEventPayload{}
	status.State = strPtr("success")
	status.Context = strPtr("Taskcluster")
	status.Branches = branchInfos{&github.Branch{Name: strPtr("gh-pages")}}
	assert.False(t, shouldProcessStatus(&status))
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

func TestExtractResultURLs(t *testing.T) {
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

	urls, err := extractResultURLs(group)
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

func TestVerifySignature(t *testing.T) {
	message := []byte("test")
	signature := "a053ee211b4693456ef071e336f74ab699250318"
	secret := "95bfab9afa719185ee7c3658356b166b7f45349a"
	assert.True(t, verifySignature(
		message, signature, secret))
	assert.False(t, verifySignature(
		[]byte("foobar"), signature, secret))
	assert.False(t, verifySignature(
		message, "875a5feef4cde4265d6d5d21c304d755903ccb60", secret))
	assert.False(t, verifySignature(
		message, signature, "875a5feef4cde4265d6d5d21c304d755903ccb60"))
	// Test an ill-formed (odd-length) signature.
	assert.False(t, verifySignature(
		message, "875a5feef4cde4265d6d5d21c304d755903ccb6", secret))
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

	err := createAllRuns(logrus.New(),
		&http.Client{},
		server.URL,
		"username",
		"password",
		map[string][]string{"chrome": []string{"1"}, "firefox": []string{"1", "2"}},
		[]string{"master"},
	)
	assert.Nil(t, err)
	assert.Equal(t, uint32(2), requested)
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

	err := createAllRuns(logrus.New(),
		&http.Client{},
		server.URL,
		"username",
		"password",
		map[string][]string{"chrome": []string{"1"}, "firefox": []string{"1", "2"}},
		[]string{"master"},
	)
	assert.NotNil(t, err)
	assert.Equal(t, uint32(2), requested)
	assert.Contains(t, err.Error(), "API error: Not found")
}

func TestCreateAllRuns_all_errors(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 2)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	err := createAllRuns(logrus.New(),
		&http.Client{Timeout: time.Second},
		server.URL,
		"username",
		"password",
		map[string][]string{"chrome": []string{"1"}, "firefox": []string{"1", "2"}},
		[]string{"master"},
	)
	assert.NotNil(t, err)
	assert.Equal(t, 2, strings.Count(err.Error(), "Client.Timeout"))
}
