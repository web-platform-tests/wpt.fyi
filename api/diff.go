package api //nolint:revive

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// apiDiffHandler takes 2 test-run results JSON blobs and produces JSON in the same format, with only the differences
// between runs.
//
// GET takes before and after params, for historical production runs.
// POST takes only a before param, and the after state is provided in the body of the POST request.
func apiDiffHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleAPIDiffGet(w, r)
	case http.MethodPost:
		handleAPIDiffPost(w, r)
	default:
		http.Error(w, fmt.Sprintf("invalid HTTP method %s", r.Method), http.StatusBadRequest)
	}
}

func loadDiffRuns(store shared.Datastore, q url.Values) (shared.TestRuns, error) {
	// nolint:nestif // TODO: Fix nestif lint error
	if runIDs, err := shared.ParseRunIDsParam(q); err != nil {
		return nil, err
	} else if len(runIDs) > 0 {
		runs, err := runIDs.LoadTestRuns(store)
		// If all errors are NoSuchEntity, we don't treat it as an error.
		// If err is nil, the type conversion will fail.
		var multiError *shared.MultiError
		if errors.As(err, &multiError) {
			all404s := true
			for _, err := range multiError.Errors() {
				if !errors.Is(err, shared.ErrNoSuchEntity) {
					all404s = false

					break
				}
			}
			if all404s {
				return nil, nil
			}
		}
		if err != nil {
			return nil, err
		}

		return runs, nil
	}

	// NOTE: We use the same params as /results, but also support
	// 'before' and 'after' and 'filter'.
	runFilter, err := shared.ParseTestRunFilterParams(q)
	if err != nil {
		return nil, err
	}
	if beforeAndAfter, err := shared.ParseBeforeAndAfterParams(q); err != nil {
		return nil, err
	} else if len(beforeAndAfter) > 0 {
		runFilter.Products = beforeAndAfter
	}
	runsByProduct, err := LoadTestRunsForFilters(store, runFilter)
	if err != nil {
		return nil, err
	}

	return runsByProduct.AllRuns(), nil
}

func handleAPIDiffGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	store := shared.NewAppEngineDatastore(ctx, true)
	q := r.URL.Query()

	diffFilter, paths, err := shared.ParseDiffFilterParams(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	runs, err := loadDiffRuns(store, q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}
	if runs == nil {
		http.NotFound(w, r)

		return
	}

	if len(runs) != 2 {
		http.Error(w, fmt.Sprintf("Diffing requires exactly 2 runs, but found %v", len(runs)), http.StatusBadRequest)

		return
	}

	diffAPI := shared.NewDiffAPI(ctx)
	diff, err := diffAPI.GetRunsDiff(runs[0], runs[1], diffFilter, paths)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var bytes []byte
	if bytes, err = json.Marshal(diff); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	_, err = w.Write(bytes)
	// nolint:godox // TODO: Golangci-lint found that we previously ignored the error.
	// We should investigate if we should return a HTTP error or not. In the meantime, we log the error.
	if err != nil {
		logger := shared.GetLogger(r.Context())
		logger.Warningf("Failed to write data in api/diff GET handler: %s", err.Error())
	}
}

// handleAPIDiffPost handles POST requests to /api/diff, which allows the caller to produce the diff of an arbitrary
// run result JSON blob against a historical production run.
func handleAPIDiffPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
	var beforeJSON shared.ResultsSummary
	if beforeJSON, err = shared.FetchRunResultsJSONForParam(ctx, r, specBefore); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	} else if beforeJSON == nil {
		http.Error(w, specBefore+" not found", http.StatusNotFound)

		return
	}

	var body []byte
	if body, err = io.ReadAll(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var afterJSON shared.ResultsSummary
	if err := json.Unmarshal(body, &afterJSON); err != nil {
		http.Error(w, "Failed to parse JSON: "+err.Error(), http.StatusBadRequest)

		return
	}

	var filter shared.DiffFilterParam
	var paths mapset.Set
	if filter, paths, err = shared.ParseDiffFilterParams(params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	diffJSON := shared.GetResultsDiff(beforeJSON, afterJSON, filter, paths, nil)
	var bytes []byte
	if bytes, err = json.Marshal(diffJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		logger := shared.GetLogger(r.Context())
		logger.Warningf("Failed to write data in api/diff POST handler: %s", err.Error())
	}
}
