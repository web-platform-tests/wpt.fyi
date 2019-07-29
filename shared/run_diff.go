// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/run_diff_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared DiffAPI

package shared

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"google.golang.org/appengine/urlfetch"
)

// ErrRunNotInSearchCache is an error for 422 responses from the searchcache.
var ErrRunNotInSearchCache = errors.New("Run is still being loaded into the searchcache")

// DiffAPI is an abstraction for computing run differences.
type DiffAPI interface {
	GetRunsDiff(before, after TestRun, filter DiffFilterParam, paths mapset.Set) (RunDiff, error)
	GetDiffURL(before, after TestRun, diffFilter *DiffFilterParam) *url.URL
	GetMasterDiffURL(testRun TestRun, diffFilter *DiffFilterParam) *url.URL
}

type diffAPIImpl struct {
	ctx   context.Context
	aeAPI AppEngineAPI
}

// NewDiffAPI return and implementation of the DiffAPI interface.
func NewDiffAPI(ctx context.Context) DiffAPI {
	return diffAPIImpl{
		ctx:   ctx,
		aeAPI: NewAppEngineAPI(ctx),
	}
}

func (d diffAPIImpl) GetDiffURL(before, after TestRun, diffFilter *DiffFilterParam) *url.URL {
	detailsURL, _ := url.Parse(fmt.Sprintf("https://%s/results/", d.aeAPI.GetHostname()))
	query := detailsURL.Query()
	query.Add("run_id", fmt.Sprintf("%v", before.ID))
	query.Add("run_id", fmt.Sprintf("%v", after.ID))
	query.Set("diff", "")
	if diffFilter != nil {
		query.Set("filter", diffFilter.String())
	}
	detailsURL.RawQuery = query.Encode()
	return detailsURL
}

// GetMasterDiffURL returns the diff url for comparing a pr_head run against the most recent
// master run for the same product channel.
func (d diffAPIImpl) GetMasterDiffURL(testRun TestRun, diffFilter *DiffFilterParam) *url.URL {
	runSpec := ProductSpec{}
	runSpec.ProductAtRevision = testRun.ProductAtRevision
	runSpec.Labels = mapset.NewSetWith(PRHeadLabel)

	masterSpec := ProductSpec{}
	masterSpec.BrowserName = testRun.BrowserName
	masterSpec.Labels = mapset.NewSetWith(testRun.Channel(), MasterLabel)

	filter := TestRunFilter{
		Products: ProductSpecs{runSpec, masterSpec},
	}
	diffURL := d.aeAPI.GetResultsURL(filter)
	query := diffURL.Query()
	query.Set("diff", "")
	if diffFilter != nil {
		query.Set("filter", diffFilter.String())
	}
	diffURL.RawQuery = query.Encode()
	return diffURL
}

// RunDiff represents a summary of the differences between 2 runs.
type RunDiff struct {
	// Differences is a map from test-path to an array of
	// [newly-passing, newly-failing, total-delta],
	// where newly-pa
	Before        TestRun           `json:"-"`
	BeforeSummary ResultsSummary    `json:"-"`
	After         TestRun           `json:"-"`
	AfterSummary  ResultsSummary    `json:"-"`
	Differences   ResultsDiff       `json:"diff"`
	Renames       map[string]string `json:"renames"`
}

// TestSummary is a pair of [passing, total] counts for a test file.
type TestSummary []int

// Add adds the other summary counts to this one. Used for summing folders.
func (s TestSummary) Add(other TestSummary) {
	s[0] += other[0]
	s[1] += other[1]
}

// ResultsSummary is a collection of [pass, total] summary pairs, keyed by test.
type ResultsSummary map[string]TestSummary

// Add adds the given summary to the summary for the given path, adding it
// to the map if it's not present already.
func (s ResultsSummary) Add(k string, other TestSummary) {
	if _, ok := s[k]; !ok {
		s[k] = TestSummary{0, 0}
	}
	s[k].Add(other)
}

// TestDiff is an array of differences between 2 tests.
type TestDiff []int

// IsEmpty returns true if the diff is empty (all zeroes)
func (d TestDiff) IsEmpty() bool {
	for _, x := range d {
		if x != 0 {
			return false
		}
	}
	return true
}

const (
	newlyPassingIndex = 0
	newlyFailingIndex = 1
	totalDeltaIndex   = 2
)

// NewlyPassing is the delta/increase in the number of passing tests when comparing before/after.
func (d TestDiff) NewlyPassing() int {
	if d == nil {
		return 0
	}
	return d[newlyPassingIndex]
}

// Regressions is the delta/increase in the number of failing tests when comparing before/after.
func (d TestDiff) Regressions() int {
	if d == nil {
		return 0
	}
	return d[newlyFailingIndex]
}

// TotalDelta is the delta in the number of total subtests when comparing before/after.
func (d TestDiff) TotalDelta() int {
	if d == nil {
		return 0
	}
	return d[totalDeltaIndex]
}

// Add adds the given other TestDiff to this TestDiff's value. Used for summing.
func (d TestDiff) Add(other TestDiff) {
	d[newlyPassingIndex] += other[newlyPassingIndex]
	d[newlyFailingIndex] += other[newlyFailingIndex]
	d[totalDeltaIndex] += other[totalDeltaIndex]
}

// Append the difference between the two given statuses, if any.
func (d TestDiff) Append(before, after TestStatus, filter *DiffFilterParam) {
	if before == TestStatusUnknown {
		if after == TestStatusUnknown || !filter.Added {
			return
		}
		if after.IsPassOrOK() {
			d[newlyPassingIndex]++
		} else {
			d[newlyFailingIndex]++
		}
		return
	}
	if after == TestStatusUnknown {
		if filter.Deleted {
			d[totalDeltaIndex]--
		}
		return
	}
	wasPassing, isPassing := before.IsPassOrOK(), after.IsPassOrOK()
	changed := wasPassing != isPassing
	if !changed || !filter.Changed {
		return
	}
	if wasPassing {
		d[newlyFailingIndex]++
	} else {
		d[newlyPassingIndex]++
	}
}

// NewTestDiff computes the differences between two test-run pass-count summaries,
// namely an array of [passing, total] counts.
func NewTestDiff(before, after []int, filter DiffFilterParam) TestDiff {
	if before == nil {
		if after == nil || !filter.Added {
			return nil
		}
		return TestDiff{
			after[0],
			after[1] - after[0],
			after[1],
		}
	}
	if after == nil {
		// NOTE(lukebjerring): Missing tests are only counted towards changes
		// in the total.
		if !filter.Deleted {
			return nil
		}
		return TestDiff{0, 0, -before[1]}
	}

	delta := before[0] - after[0]
	changed := delta != 0 || before[1] != after[1]
	if (!changed && !filter.Unchanged) || changed && !filter.Changed {
		return nil
	}

	improved, regressed := 0, 0
	if d := after[0] - before[0]; d > 0 {
		improved = d
	}
	failingBefore := before[1] - before[0]
	failingAfter := after[1] - after[0]
	if d := failingAfter - failingBefore; d > 0 {
		regressed = d
	}
	// Changed tests is at most the number of different outcomes,
	// but newly introduced tests should still be counted (e.g. 0/2 => 0/5)
	return TestDiff{
		improved,
		regressed,
		after[1] - before[1],
	}
}

// ResultsDiff is a collection of test diffs, keyed by the test path.
type ResultsDiff map[string]TestDiff

// Add adds the given diff to the TestDiff for the given key, or
// puts it in the map if it's not yet present.
func (r ResultsDiff) Add(k string, diff TestDiff) {
	if _, ok := r[k]; !ok {
		r[k] = TestDiff{0, 0, 0}
	}
	r[k].Add(diff)
}

// Regressions returns the set of test paths for tests that have a regression
// value in their diff. A change is considered a regression when tests that existed
// both before and after have an increase in the number of failing tests has increased,
// which will of course include newly-added tests that are failing.
// Additionally, we flag a decrease in the total number of tests as a regression,
// since that can often indicate a failure in a test's setup.
func (r ResultsDiff) Regressions() mapset.Set {
	regressions := mapset.NewSet()
	if r != nil {
		for test, diff := range r {
			if diff.Regressions() > 0 || diff.TotalDelta() < 0 {
				regressions.Add(test)
			}
		}
	}
	return regressions
}

// FetchRunResultsJSONForParam fetches the results JSON blob for the given [product]@[SHA] param.
func FetchRunResultsJSONForParam(
	ctx context.Context, r *http.Request, param string) (results ResultsSummary, err error) {
	afterDecoded, err := base64.URLEncoding.DecodeString(param)
	if err == nil {
		var run TestRun
		if err = json.Unmarshal([]byte(afterDecoded), &run); err != nil {
			return nil, err
		}
		return FetchRunResultsJSON(ctx, run)
	}
	var spec ProductSpec
	if spec, err = ParseProductSpec(param); err != nil {
		return nil, err
	}
	return FetchRunResultsJSONForSpec(ctx, r, spec)
}

// FetchRunResultsJSONForSpec fetches the result JSON blob for the given spec.
func FetchRunResultsJSONForSpec(
	ctx context.Context, r *http.Request, spec ProductSpec) (results ResultsSummary, err error) {
	var run *TestRun
	if run, err = FetchRunForSpec(ctx, spec); err != nil {
		return nil, err
	} else if run == nil {
		return nil, nil
	}
	return FetchRunResultsJSON(ctx, *run)
}

// FetchRunForSpec loads the wpt.fyi TestRun metadata for the given spec.
func FetchRunForSpec(ctx context.Context, spec ProductSpec) (*TestRun, error) {
	one := 1
	store := NewAppEngineDatastore(ctx, true)
	q := store.TestRunQuery()
	testRuns, err := q.LoadTestRuns(ProductSpecs{spec}, nil, SHAs{spec.Revision}, nil, nil, &one, nil)
	if err != nil {
		return nil, err
	}
	allRuns := testRuns.AllRuns()
	if len(allRuns) == 1 {
		return &allRuns[0], nil
	}
	return nil, nil
}

// FetchRunResultsJSON fetches the results JSON summary for the given test run, but does not include subtests (since
// a full run can span 20k files).
func FetchRunResultsJSON(ctx context.Context, run TestRun) (results ResultsSummary, err error) {
	client := urlfetch.Client(ctx)
	url := strings.TrimSpace(run.ResultsURL)
	var resp *http.Response
	if resp, err = client.Get(url); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s returned HTTP status %d:\n%s", url, resp.StatusCode, string(body))
	}
	if err = json.Unmarshal(body, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetRunsDiff returns a RunDiff for the given runs.
func (d diffAPIImpl) GetRunsDiff(before, after TestRun, filter DiffFilterParam, paths mapset.Set) (diff RunDiff, err error) {
	store := NewAppEngineDatastore(d.ctx, false)
	if IsFeatureEnabled(store, "searchcacheDiffs") {
		return d.getRunsDiffFromSearchCache(before, after, filter, paths)
	}
	beforeJSON, err := FetchRunResultsJSON(d.ctx, before)
	if err != nil {
		return diff, fmt.Errorf("Failed to fetch 'before' results: %s", err.Error())
	}
	afterJSON, err := FetchRunResultsJSON(d.ctx, after)
	if err != nil {
		return diff, fmt.Errorf("Failed to fetch 'after' results: %s", err.Error())
	}

	var renames map[string]string
	if d.aeAPI.IsFeatureEnabled("diffRenames") {
		beforeSHA := before.FullRevisionHash
		// Use HEAD...[sha] for PR results, since PR run results always override the value of 'revision' to the PRs HEAD revision.
		if before.FullRevisionHash == after.FullRevisionHash && before.IsPRBase() {
			beforeSHA = "HEAD"
		}
		renames = getDiffRenames(d.aeAPI, beforeSHA, after.FullRevisionHash)
	}
	return RunDiff{
		Before:        before,
		BeforeSummary: beforeJSON,
		After:         after,
		AfterSummary:  afterJSON,
		Differences:   GetResultsDiff(beforeJSON, afterJSON, filter, paths, renames),
		Renames:       renames,
	}, nil
}

func (d diffAPIImpl) getRunsDiffFromSearchCache(before, after TestRun, filter DiffFilterParam, paths mapset.Set) (diff RunDiff, err error) {
	diffURL, _ := url.Parse(fmt.Sprintf("https://%s/api/search", d.aeAPI.GetVersionedHostname()))
	query := diffURL.Query()
	query.Set("diff", "")
	query.Set("filter", filter.String())
	diffURL.RawQuery = query.Encode()

	type diffBody = struct {
		RunIDs []int64 `json:"run_ids"`
	}
	body, _ := json.Marshal(diffBody{RunIDs: []int64{before.ID, after.ID}})

	client, _ := d.aeAPI.GetSlowHTTPClient(time.Second * 10)
	resp, err := client.Post(diffURL.String(), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return diff, err
	} else if resp.StatusCode == http.StatusUnprocessableEntity {
		return diff, ErrRunNotInSearchCache
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return diff, err
	}

	var scDiff SearchResponse
	err = json.Unmarshal(body, &scDiff)
	if err != nil {
		return diff, err
	}
	return RunDiffFromSearchResponse(d.aeAPI, before, after, scDiff)
}

// RunDiffFromSearchResponse builds a RunDiff from a searchcache response.
func RunDiffFromSearchResponse(aeAPI AppEngineAPI, before, after TestRun, scDiff SearchResponse) (RunDiff, error) {
	differences := make(map[string]TestDiff)
	beforeSummary := make(ResultsSummary)
	afterSummary := make(ResultsSummary)
	for _, t := range scDiff.Results {
		differences[t.Test] = t.Diff
		if len(t.LegacyStatus) > 1 {
			beforeSummary[t.Test] = TestSummary{t.LegacyStatus[0].Passes, t.LegacyStatus[0].Total}
			afterSummary[t.Test] = TestSummary{t.LegacyStatus[1].Passes, t.LegacyStatus[1].Total}
		}
	}

	var renames map[string]string
	if aeAPI.IsFeatureEnabled("diffRenames") {
		beforeSHA := before.FullRevisionHash
		// Use HEAD...[sha] for PR results, since PR run results always override the value of 'revision' to the PRs HEAD revision.
		if before.FullRevisionHash == after.FullRevisionHash && before.IsPRBase() {
			beforeSHA = "HEAD"
		}
		renames = getDiffRenames(aeAPI, beforeSHA, after.FullRevisionHash)
	}

	return RunDiff{
		Before:        before,
		BeforeSummary: beforeSummary,
		After:         after,
		AfterSummary:  afterSummary,
		Differences:   differences,
		Renames:       renames,
	}, nil
}

// GetResultsDiff returns a map of test name to an array of [newly-passing, newly-failing, total-delta], for tests which had
// different results counts in their map (which is test name to array of [count-passed, total]).
func GetResultsDiff(
	before ResultsSummary,
	after ResultsSummary,
	filter DiffFilterParam,
	paths mapset.Set,
	renames map[string]string) map[string]TestDiff {
	diff := make(map[string]TestDiff)
	if filter.Deleted || filter.Changed {
		for test, resultsBefore := range before {
			if renames != nil {
				rename, ok := renames[test]
				if ok {
					test = rename
				}
			}
			if !anyPathMatches(paths, test) {
				continue
			}
			testDiff := NewTestDiff(resultsBefore, after[test], filter)
			if testDiff != nil {
				diff[test] = testDiff
			}
		}
	}
	if filter.Added {
		for test, resultsAfter := range after {
			// Skip 'added' results of a renamed file (handled above).
			if renames != nil {
				renamed := false
				for _, is := range renames {
					if is == test {
						renamed = true
						break
					}
				}
				if renamed {
					continue
				}
			}
			// If it was in the before set, it's already been computed.
			if _, ok := before[test]; ok {
				continue
			} else if !anyPathMatches(paths, test) {
				continue
			}
			testDiff := NewTestDiff(nil, resultsAfter, filter)
			if testDiff != nil {
				diff[test] = testDiff
			}
		}
	}
	return diff
}

func getDiffRenames(aeAPI AppEngineAPI, shaBefore, shaAfter string) map[string]string {
	if shaBefore == shaAfter {
		return nil
	}
	ctx := aeAPI.Context()
	log := GetLogger(ctx)
	githubClient, err := aeAPI.GetGitHubClient()
	if err != nil {
		log.Errorf("Failed to get github client: %s", err.Error())
		return nil
	}
	comparison, _, err := githubClient.Repositories.CompareCommits(ctx, WPTRepoOwner, WPTRepoName, shaBefore, shaAfter)
	if err != nil || comparison == nil {
		log.Errorf("Failed to fetch diff for %s...%s: %s", CropString(shaBefore, 7), CropString(shaAfter, 7), err.Error())
		return nil
	}

	renames := make(map[string]string)
	for _, file := range comparison.Files {
		if file.GetStatus() == "renamed" {
			before, after := file.GetPreviousFilename(), file.GetFilename()
			for was, is := range ExplodePossibleRenames(before, after) {
				renames["/"+was] = "/" + is
			}
		}
	}
	if len(renames) < 1 {
		log.Debugf("No renames for %s...%s", CropString(shaBefore, 7), CropString(shaAfter, 7))
	} else {
		log.Debugf("Found %v renames for %s...%s", len(renames), CropString(shaBefore, 7), CropString(shaAfter, 7))
	}
	return renames
}

func anyPathMatches(paths mapset.Set, testPath string) bool {
	if paths == nil {
		return true
	}
	for path := range paths.Iter() {
		if strings.Index(testPath, path.(string)) == 0 {
			return true
		}
	}
	return false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}
