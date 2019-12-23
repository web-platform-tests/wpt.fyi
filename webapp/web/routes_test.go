// +build small

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

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

func TestApiStatusBound(t *testing.T) {
	assertHandlerIs(t, "/api/status", "api-pending-test-runs")
	assertHandlerIs(t, "/api/status/pending", "api-pending-test-runs")
	assertHandlerIs(t, "/api/status/invalid", "api-pending-test-runs")
	assertHandlerIs(t, "/api/status/123", "api-pending-test-run-update")
	assertHandlerIsDefault(t, "/api/status/notavalidfilter")
}

func TestApiResultsBoundCORS(t *testing.T) {
	assertHandlerIs(t, "/api/results", "api-results")
	assertHSTS(t, "/api/results/upload")
	assertCORS(t, "/api/results")
}

func TestApiScreenshotBoundCORS(t *testing.T) {
	assertHandlerIs(t, "/api/screenshot/sha1:abc", "api-screenshot")
	assertHSTS(t, "/api/screenshot/sha1:abc")
	assertCORS(t, "/api/screenshot/sha1:abc")
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
}

func TestAdminCacheFlushBound(t *testing.T) {
	assertHandlerIs(t, "/admin/cache/flush", "admin-cache-flush")
	assertHSTS(t, "/admin/cache/flush")
}

func TestApiMetadataCORS(t *testing.T) {
	// TODO(kyleju): Test CORS for POST/GET request.
	assertHandlerIs(t, "/api/metadata", "api-metadata")
	successPost := httptest.NewRequest("OPTIONS", "/api/metadata", nil)
	successPost.Header.Set("Access-Control-Request-Headers", "content-type")
	successPost.Header.Add("Origin", "https://developer.mozilla.org")
	successPost.Header.Add("Access-Control-Request-Method", "POST")

	rr := sendHttptestRequest(successPost)

	assert.Equal(t, http.StatusOK, rr.StatusCode)
	assert.Equal(t, "Content-Type", rr.Header.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "*", rr.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "", rr.Header.Get("Access-Control-Allow-Credentials"))
}

func TestApiMetadataTriageCORS(t *testing.T) {
	// TODO(kyleju): Test CORS for PATCH request.
	assertHandlerIs(t, "/api/metadata/triage", "api-metadata-triage")

	successReq := httptest.NewRequest("OPTIONS", "/api/metadata/triage", nil)
	successReq.Header.Set("Access-Control-Request-Headers", "content-type")
	successReq.Header.Add("Origin", "https://developer.mozilla.org")
	successReq.Header.Add("Access-Control-Request-Method", "PATCH")

	rr := sendHttptestRequest(successReq)

	assert.Equal(t, http.StatusOK, rr.StatusCode)
	assert.Equal(t, "Content-Type", rr.Header.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "PATCH", rr.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "https://developer.mozilla.org", rr.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rr.Header.Get("Access-Control-Allow-Credentials"))

	invalidOriginReq := httptest.NewRequest("OPTIONS", "/api/metadata/triage", nil)
	invalidOriginReq.Header.Add("Origin", "https://foo")

	rr = sendHttptestRequest(invalidOriginReq)

	assert.Equal(t, "", rr.Header.Get("Access-Control-Allow-Origin"))

	invalidMethodReq := httptest.NewRequest("OPTIONS", "/api/metadata/triage", nil)
	invalidMethodReq.Header.Set("Access-Control-Request-Headers", "content-type")
	invalidMethodReq.Header.Add("Origin", "https://developer.mozilla.org")
	invalidMethodReq.Header.Add("Access-Control-Request-Method", "POST")

	rr = sendHttptestRequest(invalidMethodReq)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.StatusCode)
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

func sendHttptestRequest(req *http.Request) *http.Response {
	rr := httptest.NewRecorder()
	handler, _ := http.DefaultServeMux.Handler(req)
	handler.ServeHTTP(rr, req)
	return rr.Result()
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

func assertHandlerIsDefault(t *testing.T, path string) {
	assertHandlerIs(t, path, "results-legacy")
}
