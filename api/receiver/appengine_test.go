// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/taskqueue"
)

type mockGcs struct {
	mockWriter mockGcsWriter
}

// mockGcsWriter implements io.WriteCloser
type mockGcsWriter struct {
	bytes.Buffer
	finalContent []byte
}

func (m *mockGcsWriter) Close() error {
	m.finalContent = m.Bytes()
	return nil
}

func (m *mockGcs) NewWriter(bucketName, fileName, contentType, contentEncoding string) (io.WriteCloser, error) {
	return &m.mockWriter, nil
}

func TestUploadToGCS(t *testing.T) {
	a := appEngineAPIImpl{}
	mGcs := mockGcs{}
	a.gcs = &mGcs

	buffer := bytes.NewBufferString("test content")
	path, err := a.uploadToGCS("test.json", buffer, false)
	assert.Nil(t, err)

	assert.Equal(t, path, fmt.Sprintf("/%s/test.json", BufferBucket))
	assert.Equal(t, string(mGcs.mockWriter.finalContent), "test content", 0)
}

func TestScheduleResultsTask(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	assert.Nil(t, err)
	defer done()

	stats, err := taskqueue.QueueStats(ctx, []string{""})
	assert.Nil(t, err)
	assert.Equal(t, stats[0].Tasks, 0)

	a := appEngineAPIImpl{ctx: ctx}
	_, err = a.scheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "single")
	assert.Nil(t, err)

	stats, err = taskqueue.QueueStats(ctx, []string{""})
	assert.Nil(t, err)
	assert.Equal(t, stats[0].Tasks, 1)

}
