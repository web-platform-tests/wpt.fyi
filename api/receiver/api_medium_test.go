//go:build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

type mockGcs struct {
	mockWriter mockGcsWriter
	errOnNew   error
}

// mockGcsWriter implements io.WriteCloser
type mockGcsWriter struct {
	bytes.Buffer
	bucketName   string
	fileName     string
	finalContent []byte
	errOnClose   error
}

func (m *mockGcsWriter) Close() error {
	m.finalContent = m.Bytes()
	return m.errOnClose
}

func (m *mockGcs) NewWriter(bucketName, fileName, contentType, contentEncoding string) (io.WriteCloser, error) {
	m.mockWriter.bucketName = bucketName
	m.mockWriter.fileName = fileName
	return &m.mockWriter, m.errOnNew
}

func TestIsAdmin_failsToConstructACL(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := NewAPI(ctx).(*apiImpl)
	r := httptest.NewRequest("GET", "/api/results/upload", nil)

	// User not logged in
	a.githubACLFactory = func(_ *http.Request) (shared.GitHubAccessControl, error) {
		return nil, nil
	}
	assert.False(t, a.IsAdmin(r))

	a.githubACLFactory = func(_ *http.Request) (shared.GitHubAccessControl, error) {
		return nil, errors.New("error constructing ACL")
	}
	assert.False(t, a.IsAdmin(r))
}

func TestIsAdmin_mockACL(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := NewAPI(ctx).(*apiImpl)
	r := httptest.NewRequest("GET", "/api/results/upload", nil)

	t.Run("error", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockACL := sharedtest.NewMockGitHubAccessControl(mockCtrl)
		mockACL.EXPECT().IsValidAdmin().Return(true, errors.New("error checking admin"))
		a.githubACLFactory = func(_ *http.Request) (shared.GitHubAccessControl, error) {
			return mockACL, nil
		}
		assert.False(t, a.IsAdmin(r))
	})

	t.Run("admin", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockACL := sharedtest.NewMockGitHubAccessControl(mockCtrl)
		mockACL.EXPECT().IsValidAdmin().Return(true, nil)
		a.githubACLFactory = func(_ *http.Request) (shared.GitHubAccessControl, error) {
			return mockACL, nil
		}
		assert.True(t, a.IsAdmin(r))
	})

	t.Run("nonadmin", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockACL := sharedtest.NewMockGitHubAccessControl(mockCtrl)
		mockACL.EXPECT().IsValidAdmin().Return(false, nil)
		a.githubACLFactory = func(_ *http.Request) (shared.GitHubAccessControl, error) {
			return mockACL, nil
		}
		assert.False(t, a.IsAdmin(r))
	})
}

func TestUploadToGCS(t *testing.T) {
	ctx := context.Background()
	a := NewAPI(ctx).(*apiImpl)
	mGcs := mockGcs{}
	a.gcs = &mGcs

	err := a.UploadToGCS("gs://test_bucket/path/to/test.json", strings.NewReader("test content"), false)
	assert.Nil(t, err)
	assert.Equal(t, "test_bucket", mGcs.mockWriter.bucketName)
	assert.Equal(t, "path/to/test.json", mGcs.mockWriter.fileName)
	assert.Equal(t, "test content", string(mGcs.mockWriter.finalContent))
}

func TestUploadToGCS_handlesErrors(t *testing.T) {
	ctx := context.Background()
	a := NewAPI(ctx).(*apiImpl)

	errNew := fmt.Errorf("error creating writer")
	a.gcs = &mockGcs{errOnNew: errNew}
	err := a.UploadToGCS("gs://bucket/test.json", strings.NewReader(""), false)
	assert.Equal(t, errNew, err)

	errClose := fmt.Errorf("error closing writer")
	a.gcs = &mockGcs{mockWriter: mockGcsWriter{errOnClose: errClose}}
	err = a.UploadToGCS("gs://bucket/test.json", strings.NewReader(""), false)
	assert.Equal(t, errClose, err)

	a.gcs = &mockGcs{}
	err = a.UploadToGCS("/bucket/test.json", strings.NewReader(""), false)
	assert.EqualError(t, err, "invalid GCS path: /bucket/test.json")
}

func TestScheduleResultsTask(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAE := sharedtest.NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().Context().Return(ctx).AnyTimes()

	a := NewAPI(ctx).(*apiImpl)
	a.AppEngineAPI = mockAE
	results := []string{"gs://blade-runner/test.json", "http://wpt.fyi/test.json.gz"}
	screenshots := []string{"gs://blade-runner/test.db"}
	var id string
	mockAE.EXPECT().ScheduleTask(ResultsQueue, gomock.Any(), ResultsTarget, gomock.Any()).DoAndReturn(
		func(queueName, taskName, target string, params url.Values) (string, error) {
			assert.Equal(t, results, params["results"])
			assert.Equal(t, screenshots, params["screenshots"])
			assert.Equal(t, "blade-runner", params.Get("uploader"))
			assert.Equal(t, taskName, params.Get("id"))
			id = taskName
			return id, nil
		})
	task, err := a.ScheduleResultsTask("blade-runner", results, screenshots, nil, nil)
	assert.Equal(t, id, task)
	assert.Nil(t, err)

	intID, err := strconv.ParseInt(id, 10, 64)
	assert.Nil(t, err)
	var pendingRun shared.PendingTestRun
	store := shared.NewAppEngineDatastore(ctx, false)
	store.Get(store.NewIDKey("PendingTestRun", intID), &pendingRun)
	assert.Equal(t, "blade-runner", pendingRun.Uploader)
	assert.Equal(t, shared.StageWptFyiReceived, pendingRun.Stage)
}

func TestAddTestRun(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := NewAPI(ctx)

	testRun := shared.TestRun{
		ID: 123456,
		ProductAtRevision: shared.ProductAtRevision{
			Revision: "0123456789",
		},
	}

	key, err := a.AddTestRun(&testRun)
	assert.Nil(t, err)
	assert.Equal(t, "TestRun", key.Kind())
	assert.Equal(t, int64(123456), key.IntID())

	var testRun2 shared.TestRun
	store := shared.NewAppEngineDatastore(ctx, false)
	store.Get(key, &testRun2)
	testRun2.ID = key.IntID()
	assert.Equal(t, testRun, testRun2)
}

func TestUpdatePendingTestRun(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := NewAPI(ctx)

	sha := "0123456789012345678901234567890123456789"
	store := shared.NewAppEngineDatastore(ctx, false)
	key := store.NewIDKey("PendingTestRun", 1)
	run := shared.PendingTestRun{
		ID:         1,
		CheckRunID: 100,
		Stage:      shared.StageWptFyiReceived,
		ProductAtRevision: shared.ProductAtRevision{
			FullRevisionHash: sha,
		},
	}
	assert.Nil(t, a.UpdatePendingTestRun(run))
	var run2 shared.PendingTestRun
	store.Get(key, &run2)
	assert.Equal(t, shared.StageWptFyiReceived, run2.Stage)
	assert.Equal(t, sha, run2.FullRevisionHash)
	assert.Equal(t, sha[:10], run2.Revision)

	// CheckRunID should not be updated; Stage should be transitioned.
	run.CheckRunID = 0
	run.Stage = shared.StageValid
	assert.Nil(t, a.UpdatePendingTestRun(run))
	var run3 shared.PendingTestRun
	store.Get(key, &run3)
	assert.Equal(t, int64(100), run3.CheckRunID)
	assert.Equal(t, shared.StageValid, run3.Stage)
	assert.Equal(t, run2.Created, run3.Created)

	// Stage cannot be transitioned backwards.
	run.Stage = shared.StageWptFyiProcessing
	assert.EqualError(t, a.UpdatePendingTestRun(run),
		"cannot transition from VALID to WPTFYI_PROCESSING")
}
