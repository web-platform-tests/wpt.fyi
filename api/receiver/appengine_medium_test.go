// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"bytes"
	"fmt"
	"io"
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
	a := appEngineAPIImpl{}
	mGcs := mockGcs{}
	a.gcs = &mGcs

	err := a.uploadToGCS("/test_bucket/path/to/test.json", strings.NewReader("test content"), false)
	assert.Nil(t, err)
	assert.Equal(t, "test_bucket", mGcs.mockWriter.bucketName)
	assert.Equal(t, "path/to/test.json", mGcs.mockWriter.fileName)
	assert.Equal(t, "test content", string(mGcs.mockWriter.finalContent))
}

func TestUploadToGCS_handlesErrors(t *testing.T) {
	errNew := fmt.Errorf("error creating writer")
	a := appEngineAPIImpl{}
	a.gcs = &mockGcs{errOnNew: errNew}
	err := a.uploadToGCS("/bucket/test.json", strings.NewReader(""), false)
	assert.Equal(t, errNew, err)

	errClose := fmt.Errorf("error closing writer")
	a.gcs = &mockGcs{mockWriter: mockGcsWriter{errOnClose: errClose}}
	err = a.uploadToGCS("/bucket/test.json", strings.NewReader(""), false)
	assert.Equal(t, errClose, err)

	a.gcs = &mockGcs{}
	err = a.uploadToGCS("test.json", strings.NewReader(""), false)
	assert.EqualError(t, err, "invalid GCS path: test.json")
}

func TestScheduleResultsTask(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(false)
	assert.Nil(t, err)
	defer done()

	stats, err := taskqueue.QueueStats(ctx, []string{""})
	assert.Nil(t, err)
	assert.Equal(t, stats[0].Tasks, 0)

	a := appEngineAPIImpl{ctx: ctx}
	_, err = a.scheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "single", nil)
	assert.Nil(t, err)

	stats, err = taskqueue.QueueStats(ctx, []string{""})
	assert.Nil(t, err)
	assert.Equal(t, stats[0].Tasks, 1)
}

func TestScheduleResultsTask_error(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(false)
	assert.Nil(t, err)
	defer done()
	a := appEngineAPIImpl{ctx: ctx}

	_, err = a.scheduleResultsTask("", []string{"/blade-runner/test.json"}, "single", nil)
	assert.NotNil(t, err)

	_, err = a.scheduleResultsTask("blade-runner", []string{""}, "single", nil)
	assert.NotNil(t, err)

	_, err = a.scheduleResultsTask("blade-runner", nil, "single", nil)
	assert.NotNil(t, err)

	_, err = a.scheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "", nil)
	assert.NotNil(t, err)
}

func TestAddTestRun(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := appEngineAPIImpl{ctx: ctx}

	testRun := shared.TestRun{
		ProductAtRevision: shared.ProductAtRevision{
			Revision: "0123456789",
		},
	}

	key, err := a.addTestRun(&testRun)
	assert.Nil(t, err)
	assert.Equal(t, "TestRun", key.Kind)

	var testRun2 shared.TestRun
	datastore.Get(ctx, datastore.NewKey(ctx, key.Kind, "", key.ID, nil), &testRun2)
	assert.Equal(t, testRun, testRun2)
}

func TestAuthenticateUploader(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := appEngineAPIImpl{ctx: ctx}

	assert.False(t, a.authenticateUploader("user", "123"))

	key := datastore.NewKey(ctx, "Uploader", "user", 0, nil)
	datastore.Put(ctx, key, &shared.Uploader{Username: "user", Password: "123"})
	assert.True(t, a.authenticateUploader("user", "123"))
}
