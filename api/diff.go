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
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
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
	Renames map[string]string `json:"renames,omitempty"`
}

func handleAPIDiffGet(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)

	runIDs, err := shared.ParseRunIDsParam(r)
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

	var runs shared.TestRuns
	if len(runIDs) > 0 {
		runs, err = runIDs.LoadTestRuns(ctx)
		if err != nil {
			if multiError, ok := err.(appengine.MultiError); ok {
				all404s := true
				for _, err := range multiError {
					if err != datastore.ErrNoSuchEntity {
						all404s = false
					}
				}
				if all404s {
					http.NotFound(w, r)
					return
				}
			}
		}
	} else {
		// NOTE: We use the same params as /results, but also support
		// 'before' and 'after' and 'filter'.
		runFilter, parseErr := shared.ParseTestRunFilterParams(r)
		if parseErr != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		beforeAndAfter, parseErr := shared.ParseBeforeAndAfterParams(r)
		if parseErr != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(beforeAndAfter) > 0 {
			runFilter.Products = beforeAndAfter
		}
		runs, err = LoadTestRunsForFilters(ctx, runFilter)
	}

	if err != nil {
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

	var renames map[string]string
	if shared.IsFeatureEnabled(ctx, "diffRenames") {
		renames = getDiffRenames(ctx, runs[0].FullRevisionHash, runs[1].FullRevisionHash)
	}
	diff := diffResult{
		Diff: shared.GetResultsDiff(beforeJSON, afterJSON, diffFilter, paths, renames),
	}
	diff.Renames = renames

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

	diffJSON := shared.GetResultsDiff(beforeJSON, afterJSON, filter, paths, nil)
	var bytes []byte
	if bytes, err = json.Marshal(diffJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func getDiffRenames(ctx context.Context, shaBefore, shaAfter string) map[string]string {
	if shaBefore == shaAfter {
		return nil
	}
	log := shared.GetLogger(ctx)
	secret, err := shared.GetSecret(ctx, "github-api-token")
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
