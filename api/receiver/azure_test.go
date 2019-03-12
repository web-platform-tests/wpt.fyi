// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"bytes"
	"io"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// artifact_test.zip has a single dir, artifact_test/, containing 4 files: wpt_report_{1,2}.json, wpt_screenshot_{1,2}.txt.

func readZip(t *testing.T) *os.File {
	_, filename, _, _ := runtime.Caller(0)
	file, err := os.Open(path.Join(path.Dir(filename), "artifact_test.zip"))
	if err != nil {
		assert.FailNow(t, "Failed to open artifact_test.zip")
	}
	return file
}

func TestGetFilesMatchingPattern(t *testing.T) {
	file := readZip(t)
	defer file.Close()
	artifact, err := newAzureArtifact("artifact_test", file)
	assert.NoError(t, err)

	num, files, err := artifact.getFilesMatchingPattern(reportPathRegex)
	assert.NoError(t, err)
	assert.Equal(t, 2, num)
	assert.Equal(t, "artifact_test/wpt_report_1.json", files[0].Name)
	assert.Equal(t, "artifact_test/wpt_report_2.json", files[1].Name)

	num, files, err = artifact.getFilesMatchingPattern(screenshotPathRegex)
	assert.NoError(t, err)
	assert.Equal(t, 2, num)
	assert.Equal(t, "artifact_test/wpt_screenshot_1.txt", files[0].Name)
	assert.Equal(t, "artifact_test/wpt_screenshot_2.txt", files[1].Name)
}

func TestGzipReaderFromZip(t *testing.T) {
	file := readZip(t)
	defer file.Close()
	artifact, err := newAzureArtifact("artifact_test", file)
	assert.NoError(t, err)

	_, files, err := artifact.getFilesMatchingPattern(screenshotPathRegex)
	assert.NoError(t, err)
	getFile := gzipReaderFromZip(files)
	reader, err := getFile(0)
	assert.NoError(t, err)

	var buf bytes.Buffer
	written, err := io.Copy(&buf, reader)
	assert.NoError(t, err)
	// 10-byte header:
	assert.True(t, written > 10)
	// Gzip magic number: 1f8b
	assert.Equal(t, []byte("\x1f\x8b"), buf.Bytes()[:2])
}

func TestGetAzureArtifactName(t *testing.T) {
	url := "https://dev.azure.com/web-platform-tests/b14026b4-9423-4454-858f-bf76cf6d1faa/_apis/build/builds/4230/artifacts?artifactName=results&api-version=5.0&%24format=zip"
	a := getAzureArtifactName(url)
	assert.Equal(t, "results", a)
}
