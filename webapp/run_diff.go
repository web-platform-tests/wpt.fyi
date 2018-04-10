// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	mapset "github.com/deckarep/golang-set"
	models "github.com/w3c/wptdashboard/shared"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
)

type platformAtRevision struct {
	// Platform is the string representing browser (+ version), and OS (+ version).
	Platform string

	// Revision is the SHA[0:10] of the git repo.
	Revision string
}

func parsePlatformAtRevisionSpec(spec string) (platformAtRevision platformAtRevision, err error) {
	pieces := strings.Split(spec, "@")
	if len(pieces) > 2 {
		return platformAtRevision, errors.New("invalid platform@revision spec: " + spec)
	}
	platformAtRevision.Platform = pieces[0]
	if len(pieces) < 2 {
		// No @ is assumed to be the platform only.
		platformAtRevision.Revision = "latest"
	} else {
		platformAtRevision.Revision = pieces[1]
	}
	// TODO(lukebjerring): Also handle actual platforms (with version + os)
	if IsBrowserName(platformAtRevision.Platform) {
		return platformAtRevision, nil
	}
	return platformAtRevision, errors.New("Platform " + platformAtRevision.Platform + " not found")
}

func fetchRunResultsJSONForParam(
	ctx context.Context, r *http.Request, param string) (results map[string][]int, err error) {
	afterDecoded, err := base64.URLEncoding.DecodeString(param)
	if err == nil {
		var run models.TestRun
		if err = json.Unmarshal([]byte(afterDecoded), &run); err != nil {
			return nil, err
		}
		return fetchRunResultsJSON(ctx, r, run)
	}
	var spec platformAtRevision
	if spec, err = parsePlatformAtRevisionSpec(param); err != nil {
		return nil, err
	}
	return fetchRunResultsJSONForSpec(ctx, r, spec)
}

func fetchRunResultsJSONForSpec(
	ctx context.Context, r *http.Request, revision platformAtRevision) (results map[string][]int, err error) {
	var run models.TestRun
	if run, err = fetchRunForSpec(ctx, revision); err != nil {
		return nil, err
	} else if (run == models.TestRun{}) {
		return nil, nil
	}
	return fetchRunResultsJSON(ctx, r, run)
}

func fetchRunForSpec(ctx context.Context, revision platformAtRevision) (models.TestRun, error) {
	baseQuery := datastore.
		NewQuery("TestRun").
		Order("-CreatedAt").
		Limit(1)

	var results []models.TestRun
	// TODO(lukebjerring): Handle actual platforms (split out version + os)
	query := baseQuery.
		Filter("BrowserName =", revision.Platform)
	if revision.Revision != "latest" {
		query = query.Filter("Revision = ", revision.Revision)
	}
	if _, err := query.GetAll(ctx, &results); err != nil {
		return models.TestRun{}, err
	}
	if len(results) < 1 {
		return models.TestRun{}, nil
	}
	return results[0], nil
}

// fetchRunResultsJSON fetches the results JSON summary for the given test run, but does not include subtests (since
// a full run can span 20k files).
func fetchRunResultsJSON(ctx context.Context, r *http.Request, run models.TestRun) (results map[string][]int, err error) {
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

// getResultsDiff returns a map of test name to an array of [count-different-tests, total-tests], for tests which had
// different results counts in their map (which is test name to array of [count-passed, total-tests]).
//
func getResultsDiff(before map[string][]int, after map[string][]int, filter DiffFilterParam) map[string][]int {
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
