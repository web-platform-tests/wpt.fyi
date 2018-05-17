// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
)

// Product uniquely defines a browser version, running on an OS version.
type Product struct {
	BrowserName    string `json:"browser_name"`
	BrowserVersion string `json:"browser_version"`
	OSName         string `json:"os_name"`
	OSVersion      string `json:"os_version"`
}

func (p Product) String() string {
	s := p.BrowserName
	if p.BrowserVersion != "" {
		s = fmt.Sprintf("%s-%s", s, p.BrowserVersion)
	}
	if p.OSName != "" {
		s = fmt.Sprintf("%s-%s", s, p.OSName)
		if p.OSVersion != "" {
			s = fmt.Sprintf("%s-%s", s, p.OSVersion)
		}
	}
	return s
}

// ByBrowserName is a []Product sortable by BrowserName values.
type ByBrowserName []Product

func (e ByBrowserName) Len() int           { return len(e) }
func (e ByBrowserName) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e ByBrowserName) Less(i, j int) bool { return e[i].BrowserName < e[j].BrowserName }

// Version is a struct for a parsed semantic version string.
type Version struct {
	Major    string
	Minor    string
	Revision string
}

// ProductAtRevision defines a WPT run for a specific product, at a
// specific hash of the WPT repo.
type ProductAtRevision struct {
	Product

	// The first 10 characters of the SHA1 of the tested WPT revision
	Revision string `json:"revision"`
}

func (p ProductAtRevision) String() string {
	return fmt.Sprintf("%s@%s", p.Product.String(), p.Revision)
}

// TestRun stores metadata for a test run (produced by run/run.py)
type TestRun struct {
	ProductAtRevision

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

// Manifest represents a JSON blob of all the WPT tests.
type Manifest struct {
	Items   ManifestItems `json:"items,omitempty"`
	Version *int          `json:"version,omitempty"`
}

// FilterByPath filters all the manifest items by path.
func (m Manifest) FilterByPath(paths mapset.Set) (result Manifest, err error) {
	result = m
	if result.Items.Manual, err = m.Items.Manual.FilterByPath(paths); err != nil {
		return result, err
	}
	if result.Items.Reftest, err = m.Items.Reftest.FilterByPath(paths); err != nil {
		return result, err
	}
	if result.Items.TestHarness, err = m.Items.TestHarness.FilterByPath(paths); err != nil {
		return result, err
	}
	if result.Items.WDSpec, err = m.Items.WDSpec.FilterByPath(paths); err != nil {
		return result, err
	}
	return result, nil
}

// ManifestItems groups the different manifest item types.
type ManifestItems struct {
	Manual      ManifestItem `json:"manual"`
	Reftest     ManifestItem `json:"reftest"`
	TestHarness ManifestItem `json:"testharness"`
	WDSpec      ManifestItem `json:"wdspec"`
}

// ManifestItem represents a map of files to item details, for a specific test type.
type ManifestItem map[string][][]*json.RawMessage

// FilterByPath culls out entries in the ManifestItem that don't have any items with
// a URL that starts with the given path.
func (m ManifestItem) FilterByPath(paths mapset.Set) (item ManifestItem, err error) {
	if m == nil {
		return nil, nil
	}
	filtered := make(ManifestItem)
	for path, items := range m {
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
	return filtered, nil
}
