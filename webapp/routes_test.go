// +build small

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestLandingPageBound(t *testing.T) {
	// Note that init() is always called by the Golang runtime.
	assertHandlerIs(t, "/", "results-legacy")
	assertHSTS(t, "/")
	assertHandlerIs(t, "/2dcontext", "results-legacy")
	assertHandlerIs(t, "/BackgroundSync/interfaces.any.html", "results-legacy")
}

func TestAboutBound(t *testing.T) {
	assertHandlerIs(t, "/about", "about")
}

func TestAnalyzerBound(t *testing.T) {
	assertHandlerIs(t, "/analyzer", "analyzer")
}

func TestFlagsBound(t *testing.T) {
	assertHandlerIs(t, "/flags", "flags")
}

func TestInteropBound(t *testing.T) {
	assertHandlerIs(t, "/interop", "interop")
	assertHandlerIs(t, "/interop/", "interop")
	assertHandlerIs(t, "/interop/2dcontext", "interop")
	assertHandlerIs(t, "/interop/BackgroundSync/interfaces.any.html", "interop")
}

func TestRunsBound(t *testing.T) {
	assertHandlerIs(t, "/test-runs", "test-runs")
}

func TestApiDiffBound(t *testing.T) {
	assertHandlerIs(t, "/api/diff", "api-diff")
}

func TestApiInteropBound(t *testing.T) {
	assertHandlerIs(t, "/api/interop", "api-interop")
}

func TestApiManifestBound(t *testing.T) {
	assertHandlerIs(t, "/api/manifest", "api-manifest")
}

func TestApiRunsBound(t *testing.T) {
	assertHandlerIs(t, "/api/runs", "api-test-runs")
}

func TestApiShasBound(t *testing.T) {
	assertHandlerIs(t, "/api/shas", "api-shas")
}

func TestApiRunBound(t *testing.T) {
	assertHandlerIs(t, "/api/run", "api-test-run")
	assertHandlerIs(t, "/api/runs/123", "api-test-run")
}

func TestApiResultsBoundCORS(t *testing.T) {
	assertHandlerIs(t, "/api/results", "api-results")
	assertHSTS(t, "/api/results/upload")
	assertCORS(t, "/api/results")
}

func TestApiResultsUploadBoundHSTS(t *testing.T) {
	assertHandlerIs(t, "/api/results/upload", "api-results-upload")
	assertHSTS(t, "/api/results/upload")
	assertNoCORS(t, "/api/results/upload")
}

func TestApiResultsCreateBoundHSTS(t *testing.T) {
	assertHandlerIs(t, "/api/results/create", "api-results-create")
	assertHSTS(t, "/api/results/create")
	assertNoCORS(t, "/api/results/create")
}

func TestResultsBound(t *testing.T) {
	assertHandlerIs(t, "/results", "results")
	assertHandlerIs(t, "/results/", "results")
	assertHandlerIs(t, "/results/2dcontext", "results")
	assertHandlerIs(t, "/results/BackgroundSync/interfaces.any.html", "results")
}

func TestAdminResultsUploadBound(t *testing.T) {
	assertHandlerIs(t, "/admin/results/upload", "admin-results-upload")
	assertHSTS(t, "/admin/results/upload")
}

func TestAdminCacheFlushBound(t *testing.T) {
	assertHandlerIs(t, "/admin/cache/flush", "admin-cache-flush")
	assertHSTS(t, "/admin/cache/flush")
}

func assertBound(t *testing.T, path string) mux.RouteMatch {
	req := httptest.NewRequest("GET", path, nil)
	router := shared.Router()
	match := mux.RouteMatch{}
	assert.Truef(t, router.Match(req, &match), "%s should match a route", path)
	return match
}

func assertHandlerIs(t *testing.T, path string, name string) {
	match := assertBound(t, path)
	if match.Route != nil {
		assert.Equal(t, name, match.Route.GetName())
	}
}

func assertHSTS(t *testing.T, path string) {
	req := httptest.NewRequest("GET", path, nil)
	rr := httptest.NewRecorder()
	handler, _ := http.DefaultServeMux.Handler(req)
	handler.ServeHTTP(rr, req)
	res := rr.Result()
	assert.Equal(
		t,
		"[max-age=31536000; preload]",
		fmt.Sprintf("%s", res.Header["Strict-Transport-Security"]))
}

func assertCORS(t *testing.T, path string) {
	req := httptest.NewRequest("OPTIONS", path, nil)
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	rr := httptest.NewRecorder()
	handler, _ := http.DefaultServeMux.Handler(req)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

	req = httptest.NewRequest("GET", path, nil)
	req.Header.Add("Origin", "localhost:8080")
	rr = httptest.NewRecorder()
	handler, _ = http.DefaultServeMux.Handler(req)
	handler.ServeHTTP(rr, req)
	res := rr.Result()
	assert.Equal(
		t,
		"*",
		res.Header.Get("Access-Control-Allow-Origin"))
}

func assertNoCORS(t *testing.T, path string) {
	req := httptest.NewRequest("GET", path, nil)
	rr := httptest.NewRecorder()
	handler, _ := http.DefaultServeMux.Handler(req)
	handler.ServeHTTP(rr, req)
	res := rr.Result()
	assert.Equal(t, "", res.Header.Get("Access-Control-Allow-Origin"))
}
