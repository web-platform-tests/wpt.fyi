// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"time"
)

// TestRun stores metadata for a test run (produced by run/run.py)
type TestRun struct {
	// Platform information
	BrowserName    string `json:"browser_name"`
	BrowserVersion string `json:"browser_version"`
	OSName         string `json:"os_name"`
	OSVersion      string `json:"os_version"`

	// The first 10 characters of the SHA1 of the tested WPT revision
	Revision string `json:"revision"`

	// Results URL
	ResultsURL string `json:"results_url"`

	CreatedAt time.Time `json:"created_at"`
}

// Browser holds objects that appear in browsers.json
type Browser struct {
	InitiallyLoaded bool   `json:"initially_loaded"`
	CurrentlyRun    bool   `json:"currently_run"`
	BrowserName     string `json:"browser_name"`
	BrowserVersion  string `json:"browser_version"`
	OSName          string `json:"os_name"`
	OSVersion       string `json:"os_version"`
	Sauce           bool   `json:"sauce"`
}

// Token is used for test result uploads.
type Token struct {
	Secret string `json:"secret"`
}
