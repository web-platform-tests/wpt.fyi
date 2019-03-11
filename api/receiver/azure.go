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
	// This is the pattern for the downloadURL field in
	// https://docs.microsoft.com/en-us/rest/api/azure/devops/build/artifacts/get?view=azure-devops-rest-4.1
	azureArtifactRegex = regexp.MustCompile(`/_apis/build/builds/[0-9]+/artifacts\?artifactName=([^&]+)`)
	// These are our own naming conventions (defined by the jobs).
	reportPathRegex     = regexp.MustCompile(`/wpt_report.*\.json$`)
	screenshotPathRegex = regexp.MustCompile(`/wpt_screenshot.*\.db$`)
)

func getAzureArtifactName(url string) string {
	if match := azureArtifactRegex.FindStringSubmatch(url); len(match) > 1 {
		return match[1]
	}
	return ""
}

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

// getFilesMatchingPattern returns the number and a slice of *zip.File whose
// filenames match the given pattern.
func (a azureArtifact) getFilesMatchingPattern(p *regexp.Regexp) (int, []*zip.File, error) {
	if a.z == nil {
		r, err := zip.NewReader(bytes.NewReader(a.data), int64(len(a.data)))
		if err != nil {
			return 0, nil, err
		}
		a.z = r
	}
	files := make([]*zip.File, 0, len(a.z.File))
	for _, f := range a.z.File {
		if p.MatchString(f.Name) {
			files = append(files, f)
		}
	}
	return len(files), files, nil
}

// gzipReaderFromZip returns a *function* that returns the gzipped i-th file in
// the given slice of *zip.File (the return value is intended to be used by
// sendResultsToProcessor).
func gzipReaderFromZip(files []*zip.File) func(int) (io.ReadCloser, error) {
	return func(i int) (io.ReadCloser, error) {
		zipR, err := files[i].Open()
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		gzipW := gzip.NewWriter(buf)
		if _, err := io.Copy(gzipW, zipR); err != nil {
			return nil, err
		}
		if err := gzipW.Close(); err != nil {
			return nil, err
		}
		return ioutil.NopCloser(buf), nil
	}
}

func handleAzureArtifact(a API, artifactName string, url string) (
	results, screenshots int, getResult, getScreenshot func(int) (io.ReadCloser, error), err error) {
	log := shared.GetLogger(a.Context())
	log.Debugf("Detected Azure artifact %s", artifactName)
	artifactZip, err := fetchFile(a, url)
	if err != nil {
		log.Errorf("Failed to fetch %s: %s", url, err.Error())
		return 0, 0, nil, nil, errors.New("Failed to fetch Azure artifact")
	}
	defer artifactZip.Close()
	artifact, err := newAzureArtifact(artifactName, artifactZip)
	if err != nil {
		log.Errorf("Failed to read zip: %s", err.Error())
		return 0, 0, nil, nil, errors.New("Invalid artifact contents")
	}

	var resultFiles, screenshotFiles []*zip.File

	results, resultFiles, err = artifact.getFilesMatchingPattern(reportPathRegex)
	if err != nil {
		log.Errorf("Failed to extract result files: %s", err.Error())
		return 0, 0, nil, nil, errors.New("Invalid artifact contents")
	}
	log.Debugf("Found %v report files in artifact", results)
	getResult = gzipReaderFromZip(resultFiles)

	screenshots, screenshotFiles, err = artifact.getFilesMatchingPattern(screenshotPathRegex)
	if err != nil {
		// Non-fatal!
		log.Warningf("Failed to extract screenshot files: %s", err.Error())
	} else {
		log.Debugf("Found %v screenshot files in artifact", screenshots)
		getScreenshot = gzipReaderFromZip(screenshotFiles)
	}

	return results, screenshots, getResult, getScreenshot, nil
}
