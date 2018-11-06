package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	mapset "github.com/deckarep/golang-set"
	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"

	"golang.org/x/oauth2"
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

type diffResult struct {
	Diff    map[string][]int  `json:"diff"`
	Renames map[string]string `json:"renames"`
}

func handleAPIDiffGet(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)

	// NOTE: We use the same params as /results, but also support
	// 'before' and 'after' and 'filter'.
	runFilter, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var diffFilter shared.DiffFilterParam
	var paths mapset.Set
	if diffFilter, paths, err = shared.ParseDiffFilterParams(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var beforeAndAfter shared.ProductSpecs
	beforeAndAfter, err = shared.ParseBeforeAndAfterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(beforeAndAfter) > 0 {
		runFilter.Products = beforeAndAfter
	}
	var runs shared.TestRuns
	if runs, err = LoadTestRunsForFilters(ctx, runFilter); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(runs) != 2 {
		http.Error(w, fmt.Sprintf("Diffing requires exactly 2 runs, but found %v", len(runs)), http.StatusBadRequest)
		return
	}

	beforeJSON, err := shared.FetchRunResultsJSON(ctx, r, runs[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch 'before' results: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	afterJSON, err := shared.FetchRunResultsJSON(ctx, r, runs[1])
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch 'after' results: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	diff := diffResult{
		Diff: shared.GetResultsDiff(beforeJSON, afterJSON, diffFilter, paths),
	}
	if shared.IsFeatureEnabled(ctx, "diffRenames") {
		diff.Renames = getDiffRenames(ctx, runs[0].FullRevisionHash, runs[1].FullRevisionHash)
	}
	var bytes []byte
	if bytes, err = json.Marshal(diff); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

// handleAPIDiffPost handles POST requests to /api/diff, which allows the caller to produce the diff of an arbitrary
// run result JSON blob against a historical production run.
func handleAPIDiffPost(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)

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
	if beforeJSON, err = shared.FetchRunResultsJSONForParam(ctx, r, specBefore); err != nil {
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

	var filter shared.DiffFilterParam
	var paths mapset.Set
	if filter, paths, err = shared.ParseDiffFilterParams(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	diffJSON := shared.GetResultsDiff(beforeJSON, afterJSON, filter, paths)
	var bytes []byte
	if bytes, err = json.Marshal(diffJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

// Wrappers for including a missing field in go-github
type githubCommitFile struct {
	github.CommitFile
	PreviousFilename *string `json:"previous_filename,omitempty"`
}

type githubCommitsComparison struct {
	github.CommitsComparison
	Files []githubCommitFile
}

func getDiffRenames(ctx context.Context, shaBefore, shaAfter string) map[string]string {
	log := shared.GetLogger(ctx)
	secret, err := shared.GetSecret(ctx, "github-api-token")
	if err != nil {
		log.Debugf("Failed to load github-api-token: %s", err.Error())
		return nil
	}
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: secret,
	}))
	comparison, err := compareCommits(oauthClient, "web-platform-tests", "wpt", shaBefore, shaAfter)
	if err != nil || comparison == nil {
		log.Errorf("Failed to fetch diff for %s...%s: %s", shaBefore[:7], shaAfter[:7], err.Error())
		return nil
	}

	renames := make(map[string]string)
	for _, file := range comparison.Files {
		if file.Status != nil &&
			*file.Status == "rename" &&
			file.Filename != nil &&
			file.PreviousFilename != nil {
			renames["/"+*file.PreviousFilename] = "/" + *file.Filename
		}
	}
	return renames
}

func compareCommits(client *http.Client, owner, repo string, base, head string) (*githubCommitsComparison, error) {
	u := fmt.Sprintf("https://api.github.com/repos/%v/%v/compare/%v...%v", owner, repo, base, head)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	comp := new(githubCommitsComparison)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var bytes []byte
	var comparison githubCommitsComparison
	bytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &comparison)
	if err != nil {
		return nil, err
	}

	return comp, nil
}
