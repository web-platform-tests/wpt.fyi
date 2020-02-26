// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManifestFilterByPath(t *testing.T) {
	manifest := []byte(`{
"items": {
	"testharness": {
		"foo": {
			"bar": {
				"test.html": [
					"1d3166465cc6f2e8f9f18f53e499ca61e12d59bd",
					[null, {}]
				]

			}
		},
		"foobar": {
			"mytest.html": [
				"2d3166465cc6f2e8f9f18f53e499ca61e12d59bd",
				[null, {}]
			]
		}
	},
	"manual": {
		"foobar": {
			"test-manual.html": [
				"3d3166465cc6f2e8f9f18f53e499ca61e12d59bd",
				[null, {}]
			]
		}
	}
},
"version": 8
}`)

	t.Run("empty match", func(t *testing.T) {
		var m Manifest
		err := json.Unmarshal(manifest, &m)
		assert.Nil(t, err)

		res, err := m.FilterByPath("/non-existent")
		assert.Nil(t, err)
		assert.Equal(t, 0, len(res.Items))
		assert.Equal(t, 8, res.Version)
	})

	t.Run("match nested", func(t *testing.T) {
		var m Manifest
		err := json.Unmarshal(manifest, &m)
		assert.Nil(t, err)

		res, err := m.FilterByPath("/foo/bar")
		assert.Nil(t, err)
		assert.Equal(t, `{"foo":{"bar":{"test.html":["1d3166465cc6f2e8f9f18f53e499ca61e12d59bd",[null,{}]]}}}`, string(res.Items["testharness"]))
		_, ok := res.Items["manual"]
		assert.False(t, ok)
		assert.Equal(t, 8, res.Version)
	})

	t.Run("match single", func(t *testing.T) {
		var m Manifest
		err := json.Unmarshal(manifest, &m)
		assert.Nil(t, err)

		res, err := m.FilterByPath("/foo")
		assert.Nil(t, err)
		assert.Equal(t, `{"foo":{"bar":{"test.html":["1d3166465cc6f2e8f9f18f53e499ca61e12d59bd",[null,{}]]}}}`, string(res.Items["testharness"]))
		_, ok := res.Items["manual"]
		assert.False(t, ok)
		assert.Equal(t, 8, res.Version)
	})

	t.Run("match multiple", func(t *testing.T) {
		var m Manifest
		err := json.Unmarshal(manifest, &m)
		assert.Nil(t, err)

		res, err := m.FilterByPath("/foobar")
		assert.Nil(t, err)
		assert.Equal(t, `{"foobar":{"mytest.html":["2d3166465cc6f2e8f9f18f53e499ca61e12d59bd",[null,{}]]}}`, string(res.Items["testharness"]))
		assert.Equal(t, `{"foobar":{"test-manual.html":["3d3166465cc6f2e8f9f18f53e499ca61e12d59bd",[null,{}]]}}`, string(res.Items["manual"]))
		assert.Equal(t, 8, res.Version)
	})
}

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
