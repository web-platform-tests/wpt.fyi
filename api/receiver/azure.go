// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"regexp"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

var (
	azureArtifactRegex = regexp.MustCompile(`/_apis/build/builds/[0-9]+/artifacts\?artifactName=([^&]+)`)
	reportPathRegex    = regexp.MustCompile(`/wpt_report.*\.json$`)
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
	if a.z == nil {
		r, err := zip.NewReader(bytes.NewReader(a.data), int64(len(a.data)))
		if err != nil {
			return nil, err
		}
		a.z = r
	}
	files := make([]*zip.File, 0, len(a.z.File))
	for _, f := range a.z.File {
		if reportPathRegex.MatchString(f.Name) {
			files = append(files, f)
		}
	}
	return files, nil
}

func getAzureArtifactName(url string) string {
	if match := azureArtifactRegex.FindStringSubmatch(url); len(match) > 1 {
		return match[1]
	}
	return ""
}

func handleAzureArtifact(a AppEngineAPI, artifactName string, url string) (int, func(int) (io.ReadCloser, error), error) {
	log := shared.GetLogger(a.Context())
	log.Debugf("Detected azure artifact %s", artifactName)
	artifactZip, err := fetchFile(a, url)
	if err != nil {
		log.Errorf("Failed to fetch %s: %s", url, err.Error())
		return 0, nil, errors.New("Failed to fetch azure artifact")
	}
	defer artifactZip.Close()
	artifact, err := newAzureArtifact(artifactName, artifactZip)
	if err != nil {
		log.Errorf("Failed to read zip: %s", err.Error())
		return 0, nil, errors.New("Invalid artifact contents")
	}
	artifactFiles, err := artifact.getReportFiles()
	if err != nil {
		log.Errorf("Failed to extract files: %s", err.Error())
		return 0, nil, errors.New("Invalid artifact contents")
	}
	results := len(artifactFiles)
	log.Debugf("Found %v report files in artifact", results)
	getFile := func(i int) (io.ReadCloser, error) {
		zipR, err := artifactFiles[i].Open()
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		w := gzip.NewWriter(buf)
		if _, err := io.Copy(w, zipR); err != nil {
			return nil, err
		} else if err := w.Close(); err != nil {
			return nil, err
		}
		return ioutil.NopCloser(buf), nil
	}
	return results, getFile, nil
}
