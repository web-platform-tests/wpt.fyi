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

// IsExperimental returns true if the run is labelled experimental.
func (r TestRun) IsExperimental() bool {
	return r.hasLabel(ExperimentalLabel)
}

// IsPRBase returns true if the run is labelled experimental.
func (r TestRun) IsPRBase() bool {
	return r.hasLabel(PRBaseLabel)
}

func (r TestRun) hasLabel(label string) bool {
	return StringSliceContains(r.Labels, label)
}

// Channel return the channel label, if any, for the given run.
func (r TestRun) Channel() string {
	for _, label := range r.Labels {
		switch label {
		case StableLabel,
			BetaLabel,
			ExperimentalLabel:
			return label
		}
	}
	return ""
}

// PendingTestRunStage represents the stage of a test run in its life cycle.
type PendingTestRunStage int

// Constant enums for PendingTestRunStage
const (
	StageGitHubQueued     PendingTestRunStage = 100
	StageGitHubInProgress PendingTestRunStage = 200
	StageCIRunning        PendingTestRunStage = 300
	StageCIFinished       PendingTestRunStage = 400
	StageGitHubSuccess    PendingTestRunStage = 500
	StageGitHubFailure    PendingTestRunStage = 550
	StageWptFyiReceived   PendingTestRunStage = 600
	StageWptFyiProcessing PendingTestRunStage = 700
	StageValid            PendingTestRunStage = 800
	StageInvalid          PendingTestRunStage = 850
	StageEmpty            PendingTestRunStage = 851
	StageDuplicate        PendingTestRunStage = 852
)

func (s PendingTestRunStage) String() string {
	switch s {
	case StageGitHubQueued:
		return "GITHUB_QUEUED"
	case StageGitHubInProgress:
		return "GITHUB_IN_PROGRESS"
	case StageCIRunning:
		return "CI_RUNNING"
	case StageCIFinished:
		return "CI_FINISHED"
	case StageGitHubSuccess:
		return "GITHUB_SUCCESS"
	case StageGitHubFailure:
		return "GITHUB_FAILURE"
	case StageWptFyiReceived:
		return "WPTFYI_RECEIVED"
	case StageWptFyiProcessing:
		return "WPTFYI_PROCESSING"
	case StageValid:
		return "VALID"
	case StageInvalid:
		return "INVALID"
	case StageEmpty:
		return "EMPTY"
	case StageDuplicate:
		return "DUPLICATE"
	}
	return ""
}

// MarshalJSON is the custom JSON marshaler for PendingTestRunStage.
func (s PendingTestRunStage) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON is the custom JSON unmarshaler for PendingTestRunStage.
func (s *PendingTestRunStage) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	switch str {
	case "GITHUB_QUEUED":
		*s = StageGitHubQueued
	case "GITHUB_IN_PROGRESS":
		*s = StageGitHubInProgress
	case "CI_RUNNING":
		*s = StageCIRunning
	case "CI_FINISHED":
		*s = StageCIFinished
	case "GITHUB_SUCCESS":
		*s = StageGitHubSuccess
	case "GITHUB_FAILURE":
		*s = StageGitHubFailure
	case "WPTFYI_RECEIVED":
		*s = StageWptFyiReceived
	case "WPTFYI_PROCESSING":
		*s = StageWptFyiProcessing
	case "VALID":
		*s = StageValid
	case "INVALID":
		*s = StageInvalid
	case "EMPTY":
		*s = StageEmpty
	case "DUPLICATE":
		*s = StageDuplicate
	default:
		return fmt.Errorf("unknown stage: %s", str)
	}
	if s.String() != str {
		return fmt.Errorf("enum conversion error: %s != %s", s.String(), str)
	}
	return nil
}

// PendingTestRun represents a TestRun that has started, but is not yet
// completed.
type PendingTestRun struct {
	ID               int64               `json:"id" datastore:"-"`
	CheckRunID       int64               `json:"check_run_id" datastore:",omitempty"`
	FullRevisionHash string              `json:"full_revision_hash"`
	Uploader         string              `json:"uploader"`
	Error            string              `json:"error" datastore:",omitempty"`
	Stage            PendingTestRunStage `json:"stage"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// Transition sets Stage to next if the transition is allowed; otherwise an
// error is returned.
func (s *PendingTestRun) Transition(next PendingTestRunStage) error {
	if next == 0 || s.Stage > next {
		return fmt.Errorf("cannot transition from %s to %s", s.Stage.String(), next.String())
	}
	s.Stage = next
	return nil
}

// PendingTestRunByUpdated sorts the pending test runs by updated (asc)
type PendingTestRunByUpdated []PendingTestRun

func (a PendingTestRunByUpdated) Len() int           { return len(a) }
func (a PendingTestRunByUpdated) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PendingTestRunByUpdated) Less(i, j int) bool { return a[i].Updated.Before(a[j].Updated) }

// CheckSuite entities represent a GitHub check request that has been noted by
// wpt.fyi, and will cause creation of a completed check_run when results arrive
// for the PR.
type CheckSuite struct {
	// SHA of the revision that requested a check suite.
	SHA string `json:"sha"`
	// The GitHub app ID for the custom wpt.fyi check.
	AppID int64 `json:"app_id"`
	// The GitHub app installation ID for custom wpt.fyi check
	InstallationID int64  `json:"installation"`
	Owner          string `json:"owner"` // Owner username
	Repo           string `json:"repo"`
	PRNumbers      []int  `json:"pr_numbers"`
}

// LabelsSet creates a set from the run's labels.
func (r TestRun) LabelsSet() mapset.Set {
	runLabels := mapset.NewSet()
	for _, label := range r.Labels {
		runLabels.Add(label)
	}
	return runLabels
}

// TestRuns is a helper type for an array of TestRun entities.
type TestRuns []TestRun

func (t TestRuns) Len() int           { return len(t) }
func (t TestRuns) Less(i, j int) bool { return t[i].TimeStart.Before(t[j].TimeStart) }
func (t TestRuns) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

// SetTestRunIDs sets the ID field for each run, from the given ids.
func (t TestRuns) SetTestRunIDs(ids TestRunIDs) {
	for i := 0; i < len(ids) && i < len(t); i++ {
		t[i].ID = ids[i]
	}
}

// GetTestRunIDs gets an array of the IDs for the TestRun entities in the array.
func (t TestRuns) GetTestRunIDs() TestRunIDs {
	ids := make([]int64, len(t))
	for i, run := range t {
		ids[i] = run.ID
	}
	return ids
}

// OldestRunTimeStart returns the TimeStart of the oldest run in the set.
func (t TestRuns) OldestRunTimeStart() time.Time {
	if len(t) < 1 {
		return time.Time{}
	}
	oldest := time.Now()
	for _, run := range t {
		if run.TimeStart.Before(oldest) {
			oldest = run.TimeStart
		}
	}
	return oldest
}

// ProductTestRuns is a tuple of a product and test runs loaded for it.
type ProductTestRuns struct {
	Product  ProductSpec
	TestRuns TestRuns
}

// TestRunsByProduct is an array of tuples of {product, matching runs}, returned
// when a TestRun query is executed.
type TestRunsByProduct []ProductTestRuns

// AllRuns returns an array of all the loaded runs.
func (t TestRunsByProduct) AllRuns() TestRuns {
	var runs TestRuns
	for _, p := range t {
		runs = append(runs, p.TestRuns...)
	}
	return runs
}

// First returns the first TestRun
func (t TestRunsByProduct) First() *TestRun {
	all := t.AllRuns()
	if len(all) > 0 {
		return &all[0]
	}
	return nil
}

// ProductTestRunKeys is a tuple of a product and test run keys loaded for it.
type ProductTestRunKeys struct {
	Product ProductSpec
	Keys    []Key
}

// KeysByProduct is an array of tuples of {product, matching keys}, returned
// when a TestRun key query is executed.
type KeysByProduct []ProductTestRunKeys

// AllKeys returns an array of all the loaded keys.
func (t KeysByProduct) AllKeys() []Key {
	var keys []Key
	for _, v := range t {
		keys = append(keys, v.Keys...)
	}
	return keys
}

// TestRunIDs is a helper for an array of TestRun IDs.
type TestRunIDs []int64

// GetTestRunIDs extracts the TestRunIDs from loaded datastore keys.
func GetTestRunIDs(keys []Key) TestRunIDs {
	result := make(TestRunIDs, len(keys))
	for i := range keys {
		result[i] = keys[i].IntID()
	}
	return result
}

// GetKeys returns a slice of keys for the TestRunIDs in the given datastore.
func (ids TestRunIDs) GetKeys(store Datastore) []Key {
	keys := make([]Key, len(ids))
	for i := range ids {
		keys[i] = store.NewIDKey("TestRun", ids[i])
	}
	return keys
}

// LoadTestRuns is a helper for fetching the TestRuns from the datastore,
// for the gives TestRunIDs.
func (ids TestRunIDs) LoadTestRuns(store Datastore) (testRuns TestRuns, err error) {
	if len(ids) > 0 {
		keys := ids.GetKeys(store)
		testRuns = make(TestRuns, len(keys))
		if err = store.GetMulti(keys, testRuns); err != nil {
			return testRuns, err
		}
		testRuns.SetTestRunIDs(ids)
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

// LegacySearchRunResult is the results data from legacy test summarys.  These
// summaries contain a "pass count" and a "total count", where the test itself
// counts as 1, and each subtest counts as 1. The "pass count" contains any
// status values that are "PASS" or "OK".
type LegacySearchRunResult struct {
	// Passes is the number of test results in a PASS/OK state.
	Passes int `json:"passes"`
	// Total is the total number of test results for this run/file pair.
	Total int `json:"total"`
}

// SearchResult contains data regarding a particular test file over a collection
// of runs. The runs are identified externally in a parallel slice (see
// SearchResponse).
type SearchResult struct {
	// Test is the name of a test; this often corresponds to a test file path in
	// the WPT source reposiory.
	Test string `json:"test"`
	// LegacyStatus is the results data from legacy test summaries. These
	// summaries contain a "pass count" and a "total count", where the test itself
	// counts as 1, and each subtest counts as 1. The "pass count" contains any
	// status values that are "PASS" or "OK".
	LegacyStatus []LegacySearchRunResult `json:"legacy_status,omitempty"`

	// Interoperability scores. For N browsers, we have an array of
	// N+1 items, where the index X is the number of items passing in exactly
	// X of the N browsers. e.g. for 4 browsers, [0/4, 1/4, 2/4, 3/4, 4/4].
	Interop []int `json:"interop,omitempty"`

	// Subtests (names) which are included in the LegacyStatus summary.
	Subtests []string `json:"subtests,omitempty"`

	// Diff count of subtests which are included in the LegacyStatus summary.
	Diff TestDiff `json:"diff,omitempty"`
}

// SearchResponse contains a response to search API calls, including specific
// runs whose results were searched and the search results themselves.
type SearchResponse struct {
	// Runs is the specific runs for which results were retrieved. Each run, in
	// order, corresponds to a Status entry in each SearchResult in Results.
	Runs []TestRun `json:"runs"`
	// IgnoredRuns is any runs that the client requested to be included in the
	// query, but were not included. This optional field may be non-nil if, for
	// example, results are being served from an incompelte cache of runs and some
	// runs described in the query request are not resident in the cache.
	IgnoredRuns []TestRun `json:"ignored_runs,omitempty"`
	// Results is the collection of test results, grouped by test file name.
	Results []SearchResult `json:"results"`
	// MetadataResponse is a response to a wpt-metadata query.
	MetadataResponse MetadataResults `json:"metadata,omitempty"`
}
