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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

	err := a.UploadToGCS("/test_bucket/path/to/test.json", strings.NewReader("test content"), false)
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
	err := a.UploadToGCS("/bucket/test.json", strings.NewReader(""), false)
	assert.Equal(t, errNew, err)

	errClose := fmt.Errorf("error closing writer")
	a.gcs = &mockGcs{mockWriter: mockGcsWriter{errOnClose: errClose}}
	err = a.UploadToGCS("/bucket/test.json", strings.NewReader(""), false)
	assert.Equal(t, errClose, err)

	a.gcs = &mockGcs{}
	err = a.UploadToGCS("test.json", strings.NewReader(""), false)
	assert.EqualError(t, err, "invalid GCS path: test.json")
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
	_, err = a.ScheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "single", nil)
	assert.Nil(t, err)

	stats, err = taskqueue.QueueStats(ctx, []string{""})
	assert.Nil(t, err)
	assert.Equal(t, stats[0].Tasks, 1)
}

func TestScheduleResultsTask_error(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(false)
	assert.Nil(t, err)
	defer done()
	a := NewAPI(ctx)

	_, err = a.ScheduleResultsTask("", []string{"/blade-runner/test.json"}, "single", nil)
	assert.NotNil(t, err)

	_, err = a.ScheduleResultsTask("blade-runner", []string{""}, "single", nil)
	assert.NotNil(t, err)

	_, err = a.ScheduleResultsTask("blade-runner", nil, "single", nil)
	assert.NotNil(t, err)

	_, err = a.ScheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "", nil)
	assert.NotNil(t, err)
}

func TestAddTestRun(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := NewAPI(ctx)

	testRun := shared.TestRun{
		ProductAtRevision: shared.ProductAtRevision{
			Revision: "0123456789",
		},
	}

	key, err := a.AddTestRun(&testRun)
	assert.Nil(t, err)
	assert.Equal(t, "TestRun", key.Kind())

	var testRun2 shared.TestRun
	datastore.Get(ctx, datastore.NewKey(ctx, key.Kind(), "", key.IntID(), nil), &testRun2)
	assert.Equal(t, testRun, testRun2)
}

func TestAuthenticateUploader(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := NewAPI(ctx)

	assert.False(t, a.AuthenticateUploader("user", "123"))

	key := datastore.NewKey(ctx, "Uploader", "user", 0, nil)
	datastore.Put(ctx, key, &shared.Uploader{Username: "user", Password: "123"})
	assert.True(t, a.AuthenticateUploader("user", "123"))
}

func TestFetchWithTimeout_success(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := NewAPI(ctx)

	hello := []byte("Hello, world!")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(hello)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	body, err := a.FetchGzip(server.URL, time.Second)
	assert.Nil(t, err)
	defer body.Close()
	content, err := ioutil.ReadAll(body)
	assert.Nil(t, err)
	assert.Equal(t, hello, content)
}

func TestFetchWithTimeout_404(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := NewAPI(ctx)

	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	body, err := a.FetchGzip(server.URL, time.Second)
	assert.Nil(t, body)
	assert.EqualError(t, err, "server returned 404 Not Found")
}
