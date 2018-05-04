// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	models "github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
)

const experimentalLabel = `experimental`

// apiTestRunsHandler is responsible for emitting test-run JSON for all the runs at a given SHA.
//
// URL Params:
//     sha: SHA[0:10] of the repo when the tests were executed (or 'latest')
func apiTestRunsHandler(w http.ResponseWriter, r *http.Request) {
	runSHA, err := ParseSHAParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)
	// When ?complete=true, make sure to show results for the same complete run (executed for all browsers).
	if complete, err := strconv.ParseBool(r.URL.Query().Get("complete")); err == nil && complete {
		if runSHA == "latest" {
			runSHA, err = getLastCompleteRunSHA(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	var browserNames []string
	if browserNames, err = ParseBrowsersParam(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	labels := ParseLabelsParam(r)
	experimentalBrowsers := labels != nil && labels.Contains(experimentalLabel)

	var testRuns []models.TestRun
	var limit int
	if limit, err = ParseMaxCountParam(r); err != nil {
		http.Error(w, "Invalid 'max-count' param: "+err.Error(), http.StatusBadRequest)
		return
	}
	baseQuery := datastore.
		NewQuery("TestRun").
		Order("-CreatedAt").
		Limit(limit)

	for _, browserName := range browserNames {
		var testRunResults []models.TestRun
		if experimentalBrowsers {
			browserName = browserName + "-" + experimentalLabel
		}
		query := baseQuery.Filter("BrowserName =", browserName)
		if runSHA != "" && runSHA != "latest" {
			query = query.Filter("Revision =", runSHA)
		}
		if _, err := query.GetAll(ctx, &testRunResults); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		testRuns = append(testRuns, testRunResults...)
	}

	testRunsBytes, err := json.Marshal(testRuns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(testRunsBytes)
}

func apiTestRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		apiTestRunPostHandler(w, r)
	} else if r.Method == "GET" {
		apiTestRunGetHandler(w, r)
	} else {
		http.Error(w, "This endpoint only supports GET and POST.", http.StatusMethodNotAllowed)
	}
}

// apiTestRunGetHandler is responsible for emitting the test-run JSON a specific run,
// identified by a named browser (platform) at a given SHA.
//
// URL Params:
//     sha: SHA[0:10] of the repo when the test was executed (or 'latest')
//     browser: Browser for the run (e.g. 'chrome', 'safari-10')
func apiTestRunGetHandler(w http.ResponseWriter, r *http.Request) {
	runSHA, err := ParseSHAParam(r)
	if err != nil {
		http.Error(w, "Invalid query params", http.StatusBadRequest)
		return
	}

	var browserName string
	browserName, err = ParseBrowserParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if browserName == "" {
		http.Error(w, "Invalid 'browser' param", http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)

	query := datastore.
		NewQuery("TestRun").
		Order("-CreatedAt").
		Limit(1).
		Filter("BrowserName =", browserName)
	if runSHA != "" && runSHA != "latest" {
		query = query.Filter("Revision =", runSHA)
	}

	var testRuns []models.TestRun
	if _, err := query.GetAll(ctx, &testRuns); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(testRuns) == 0 {
		http.NotFound(w, r)
		return
	}

	testRunsBytes, err := json.Marshal(testRuns[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(testRunsBytes)
}

// apiTestRunPostHandler is responsible for handling TestRun submissions (via HTTP POST requests).
// It asserts the presence of a required secret token, then saves the JSON blob to the Datastore.
// See models.go for the JSON format expected.
func apiTestRunPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var err error

	// Fetch pre-uploaded models.Token entity.
	suppliedSecret := r.URL.Query().Get("secret")
	tokenKey := datastore.NewKey(ctx, "Token", "upload-token", 0, nil)
	var token models.Token
	if err = datastore.Get(ctx, tokenKey, &token); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if suppliedSecret != token.Secret {
		http.Error(w, fmt.Sprintf("Invalid token '%s'", suppliedSecret), http.StatusUnauthorized)
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var testRun models.TestRun
	if err = json.Unmarshal(body, &testRun); err != nil {
		http.Error(w, "Failed to parse JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Use 'now' as created time, unless flagged as retroactive.
	if retro, err := strconv.ParseBool(r.URL.Query().Get("retroactive")); err != nil || !retro {
		testRun.CreatedAt = time.Now()
	}

	// Create a new models.TestRun out of the JSON body of the request.
	key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
	if _, err := datastore.Put(ctx, key, &testRun); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var jsonOutput []byte
	if jsonOutput, err = json.Marshal(testRun); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonOutput)
	w.WriteHeader(http.StatusCreated)
}

// getLastCompleteRunSHA returns the SHA[0:10] for the most recent run that exists for all initially-loaded browser
// names (see GetBrowserNames).
func getLastCompleteRunSHA(ctx context.Context) (sha string, err error) {
	baseQuery := datastore.
		NewQuery("TestRun").
		Order("-CreatedAt").
		Limit(100).
		Project("Revision")

	// Map is sha -> browser -> seen yet?  - this prevents over-counting dupes.
	runSHAs := make(map[string]map[string]bool)
	var browserNames []string
	if browserNames, err = GetBrowserNames(); err != nil {
		return sha, err
	}

	for _, browser := range browserNames {
		it := baseQuery.Filter("BrowserName = ", browser).Run(ctx)
		for {
			var testRun models.TestRun
			_, err := it.Next(&testRun)
			if err == datastore.Done {
				break
			}
			if err != nil {
				return "latest", err
			}
			if _, ok := runSHAs[testRun.Revision]; !ok {
				runSHAs[testRun.Revision] = make(map[string]bool)
			}
			browsersSeen := runSHAs[testRun.Revision]
			browsersSeen[browser] = true
			if len(browsersSeen) == len(browserNames) {
				return testRun.Revision, nil
			}
		}
	}
	return "latest", nil
}

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

func apiManifestHandler(w http.ResponseWriter, r *http.Request) {
	sha, err := ParseSHAParamFull(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(r)
	if manifest, err := getManifestForSHA(ctx, sha); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else {
		w.Header().Add("content-type", "application/json")
		w.Write(manifest)
	}
}

type gitHubClient interface {
	fetch(url string) ([]byte, error)
}

func getManifestForSHA(ctx context.Context, sha string) (manifest []byte, err error) {
	// Fetch models.Token entity for GitHub API Token.
	tokenKey := datastore.NewKey(ctx, "Token", "github-api-token", 0, nil)
	var token models.Token
	datastore.Get(ctx, tokenKey, &token)

	client := gitHubClientImpl{
		Token:   &token,
		Context: ctx,
	}
	return getGitHubReleaseAsset(&client, sha)
}

func getGitHubReleaseAsset(client gitHubClient, sha string) (manifest []byte, err error) {
	// Search for the PR associated with the SHA.
	const githubSearch = `https://api.github.com/search/issues?q=SHA:%s+user:w3c+repo:web-platform-tests`
	url := fmt.Sprintf(githubSearch, sha)
	var body []byte
	if body, err = client.fetch(url); err != nil {
		return nil, err
	}

	var queryResults map[string]*json.RawMessage
	if err = json.Unmarshal(body, &queryResults); err != nil {
		return nil, err
	}
	var issues []map[string]*json.RawMessage
	if err = json.Unmarshal(*queryResults["items"], &issues); err != nil {
		return nil, err
	}
	if len(issues) < 1 {
		return nil, fmt.Errorf("No search results found for SHA %s", sha)
	}

	// Load the release by the presumed tag name merge_pr_*
	var prNumber int
	if err = json.Unmarshal(*issues[0]["number"], &prNumber); err != nil {
		return nil, err
	}

	releaseTag := fmt.Sprintf("merge_pr_%d", prNumber)
	const githubRelease = `https://api.github.com/repos/w3c/web-platform-tests/releases/tags/%s`
	url = fmt.Sprintf(githubRelease, releaseTag)
	if body, err = client.fetch(url); err != nil {
		return nil, err
	}

	var release map[string]*json.RawMessage
	if err = json.Unmarshal(body, &release); err != nil {
		return nil, err
	}
	var assets []map[string]*json.RawMessage
	if err = json.Unmarshal(*release["assets"], &assets); err != nil {
		return nil, err
	}
	if len(assets) < 1 {
		return nil, fmt.Errorf("No assets found for release %s", releaseTag)
	}
	// Get (and unzip) the asset with name "MANIFEST-{sha}.json.gz"
	for _, asset := range assets {
		var name string
		if err = json.Unmarshal(*asset["name"], &name); err != nil {
			return nil, err
		}
		if strings.Contains(name, sha) {
			if err = json.Unmarshal(*asset["browser_download_url"], &url); err != nil {
				return nil, err
			}

			if body, err = client.fetch(url); err != nil {
				return nil, err
			}
			gzReader, err := gzip.NewReader(bytes.NewReader(body))
			if err != nil {
				return nil, err
			}
			if body, err = ioutil.ReadAll(gzReader); err != nil {
				return nil, err
			}
			return body, nil
		}
	}
	return nil, fmt.Errorf("No manifest asset found for release %s", releaseTag)
}

type gitHubClientImpl struct {
	Token   *models.Token
	Context context.Context
}

func (g *gitHubClientImpl) fetch(url string) ([]byte, error) {
	client := urlfetch.Client(g.Context)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != nil {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", g.Token.Secret))
	}
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
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
	return body, nil
}
