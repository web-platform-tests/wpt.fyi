// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
)

func TestLandingPageBound(t *testing.T) {
	// Note that init() is always called by the Golang runtime.
	assertBound(t, "/")
	assertHandlerMatch(t, "/2dcontext", "/")
}

func TestAboutBound(t *testing.T) {
	assertBound(t, "/about")
	assertHandlerMatch(t, "/about", "/about")
}

func TestInteropBound(t *testing.T) {
	const pattern = "/interop/"
	assertBound(t, pattern)
	assertHandlerMatch(t, pattern, pattern)
	assertHandlerMatch(t, "/interop/2dcontext", pattern)
	// NOTE(lukebjerring): Trailing slash makes it a path.
	assertHandlerMatch(t, "/interop/anomalies/", pattern)
	assertHandlerMatch(t, "/interop/anomalies/2dcontext", pattern)
}

func TestInteropAnomaliesBound(t *testing.T) {
	const pattern = "/interop/anomalies"
	assertBound(t, pattern)
	assertHandlerMatch(t, pattern, pattern)
}

func TestRunsBound(t *testing.T) {
	assertBound(t, "/test-runs")
}

func TestApiDiffBound(t *testing.T) {
	assertBound(t, "/api/diff")
}

func TestApiRunsBound(t *testing.T) {
	assertBound(t, "/api/runs")
}

func TestApiRunBound(t *testing.T) {
	assertBound(t, "/api/run")
}

func TestResultsBound(t *testing.T) {
	assertBound(t, "/results")
}

func assertBound(t *testing.T, path string) {
	req := httptest.NewRequest("GET", path, nil)
	handler, _ := http.DefaultServeMux.Handler(req)
	assert.NotNil(t, handler)
}

func assertHandlerMatch(t *testing.T, path string, pattern string) {
	req := httptest.NewRequest("GET", path, nil)
	handler, handlerPattern := http.DefaultServeMux.Handler(req)
	assert.NotNil(t, handler)
	assert.Equal(t, pattern, handlerPattern)
}
