// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	cq "github.com/web-platform-tests/wpt.fyi/api/query/cache/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

type searchError struct {
	// Detail is the internal error that should not be exposed to the end-user.
	Detail error
	// Message is the user-facing error message.
	Message string
	// Code is the HTTP status code for this error.
	Code int
}

func (e searchError) Error() string {
	if e.Detail == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %v", e.Message, e.Detail)
}

// nolint:gocognit // TODO: Fix gocognit lint error
func searchHandlerImpl(w http.ResponseWriter, r *http.Request) *searchError {
	ctx := r.Context()
	log := shared.GetLogger(ctx)
	if r.Method != http.MethodPost {
		return &searchError{ // nolint:exhaustruct // TODO: Fix exhaustruct lint error.
			Message: "Invalid HTTP method " + r.Method,
			Code:    http.StatusBadRequest,
		}
	}

	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		return &searchError{
			Detail:  err,
			Message: "Failed to read request body",
			Code:    http.StatusInternalServerError,
		}
	}
	log.Debugf("%s", string(reqData))
	if err := r.Body.Close(); err != nil {
		return &searchError{
			Detail:  err,
			Message: "Failed to close request body",
			Code:    http.StatusInternalServerError,
		}
	}

	var rq query.RunQuery
	if err := json.Unmarshal(reqData, &rq); err != nil {
		return &searchError{
			Detail:  err,
			Message: "Failed to unmarshal request body",
			Code:    http.StatusBadRequest,
		}
	}

	if len(rq.RunIDs) > *maxRunsPerRequest {
		return &searchError{ // nolint:exhaustruct // TODO: Fix exhaustruct lint error.
			Message: maxRunsPerRequestMsg,
			Code:    http.StatusBadRequest,
		}
	}

	// Ensure runs are loaded before executing query. This is best effort: It is
	// possible, though unlikely, that a run may exist in the cache at this point
	// and be evicted before binding the query to a query execution plan. In such
	// a case, `idx.Bind()` below will return an error.
	//
	// Accumulate missing runs in `missing` to report which runs have initiated
	// write-on-read. Return to client `http.StatusUnprocessableEntity`
	// immediately if any runs are missing.
	//
	// `ids` and `runs` tracks run IDs and run metadata for requested runs that
	// are currently resident in `idx`.
	store, err := getDatastore(ctx)
	if err != nil {
		return &searchError{
			Detail:  err,
			Message: "Failed to open Datastore",
			Code:    http.StatusInternalServerError,
		}
	}

	ids := make([]int64, 0, len(rq.RunIDs))
	runs := make([]shared.TestRun, 0, len(rq.RunIDs))
	missing := make([]shared.TestRun, 0, len(rq.RunIDs))
	for i := range rq.RunIDs {
		id := index.RunID(rq.RunIDs[i])
		run, err := idx.Run(id)
		// If getting run metadata fails, attempt write-on-read for this run.
		if err != nil {
			runPtr := new(shared.TestRun)
			if err := store.Get(store.NewIDKey("TestRun", int64(id)), runPtr); err != nil {
				return &searchError{
					Detail:  err,
					Message: fmt.Sprintf("Unknown test run ID %d", id),
					Code:    http.StatusBadRequest,
				}
			}
			runPtr.ID = int64(id)

			go func() {
				err := idx.IngestRun(*runPtr)
				if err != nil {
					log.Warningf("Failed to ingest runs: %s", err.Error())
				}
			}()

			missing = append(missing, *runPtr)
		} else {
			// Ensure that both `ids` and `runs` correspond to the same test runs.
			ids = append(ids, rq.RunIDs[i])
			runs = append(runs, run)
		}
	}

	// Return to client `http.StatusUnprocessableEntity` immediately if any runs
	// are missing.
	if len(runs) == 0 && len(missing) > 0 {
		data, err := json.Marshal(shared.SearchResponse{ // nolint:exhaustruct // TODO: Fix exhaustruct lint error.
			IgnoredRuns: missing,
		})
		if err != nil {
			return &searchError{
				Detail:  err,
				Message: "Failed to marshal results to JSON",
				Code:    http.StatusInternalServerError,
			}
		}
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, err = w.Write(data)
		if err != nil {
			log.Warningf("Failed to write data in api/search/cache handler: %s", err.Error())
		}

		return nil
	}

	// Prepare user query based on `ids` that are (or at least were a moment ago)
	// resident in `idx`. In the unlikely event that a run in `ids`/`runs` is no
	// longer in `idx`, `idx.Bind()` below will return an error.
	q := cq.PrepareUserQuery(ids, rq.BindToRuns(runs...))

	// Configure format, from request params.
	urlQuery := r.URL.Query()
	subtests, _ := shared.ParseBooleanParam(urlQuery, "subtests")
	interop, _ := shared.ParseBooleanParam(urlQuery, "interop")
	diff, _ := shared.ParseBooleanParam(urlQuery, "diff")
	diffFilter, _, err := shared.ParseDiffFilterParams(urlQuery)
	if err != nil {
		return &searchError{
			Detail:  err,
			Message: "Failed to parse diff filter",
			Code:    http.StatusBadRequest,
		}
	}
	opts := query.AggregationOpts{
		IncludeSubtests:         subtests != nil && *subtests,
		InteropFormat:           interop != nil && *interop,
		IncludeDiff:             diff != nil && *diff,
		DiffFilter:              diffFilter,
		IgnoreTestHarnessResult: shared.IsFeatureEnabled(store, "ignoreHarnessInTotal"),
	}
	plan, err := idx.Bind(runs, q)
	if err != nil {
		return &searchError{
			Detail:  err,
			Message: "Failed to create query plan",
			Code:    http.StatusInternalServerError,
		}
	}

	results := plan.Execute(runs, opts)
	res, ok := results.([]shared.SearchResult)
	if !ok {
		// nolint:exhaustruct // TODO: Fix exhaustruct lint error.
		return &searchError{
			Message: "Search index returned bad results",
			Code:    http.StatusInternalServerError,
		}
	}

	// Cull unchanged diffs, if applicable.
	if opts.IncludeDiff && !opts.DiffFilter.Unchanged {
		for i := range res {
			if res[i].Diff.IsEmpty() {
				res[i].Diff = nil
			}
		}
	}

	// Response always contains Runs and Results. If some runs are missing, then:
	// - Add missing runs to IgnoredRuns;
	// - (If no other error occurs) return `http.StatusUnprocessableEntity` to
	//   client.
	// nolint:exhaustruct // Not required since missing fields have omitempty.
	resp := shared.SearchResponse{
		Runs:    runs,
		Results: res,
	}
	if len(missing) != 0 {
		resp.IgnoredRuns = missing
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		return &searchError{
			Detail:  err,
			Message: "Failed to marshal results to JSON",
			Code:    http.StatusInternalServerError,
		}
	}
	if len(missing) != 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}

	_, err = w.Write(respData)
	if err != nil {
		log.Warningf("Failed to write data in api/search/cache handler: %s", err.Error())
	}

	return nil
}
