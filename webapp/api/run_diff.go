// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	mapset "github.com/deckarep/golang-set"
	models "github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
)

// apiDiffHandler takes 2 test-run results JSON blobs and produces JSON in the same format, with only the differences
// between runs.
//
// GET takes before and after params, for historical production runs.
// POST takes only a before param, and the after state is provided in the body of the POST request.
func apiDiffHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleAPIDiffGet(w, r)
	case "POST":
		handleAPIDiffPost(w, r)
	default:
		http.Error(w, fmt.Sprintf("invalid HTTP method %s", r.Method), http.StatusBadRequest)
	}
}

func handleAPIDiffGet(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	var err error
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	specBefore := params.Get("before")
	if specBefore == "" {
		http.Error(w, "before param missing", http.StatusBadRequest)
		return
	}
	var beforeJSON map[string][]int
	if beforeJSON, err = fetchRunResultsJSONForParam(ctx, r, specBefore); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if beforeJSON == nil {
		http.Error(w, specBefore+" not found", http.StatusNotFound)
		return
	}

	specAfter := params.Get("after")
	if specAfter == "" {
		http.Error(w, "after param missing", http.StatusBadRequest)
		return
	}
	var afterJSON map[string][]int
	if afterJSON, err = fetchRunResultsJSONForParam(ctx, r, specAfter); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if afterJSON == nil {
		http.Error(w, specAfter+" not found", http.StatusNotFound)
		return
	}

	var filter DiffFilterParam
	if filter, err = ParseDiffFilterParams(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	diffJSON := getResultsDiff(beforeJSON, afterJSON, filter)
	var bytes []byte
	if bytes, err = json.Marshal(diffJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

// handleAPIDiffPost handles POST requests to /api/diff, which allows the caller to produce the diff of an arbitrary
// run result JSON blob against a historical production run.
func handleAPIDiffPost(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	var err error
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	specBefore := params.Get("before")
	if specBefore == "" {
		http.Error(w, "before param missing", http.StatusBadRequest)
		return
	}
	var beforeJSON map[string][]int
	if beforeJSON, err = fetchRunResultsJSONForParam(ctx, r, specBefore); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if beforeJSON == nil {
		http.Error(w, specBefore+" not found", http.StatusNotFound)
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var afterJSON map[string][]int
	if err = json.Unmarshal(body, &afterJSON); err != nil {
		http.Error(w, "Failed to parse JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	var filter DiffFilterParam
	if filter, err = ParseDiffFilterParams(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	diffJSON := getResultsDiff(beforeJSON, afterJSON, filter)
	var bytes []byte
	if bytes, err = json.Marshal(diffJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

// PlatformAtRevision is represents a test-run spec, of [platform]@[SHA],
// e.g. 'chrome@latest' or 'safari-10@abcdef1234'
type PlatformAtRevision struct {
	// Platform is the string representing browser (+ version), and OS (+ version).
	Platform string

	// Revision is the SHA[0:10] of the git repo.
	Revision string
}

// ParsePlatformAtRevisionSpec parses a test-run spec into a PlatformAtRevision struct.
func ParsePlatformAtRevisionSpec(spec string) (platformAtRevision PlatformAtRevision, err error) {
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
	var spec PlatformAtRevision
	if spec, err = ParsePlatformAtRevisionSpec(param); err != nil {
		return nil, err
	}
	return fetchRunResultsJSONForSpec(ctx, r, spec)
}

func fetchRunResultsJSONForSpec(
	ctx context.Context, r *http.Request, revision PlatformAtRevision) (results map[string][]int, err error) {
	var run models.TestRun
	if run, err = fetchRunForSpec(ctx, revision); err != nil {
		return nil, err
	} else if (run == models.TestRun{}) {
		return nil, nil
	}
	return fetchRunResultsJSON(ctx, r, run)
}

func fetchRunForSpec(ctx context.Context, revision PlatformAtRevision) (models.TestRun, error) {
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
