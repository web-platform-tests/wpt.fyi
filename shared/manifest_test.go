// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExplodePossibleFilename_AnyJS(t *testing.T) {
	filename := "/test/file.any.js"
	exploded := []string{
		"/test/file.any.html",
		"/test/file.any.worker.html",
		"/test/file.any.serviceworker.html",
		"/test/file.any.sharedworker.html",
	}
	assert.Equal(t, ExplodePossibleFilenames(filename), exploded)
}

func TestExplodePossibleFilename_WindowJS(t *testing.T) {
	filename := "/test/file.window.js"
	exploded := []string{
		"/test/file.window.html",
	}
	assert.Equal(t, ExplodePossibleFilenames(filename), exploded)
}

func TestExplodePossibleFilename_Standard(t *testing.T) {
	filename := "/test/file.html"
	assert.Nil(t, ExplodePossibleFilenames(filename))
}

func TestExplodePossibleRenames_AnyJS(t *testing.T) {
	before, after := "/test/file.any.js", "/test/file.https.any.js"
	renames := map[string]string{
		before:                              after,
		"/test/file.any.html":               "/test/file.https.any.html",
		"/test/file.any.worker.html":        "/test/file.https.any.worker.html",
		"/test/file.any.serviceworker.html": "/test/file.https.any.serviceworker.html",
		"/test/file.any.sharedworker.html":  "/test/file.https.any.sharedworker.html",
	}
	assert.Equal(t, ExplodePossibleRenames(before, after), renames)
}

func TestExplodePossibleRenames_WindowJS(t *testing.T) {
	before, after := "/test/file.window.js", "/test/file.https.window.js"
	renames := map[string]string{
		before:                   after,
		"/test/file.window.html": "/test/file.https.window.html",
	}
	assert.Equal(t, ExplodePossibleRenames(before, after), renames)
}

func TestExplodePossibleRenames_WorkerJS(t *testing.T) {
	before, after := "/test/file.worker.js", "/test/file.https.worker.js"
	renames := map[string]string{
		before:                   after,
		"/test/file.worker.html": "/test/file.https.worker.html",
	}
	assert.Equal(t, ExplodePossibleRenames(before, after), renames)
}
