// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/lukebjerring/go-github/github"
	"golang.org/x/oauth2"
	"google.golang.org/appengine/urlfetch"
)

// DiffAPI is an abstraction for computing run differences.
type DiffAPI interface {
	GetRunsDiff(before, after TestRun, filter DiffFilterParam, paths mapset.Set) (RunDiff, error)
	GetDiffURL(before, after TestRun, diffFilter *DiffFilterParam) *url.URL
	GetMasterDiffURL(sha string, product ProductSpec) *url.URL
}

type diffAPIImpl struct {
	ctx   context.Context
	aeAPI AppEngineAPIImpl
}

// NewDiffAPI return and implementation of the DiffAPI interface.
func NewDiffAPI(ctx context.Context) DiffAPI {
	return diffAPIImpl{
		ctx:   ctx,
		aeAPI: NewAppEngineAPI(ctx),
	}
}

func (d diffAPIImpl) GetDiffURL(before, after TestRun, diffFilter *DiffFilterParam) *url.URL {
	filter := TestRunFilter{}
	filter.Products = ProductSpecs{
		ProductSpec{ProductAtRevision: before.ProductAtRevision},
		ProductSpec{ProductAtRevision: after.ProductAtRevision},
	}
	detailsURL := d.aeAPI.GetResultsURL(filter)
	query := detailsURL.Query()
	query.Set("diff", "")
	if diffFilter != nil {
		query.Set("filter", diffFilter.String())
	}
	detailsURL.RawQuery = query.Encode()
	return detailsURL
}

func (d diffAPIImpl) GetMasterDiffURL(sha string, product ProductSpec) *url.URL {
	filter := TestRunFilter{}
	filter.Products = ProductSpecs{product, product}
	filter.Products[0].Labels = mapset.NewSet("master")
	filter.Products[1].Revision = sha
	detailsURL := d.aeAPI.GetResultsURL(filter)
	query := detailsURL.Query()
	query.Set("diff", "")
	query.Set("filter", DiffFilterParam{
		Added:     true,
		Changed:   true,
		Unchanged: true,
	}.String())
	detailsURL.RawQuery = query.Encode()
	return detailsURL
}

// RunDiff represents a summary of the differences between 2 runs.
type RunDiff struct {
	// Differences is a map from test-path to an array of
	// [newly-passing, newly-failing, total-delta],
	// where newly-pa
	Before        TestRun             `json:"-"`
	BeforeSummary map[string][]int    `json:"-"`
	After         TestRun             `json:"-"`
	AfterSummary  map[string][]int    `json:"-"`
	Differences   map[string]TestDiff `json:"diff"`
	Renames       map[string]string   `json:"renames"`
}

// TestDiff is an array of differences between 2 tests.
type TestDiff []int

// NewlyPassing is the delta/increase in the number of passing tests when comparing before/after.
func (d TestDiff) NewlyPassing() int {
	return d[0]
}

// Regressions is the delta/increase in the number of failing tests when comparing before/after.
func (d TestDiff) Regressions() int {
	return d[1]
}

// TotalDelta is the delta in the number of total subtests when comparing before/after.
func (d TestDiff) TotalDelta() int {
	return d[2]
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

// Regressions returns the set of test paths for tests that have a regression
// value in their diff. A change is considered a regression when tests that existed
// both before and after have an increase in the number of failing tests has increased,
// which will of course include newly-added tests that are failing.
// Additionally, we flag a decrease in the total number of tests as a regression,
// since that can often indicate a failure in a test's setup.
func (r RunDiff) Regressions() mapset.Set {
	regressions := mapset.NewSet()
	for test, diff := range r.Differences {
		if diff.Regressions() > 0 || diff.TotalDelta() < 0 {
			regressions.Add(test)
		}
	}
	return regressions
}

// FetchRunResultsJSONForParam fetches the results JSON blob for the given [product]@[SHA] param.
func FetchRunResultsJSONForParam(
	ctx context.Context, r *http.Request, param string) (results map[string][]int, err error) {
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
	ctx context.Context, r *http.Request, spec ProductSpec) (results map[string][]int, err error) {
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
	testRuns, err := LoadTestRuns(ctx, []ProductSpec{spec}, nil, spec.Revision, nil, nil, &one, nil)
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
func FetchRunResultsJSON(ctx context.Context, run TestRun) (results map[string][]int, err error) {
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
	beforeJSON, err := FetchRunResultsJSON(d.ctx, before)
	if err != nil {
		return diff, fmt.Errorf("Failed to fetch 'before' results: %s", err.Error())
	}
	afterJSON, err := FetchRunResultsJSON(d.ctx, after)
	if err != nil {
		return diff, fmt.Errorf("Failed to fetch 'after' results: %s", err.Error())
	}

	var renames map[string]string
	if IsFeatureEnabled(d.ctx, "diffRenames") {
		renames = getDiffRenames(d.ctx, before.FullRevisionHash, after.FullRevisionHash)
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

// GetResultsDiff returns a map of test name to an array of [newly-passing, newly-failing, total-delta], for tests which had
// different results counts in their map (which is test name to array of [count-passed, total]).
func GetResultsDiff(
	before map[string][]int,
	after map[string][]int,
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

func getDiffRenames(ctx context.Context, shaBefore, shaAfter string) map[string]string {
	if shaBefore == shaAfter {
		return nil
	}
	log := GetLogger(ctx)
	secret, err := GetSecret(ctx, "github-api-token")
	if err != nil {
		log.Debugf("Failed to load github-api-token: %s", err.Error())
		return nil
	}
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: secret,
	}))
	githubClient := github.NewClient(oauthClient)
	comparison, _, err := githubClient.Repositories.CompareCommits(ctx, "web-platform-tests", "wpt", shaBefore, shaAfter)
	if err != nil || comparison == nil {
		log.Errorf("Failed to fetch diff for %s...%s: %s", shaBefore[:7], shaAfter[:7], err.Error())
		return nil
	}

	renames := make(map[string]string)
	for _, file := range comparison.Files {
		if file.GetStatus() == "renamed" {
			is, was := file.GetFilename(), file.GetPreviousFilename()
			renames["/"+was] = "/" + is
		}
	}
	if len(renames) < 1 {
		log.Debugf("No renames for %s...%s", shaBefore[:7], shaAfter[:7])
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
