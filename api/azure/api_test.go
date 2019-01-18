// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package azure

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFiles(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	for i, name := range []string{"artifact_test", "artifact_test_2"} {
		t.Run(name, func(t *testing.T) {
			// artifact_test.zip has a single dir, artifact_test/, containing 2 files, wpt_report_{1,2}.json
			data, err := ioutil.ReadFile(path.Join(path.Dir(filename), fmt.Sprintf("%s.zip", name)))
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("Failed to read %s.zip", name), err.Error())
			}
			buf := new(bytes.Buffer)
			err = extractReports(context.Background(), name, data, buf)
			if err != nil {
				assert.FailNow(t, "Failed to extract reports", err.Error())
			}
			assert.Nil(t, err)
			content := string(buf.Bytes())
			if i == 0 {
				assert.Contains(t, content, "wpt_report.json")
			} else {
				assert.Contains(t, content, "wpt_report_1.json")
				assert.Contains(t, content, "wpt_report_2.json")
			}
		})
	}
}
