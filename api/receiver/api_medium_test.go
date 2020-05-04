// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/taskqueue"
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
	ctx, done, err := sharedtest.NewAEContext(false)
	assert.Nil(t, err)
	defer done()

	stats, err := taskqueue.QueueStats(ctx, []string{""})
	assert.Nil(t, err)
	assert.Equal(t, stats[0].Tasks, 0)

	a := NewAPI(ctx).(*apiImpl)
	// dev_appserver does not support non-default queues, so we override the name of the queue here.
	a.queue = ""
	results := []string{"gs://blade-runner/test.json", "http://wpt.fyi/test.json.gz"}
	screenshots := []string{"gs://blade-runner/test.db"}
	task, err := a.ScheduleResultsTask("blade-runner", results, screenshots, nil)
	assert.Nil(t, err)

	payload, err := url.ParseQuery(string(task.Payload))
	assert.Nil(t, err)
	assert.Equal(t, task.Name, payload.Get("id"))
	assert.Equal(t, "blade-runner", payload.Get("uploader"))
	assert.Equal(t, results, payload["results"])
	assert.Equal(t, screenshots, payload["screenshots"])

	stats, err = taskqueue.QueueStats(ctx, []string{""})
	assert.Nil(t, err)
	assert.Equal(t, stats[0].Tasks, 1)

	var pendingRun shared.PendingTestRun
	id, err := strconv.Atoi(task.Name)
	assert.Nil(t, err)
	datastore.Get(ctx, datastore.NewKey(ctx, "PendingTestRun", "", int64(id), nil), &pendingRun)
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
	datastore.Get(ctx, datastore.NewKey(ctx, key.Kind(), "", key.IntID(), nil), &testRun2)
	testRun2.ID = key.IntID()
	assert.Equal(t, testRun, testRun2)
}

func TestUpdatePendingTestRun(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := NewAPI(ctx)

	sha := "0123456789012345678901234567890123456789"
	key := datastore.NewKey(ctx, "PendingTestRun", "", 1, nil)
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
	datastore.Get(ctx, key, &run2)
	assert.Equal(t, shared.StageWptFyiReceived, run2.Stage)
	assert.Equal(t, sha, run2.FullRevisionHash)
	assert.Equal(t, sha[:10], run2.Revision)

	// CheckRunID should not be updated; Stage should be transitioned.
	run.CheckRunID = 0
	run.Stage = shared.StageValid
	assert.Nil(t, a.UpdatePendingTestRun(run))
	var run3 shared.PendingTestRun
	datastore.Get(ctx, key, &run3)
	assert.Equal(t, int64(100), run3.CheckRunID)
	assert.Equal(t, shared.StageValid, run3.Stage)
	assert.Equal(t, run2.Created, run3.Created)

	// Stage cannot be transitioned backwards.
	run.Stage = shared.StageWptFyiProcessing
	assert.EqualError(t, a.UpdatePendingTestRun(run),
		"cannot transition from VALID to WPTFYI_PROCESSING")
}

func TestAuthenticateUploader(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := NewAPI(ctx)

	req := httptest.NewRequest("", "/api/foo", &bytes.Buffer{})
	assert.Equal(t, "", AuthenticateUploader(a, req))

	req.SetBasicAuth(InternalUsername, "123")
	assert.Equal(t, "", AuthenticateUploader(a, req))

	key := datastore.NewKey(ctx, "Uploader", InternalUsername, 0, nil)
	datastore.Put(ctx, key, &shared.Uploader{Username: InternalUsername, Password: "123"})
	assert.Equal(t, InternalUsername, AuthenticateUploader(a, req))

	req.SetBasicAuth(InternalUsername, "456")
	assert.Equal(t, "", AuthenticateUploader(a, req))
}
