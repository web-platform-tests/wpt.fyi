package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/shared"
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

	diffJSON := shared.GetResultsDiff(beforeJSON, afterJSON, diffFilter, paths)
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
