// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"bytes"
	"io/ioutil"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFiles(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	// artifact_test.zip has a single dir, artifact_test/, containing 2 files, wpt_report_{1,2}.json
	data, err := ioutil.ReadFile(path.Join(path.Dir(filename), "artifact_test.zip"))
	if err != nil {
		assert.FailNow(t, "Failed to read artifact_test.zip", err.Error())
	}

	artifact, err := newAzureArtifact("artifact_test", bytes.NewReader(data))
	files, err := artifact.getReportFiles()
	if err != nil {
		assert.FailNow(t, "Failed to read zip", err.Error())
	}
	assert.Equal(t, files[0].Name, "artifact_test/wpt_report_1.json")
	assert.Equal(t, files[1].Name, "artifact_test/wpt_report_2.json")
}

func TestGetAzureArtifactName(t *testing.T) {
	url := "https://dev.azure.com/web-platform-tests/b14026b4-9423-4454-858f-bf76cf6d1faa/_apis/build/builds/4230/artifacts?artifactName=results&api-version=5.0&%24format=zip"
	a := getAzureArtifactName(url)
	assert.Equal(t, "results", a)
}
