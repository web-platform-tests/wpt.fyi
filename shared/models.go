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
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
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

// Version is a struct for a parsed version string.
type Version struct {
	Major    int
	Minor    *int
	Build    *int
	Revision *int
	Channel  string
}

func (v Version) String() string {
	s := fmt.Sprintf("%v", v.Major)
	if v.Minor != nil {
		s = fmt.Sprintf("%s.%v", s, *v.Minor)
	}
	if v.Build != nil {
		s = fmt.Sprintf("%s.%v", s, *v.Build)
	}
	if v.Revision != nil {
		s = fmt.Sprintf("%s.%v", s, *v.Revision)
	}
	if v.Channel != "" {
		s = fmt.Sprintf("%s%s", s, v.Channel)
	}
	return s
}

// ProductAtRevision defines a WPT run for a specific product, at a
// specific hash of the WPT repo.
type ProductAtRevision struct {
	Product

	// The first 10 characters of the SHA1 of the tested WPT revision.
	//
	// Deprecated: The authoritative git revision indicator is FullRevisionHash.
	Revision string `json:"revision"`

	// The complete SHA1 hash of the tested WPT revision.
	FullRevisionHash string `json:"full_revision_hash"`
}

func (p ProductAtRevision) String() string {
	return fmt.Sprintf("%s@%s", p.Product.String(), p.Revision)
}

// TestRun stores metadata for a test run (produced by run/run.py)
type TestRun struct {
	ID int64 `json:"id" datastore:"-"`

	ProductAtRevision

	// URL for summary of results, which is derived from raw results.
	ResultsURL string `json:"results_url"`

	// Time when the test run metadata was first created.
	CreatedAt time.Time `json:"created_at"`

	// Time when the test run started.
	TimeStart time.Time `json:"time_start"`

	// Time when the test run ended.
	TimeEnd time.Time `json:"time_end"`

	// URL for raw results JSON object. Resembles the JSON output of the
	// wpt report tool.
	RawResultsURL string `json:"raw_results_url"`

	// Labels for the test run.
	Labels []string `json:"labels"`
}

// LabelsSet creates a set from the run's labels.
func (run TestRun) LabelsSet() mapset.Set {
	runLabels := mapset.NewSet()
	for _, label := range run.Labels {
		runLabels.Add(label)
	}
	return runLabels
}

// TestRuns is a helper type for an array of TestRun entities.
type TestRuns []TestRun

// GetTestRunIDs gets an array of the IDs for the TestRun entities in the array.
func (runs TestRuns) GetTestRunIDs() TestRunIDs {
	ids := make([]int64, len(runs))
	for i, run := range runs {
		ids[i] = run.ID
	}
	return ids
}

// TestRunIDs is a helper for an array of TestRun IDs.
type TestRunIDs []int64

// LoadTestRuns is a helper for fetching the TestRuns from the datastore,
// for the gives TestRunIDs.
func (ids TestRunIDs) LoadTestRuns(ctx context.Context) (testRuns TestRuns, err error) {
	if len(ids) > 0 {
		keys := make([]*datastore.Key, len(ids))
		for i, id := range ids {
			keys[i] = datastore.NewKey(ctx, "TestRun", "", id, nil)
		}
		testRuns = make(TestRuns, len(keys))
		if err = datastore.GetMulti(ctx, keys, testRuns); err != nil {
			return testRuns, err
		}
		for i, id := range ids {
			testRuns[i].ID = id
		}
	}
	return testRuns, err
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
func (m Manifest) FilterByPath(paths ...string) (result Manifest, err error) {
	result = m
	if result.Items.Manual, err = m.Items.Manual.FilterByPath(paths...); err != nil {
		return result, err
	}
	if result.Items.Reftest, err = m.Items.Reftest.FilterByPath(paths...); err != nil {
		return result, err
	}
	if result.Items.TestHarness, err = m.Items.TestHarness.FilterByPath(paths...); err != nil {
		return result, err
	}
	if result.Items.WDSpec, err = m.Items.WDSpec.FilterByPath(paths...); err != nil {
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
func (m ManifestItem) FilterByPath(paths ...string) (item ManifestItem, err error) {
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
			for _, prefix := range paths {
				if strings.Index(url, prefix) == 0 {
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

// Uploader is a username/password combo accepted by
// the results receiver.
type Uploader struct {
	Username string
	Password string
}

// Flag represents an enviroment feature flag's default state.
type Flag struct {
	Name    string `datastore:"-"` // Name is the key in datastore.
	Enabled bool
}
