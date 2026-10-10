// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestPackageRegex(t *testing.T) {
	assert.True(t, packageRegex.MatchString(`import '@polymer/Stuff';`))
	assert.True(t, packageRegex.MatchString(`import "@polymer/Stuff";`))
	assert.True(t, packageRegex.MatchString(`import { Cats, Dogs } from '@polymer/Animals';`))
	assert.True(t, packageRegex.MatchString(`import { Cats, Dogs } from "@polymer/Animals";`))
	assert.True(t, packageRegex.MatchString(`import '@polymer/polymer/lib/utils/gestures.js'`))
	assert.True(t, packageRegex.MatchString(`import "@polymer/polymer/lib/utils/gestures.js"`))
	assert.True(t, packageRegex.MatchString(`import * as gestures from '@polymer/polymer/lib/utils/gestures.js';`))
	assert.True(t, packageRegex.MatchString(`import * as gestures from "@polymer/polymer/lib/utils/gestures.js";`))

	assert.False(t, packageRegex.MatchString(`function import() { return "no" };`))

	// Minified imports, as published by Lit, are left for the import map in
	// templates/_import_map.html to resolve.
	assert.False(t, packageRegex.MatchString(`import"@lit/reactive-element";`))
	assert.False(t, packageRegex.MatchString(`import{ReactiveElement as t}from"@lit/reactive-element";`))
}

func TestComponentsHandler_LitServedUnchanged(t *testing.T) {
	for _, path := range []string{
		"lit/index.js",
		"lit/node_modules/lit-element/lit-element.js",
		"lit/node_modules/lit-html/lit-html.js",
		"@lit/reactive-element/reactive-element.js",
	} {
		want, err := nodeModules.ReadFile("node_modules/" + path)
		if !assert.Nil(t, err, path) {
			continue
		}
		r := httptest.NewRequest(http.MethodGet, "/node_modules/"+path, nil)
		r = mux.SetURLVars(r, map[string]string{"path": path})
		w := httptest.NewRecorder()
		componentsHandler(w, r)
		assert.Equal(t, http.StatusOK, w.Code, path)
		assert.Equal(t, string(want), w.Body.String(), path)
	}
}

func TestImportMapTargetsExist(t *testing.T) {
	var out bytes.Buffer
	if !assert.Nil(t, templates.ExecuteTemplate(&out, "_import_map.html", nil)) {
		return
	}
	html := out.String()
	start := strings.Index(html, `<script type="importmap">`) + len(`<script type="importmap">`)
	end := strings.LastIndex(html, "</script>")
	if !assert.True(t, start < end, "no import map in %q", html) {
		return
	}
	var importMap struct {
		Imports map[string]string            `json:"imports"`
		Scopes  map[string]map[string]string `json:"scopes"`
	}
	if !assert.Nil(t, json.Unmarshal([]byte(html[start:end]), &importMap)) {
		return
	}
	assert.NotEmpty(t, importMap.Imports)

	// Scopes, and targets that end in "/", are directories. Other targets are
	// files.
	urls := []string{}
	for _, target := range importMap.Imports {
		urls = append(urls, target)
	}
	for scope, imports := range importMap.Scopes {
		urls = append(urls, scope)
		for _, target := range imports {
			urls = append(urls, target)
		}
	}
	for _, url := range urls {
		if !assert.True(t, strings.HasPrefix(url, "/node_modules/"), url) {
			continue
		}
		info, err := fs.Stat(nodeModules, strings.TrimSuffix(strings.TrimPrefix(url, "/"), "/"))
		if assert.Nil(t, err, url) {
			assert.Equal(t, strings.HasSuffix(url, "/"), info.IsDir(), url)
		}
	}
}
