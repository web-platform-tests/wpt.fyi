// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
)

type azureArtifact struct {
	name string
	data []byte
	z    *zip.Reader
}

func newAzureArtifact(artifactName string, zip io.Reader) (azureArtifact, error) {
	data, err := ioutil.ReadAll(zip)
	return azureArtifact{
		name: artifactName,
		data: data,
	}, err
}

// extractReports extracts report files from the given zip.
func (a azureArtifact) getReportFiles() ([]*zip.File, error) {
	reportPath, err := regexp.Compile(fmt.Sprintf(`%s/wpt_report.*\.json$`, a.name))
	if err != nil {
		return nil, err
	}
	if a.z == nil {
		r, err := zip.NewReader(bytes.NewReader(a.data), int64(len(a.data)))
		if err != nil {
			return nil, err
		}
		a.z = r
	}
	files := make([]*zip.File, 0, len(a.z.File))
	for _, f := range a.z.File {
		if reportPath.MatchString(f.Name) {
			files = append(files, f)
		}
	}
	return files, nil
}
