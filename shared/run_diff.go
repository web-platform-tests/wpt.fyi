// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

// FetchRunResultsJSONForParam fetches the results JSON blob for the given [product]@[SHA] param.
func FetchRunResultsJSONForParam(
	ctx context.Context, r *http.Request, param string) (results map[string][]int, err error) {
	afterDecoded, err := base64.URLEncoding.DecodeString(param)
	if err == nil {
		var run TestRun
		if err = json.Unmarshal([]byte(afterDecoded), &run); err != nil {
			return nil, err
		}
		return FetchRunResultsJSON(ctx, r, run)
	}
	var spec ProductAtRevision
	if spec, err = ParseProductAtRevision(param); err != nil {
		return nil, err
	}
	return FetchRunResultsJSONForSpec(ctx, r, spec)
}

// FetchRunResultsJSONForSpec fetches the result JSON blob for the given spec.
func FetchRunResultsJSONForSpec(
	ctx context.Context, r *http.Request, revision ProductAtRevision) (results map[string][]int, err error) {
	var run *TestRun
	if run, err = FetchRunForSpec(ctx, revision); err != nil {
		return nil, err
	} else if run == nil {
		return nil, nil
	}
	return FetchRunResultsJSON(ctx, r, *run)
}

// FetchRunForSpec loads the wpt.fyi TestRun metadata for the given spec.
func FetchRunForSpec(ctx context.Context, revision ProductAtRevision) (*TestRun, error) {
	testRuns, err := LoadTestRuns(ctx, []Product{revision.Product}, revision.Revision, nil, 1)
	if err != nil {
		return nil, err
	}
	if len(testRuns) < 1 {
		return nil, nil
	}
	return &testRuns[0], nil
}

// FetchRunResultsJSON fetches the results JSON summary for the given test run, but does not include subtests (since
// a full run can span 20k files).
func FetchRunResultsJSON(ctx context.Context, r *http.Request, run TestRun) (results map[string][]int, err error) {
	client := urlfetch.Client(ctx)
	url := strings.TrimSpace(run.ResultsURL)
	if strings.Index(url, "/") == 0 {
		reqURL := *r.URL
		reqURL.Path = url
	}
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

// GetResultsDiff returns a map of test name to an array of [count-different-tests, total-tests], for tests which had
// different results counts in their map (which is test name to array of [count-passed, total-tests]).
//
func GetResultsDiff(before map[string][]int, after map[string][]int, filter DiffFilterParam) map[string][]int {
	diff := make(map[string][]int)
	if filter.Deleted || filter.Changed {
		for test, resultsBefore := range before {
			if !anyPathMatches(filter.Paths, test) {
				continue
			}

			if resultsAfter, ok := after[test]; !ok {
				// Missing? Then N / N tests are 'different'.
				if !filter.Deleted {
					continue
				}
				diff[test] = []int{resultsBefore[1], resultsBefore[1]}
			} else {
				if !filter.Changed && !filter.Unchanged {
					continue
				}
				passDiff := abs(resultsBefore[0] - resultsAfter[0])
				countDiff := abs(resultsBefore[1] - resultsAfter[1])
				changed := passDiff != 0 || countDiff != 0
				if (!changed && !filter.Unchanged) || changed && !filter.Changed {
					continue
				}
				// Changed tests is at most the number of different outcomes,
				// but newly introduced tests should still be counted (e.g. 0/2 => 0/5)
				diff[test] = []int{
					max(passDiff, countDiff),
					max(resultsBefore[1], resultsAfter[1]),
				}
			}
		}
	}
	if filter.Added {
		for test, resultsAfter := range after {
			if !anyPathMatches(filter.Paths, test) {
				continue
			}

			if _, ok := before[test]; !ok {
				// Missing? Then N / N tests are 'different'
				diff[test] = []int{resultsAfter[1], resultsAfter[1]}
			}
		}
	}
	return diff
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
