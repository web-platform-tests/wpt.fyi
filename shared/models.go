// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
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

type Manifest struct {
	Items   ManifestItems `json:"items,omitempty"`
	Version *int          `json:"version,omitempty"`
}

func (m Manifest) FilterByPath(paths mapset.Set) (result Manifest, err error) {
	if result.Items.Manual, err = m.Items.Manual.FilterByPath(paths); err != nil {
		return result, err
	}
	if result.Items.Reftest, err = m.Items.Reftest.FilterByPath(paths); err != nil {
		return result, err
	}
	if result.Items.TestHarness, err = m.Items.TestHarness.FilterByPath(paths); err != nil {
		return result, err
	}
	return result, nil
}

type ManifestItems struct {
	Manual      *ManifestItem `json:"manual,omitempty"`
	Reftest     *ManifestItem `json:"reftest,omitempty"`
	TestHarness *ManifestItem `json:"testharness,omitempty"`
}

type ManifestItem map[string][][]*json.RawMessage

func (m *ManifestItem) FilterByPath(paths mapset.Set) (item *ManifestItem, err error) {
	if m == nil {
		return nil, nil
	}
	filtered := make(ManifestItem)
	for path, items := range *m {
		match := false
		for _, item := range items {
			var url string
			if err = json.Unmarshal(*item[0], &url); err != nil {
				return nil, err
			}
			for prefix := range paths.Iter() {
				if strings.Index(url, prefix.(string)) == 0 {
					match = true
					break
				}
			}
		}
		if !match {
			continue
		}
		filtered[path] = items
	}
	return &filtered, nil
}
