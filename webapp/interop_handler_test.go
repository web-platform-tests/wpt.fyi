//go:build small
// +build small

package webapp

// Copyright 2022 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestInteropHandler_redirect(t *testing.T) {
	// 1999 is an invalid interop year and should be redirected.
	req := httptest.NewRequest("GET", "/interop-1999?embedded", strings.NewReader("{}"))
	req = mux.SetURLVars(req, map[string]string{
		"name":     "interop",
		"year":     "1999",
		"embedded": "true",
	})

	w := httptest.NewRecorder()
	interopHandler(w, req)
	resp := w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusTemporaryRedirect)

	loc, err := resp.Location()
	assert.Nil(t, err)
	// Check that the path has been properly updated to the current interop effort.
	assert.Equal(t, loc.Path, "/interop-2025")
	// Check if embedded param is maintained after redirect.
	assert.Equal(t, loc.RawQuery, "embedded")
}

func TestInteropHandler_redirectMobile(t *testing.T) {
	// 2021 is an invalid interop mobile year and should be redirected.
	req := httptest.NewRequest("GET", "/interop-2021?mobile-view", strings.NewReader("{}"))
	req = mux.SetURLVars(req, map[string]string{
		"name":       "interop",
		"year":       "2021",
		"mobileView": "true",
	})

	w := httptest.NewRecorder()
	interopHandler(w, req)
	resp := w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusTemporaryRedirect)

	loc, err := resp.Location()
	assert.Nil(t, err)
	// Check that the path has been properly updated to the current interop effort.
	assert.Equal(t, loc.Path, "/interop-2025")
	// Check if mobileView param is maintained after redirect.
	assert.Equal(t, loc.RawQuery, "mobile-view")
}

func TestInteropHandler_redirectdefault(t *testing.T) {
	// /interop route should redirect to the current default interop year dashboard.
	req := httptest.NewRequest("GET", "/interop?embedded", strings.NewReader("{}"))
	req = mux.SetURLVars(req, map[string]string{
		"name":     "interop",
		"embedded": "true",
	})

	w := httptest.NewRecorder()
	interopHandler(w, req)
	resp := w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusTemporaryRedirect)

	loc, err := resp.Location()
	assert.Nil(t, err)
	// Check that the path has been properly updated to the current interop effort.
	assert.Equal(t, loc.Path, "/interop-2025")
	// Check if embedded param is maintained after redirect.
	assert.Equal(t, loc.RawQuery, "embedded")
}

func TestInteropHandler_compatRedirect(t *testing.T) {
	// "/compat20XX" paths should redirect to the interop version of the given year.
	req := httptest.NewRequest("GET", "/compat2021", strings.NewReader("{}"))
	req = mux.SetURLVars(req, map[string]string{
		"name": "compat",
		"year": "2021",
	})

	w := httptest.NewRecorder()
	interopHandler(w, req)
	resp := w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusTemporaryRedirect)
}

func TestInteropHandler_success(t *testing.T) {
	// A typical "/interop-20XX" path with a valid year should not redirect.
	req := httptest.NewRequest("GET", "/interop-"+defaultRedirectYear, strings.NewReader("{}"))
	req = mux.SetURLVars(req, map[string]string{
		"name": "interop",
		"year": defaultRedirectYear,
	})

	w := httptest.NewRecorder()
	interopHandler(w, req)
	resp := w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestInteropHandler_mobileSuccess(t *testing.T) {
	// A typical "/interop-20XX" path with a valid mobile year should not redirect.
	req := httptest.NewRequest(
		"GET", "/interop-"+defaultRedirectYear+"?mobile-view", strings.NewReader("{}"))
	req = mux.SetURLVars(req, map[string]string{
		"name":       "interop",
		"year":       defaultRedirectYear,
		"mobileView": "true",
	})

	w := httptest.NewRecorder()
	interopHandler(w, req)
	resp := w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}
