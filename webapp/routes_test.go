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
}

func TestAboutBound(t *testing.T) {
	assertHandlerIs(t, "/about", "about")
}

func TestInteropBound(t *testing.T) {
	assertHandlerIs(t, "/interop", "interop")
	assertHandlerIs(t, "/interop/", "interop")
	assertHandlerIs(t, "/interop/2dcontext", "interop")
}

func TestInteropAnomaliesBound(t *testing.T) {
	assertHandlerIs(t, "/anomalies", "anomaly")
}

func TestRunsBound(t *testing.T) {
	assertBound(t, "/test-runs")
}

func TestRunsBoundHSTS(t *testing.T) {
	assertHSTS(t, "/test-runs")
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

func TestApiResultsUploadBound(t *testing.T) {
	assertHSTS(t, "/api/results/upload")
}

func TestResultsBound(t *testing.T) {
	assertBound(t, "/results")
}

func TestAdminResultsUploadBound(t *testing.T) {
	assertHandlerIs(t, "/admin/results/upload", "admin-results-upload")
	assertHSTS(t, "/admin/results/upload")
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
	assert.Equal(
		t,
		"[max-age=31536000; preload]",
		fmt.Sprintf("%s", rr.HeaderMap["Strict-Transport-Security"]))
}
