package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/web-platform-tests/wpt.fyi/shared"
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

	// /results has the same params (before + after), so we re-use the logic there.
	runFilter, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var diffFilter shared.DiffFilterParam
	if diffFilter, err = shared.ParseDiffFilterParams(r); err != nil {
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
	if len(runFilter.Products) != 2 {
		http.Error(w, fmt.Sprintf("Diffing requires before/after, or exactly 2 products, but found %v", len(runFilter.Products)), http.StatusBadRequest)
		return
	}
	var runs shared.TestRuns
	if runs, err = LoadTestRunsForFilters(ctx, runFilter); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	diffJSON := shared.GetResultsDiff(beforeJSON, afterJSON, diffFilter)
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
	if filter, err = shared.ParseDiffFilterParams(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	diffJSON := shared.GetResultsDiff(beforeJSON, afterJSON, filter)
	var bytes []byte
	if bytes, err = json.Marshal(diffJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
