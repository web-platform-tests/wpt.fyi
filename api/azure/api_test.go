// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package azure

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFiles(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	data, err := ioutil.ReadFile(path.Join(path.Dir(filename), "artifact_test.zip"))
	if err != nil {
		assert.FailNow(t, "Failed to read artifact_test.zip", err.Error())
	}
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	err = extractReports(context.Background(), "artifact_test", data, writer)
	if err != nil {
		assert.FailNow(t, "Failed to extract reports", err.Error())
	}
	writer.Close()
	assert.Nil(t, err)
	content := string(buf.Bytes())
	assert.Contains(t, content, "wpt_report_1.json")
	assert.Contains(t, content, "wpt_report_2.json")
}
