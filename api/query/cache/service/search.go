// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	cq "github.com/web-platform-tests/wpt.fyi/api/query/cache/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

type searchErr struct {
	// Detail is the internal error that should not be exposed to the end-user.
	Detail error
	// Message is the user-facing error message.
	Message string
	// Code is the HTTP status code for this error.
	Code int
}

func (e searchErr) Error() string {
	if e.Detail == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Detail)
}

func searchHandlerImpl(w http.ResponseWriter, r *http.Request) *searchErr {
	ctx := r.Context()
	log := shared.GetLogger(ctx)
	if r.Method != "POST" {
		return &searchErr{
			Message: "Invalid HTTP method " + r.Method,
			Code:    http.StatusBadRequest,
		}
	}

	reqData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &searchErr{
			Detail:  err,
			Message: "Failed to read request body",
			Code:    http.StatusInternalServerError,
		}
	}
	log.Debugf(string(reqData))
	if err := r.Body.Close(); err != nil {
		return &searchErr{
			Detail:  err,
			Message: "Failed to close request body",
			Code:    http.StatusInternalServerError,
		}
	}

	var rq query.RunQuery
	if err := json.Unmarshal(reqData, &rq); err != nil {
		return &searchErr{
			Detail:  err,
			Message: "Failed to unmarshal request body",
			Code:    http.StatusBadRequest,
		}
	}

	if len(rq.RunIDs) > *maxRunsPerRequest {
		return &searchErr{
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
		return &searchErr{
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
				return &searchErr{
					Detail:  err,
					Message: "Unknown test run ID " + string(id),
					Code:    http.StatusBadRequest,
				}
			}
			runPtr.ID = int64(id)
			go idx.IngestRun(*runPtr)
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
		data, err := json.Marshal(shared.SearchResponse{
			IgnoredRuns: missing,
		})
		if err != nil {
			return &searchErr{
				Detail:  err,
				Message: "Failed to marshal results to JSON",
				Code:    http.StatusInternalServerError,
			}
		}
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(data)
		return nil
	}

	// Prepare user query based on `ids` that are (or at least were a moment ago)
	// resident in `idx`. In the unlikely event that a run in `ids`/`runs` is no
	// longer in `idx`, `idx.Bind()` below will return an error.
	q := cq.PrepareUserQuery(ids, rq.AbstractQuery.BindToRuns(runs...))

	// Configure format, from request params.
	urlQuery := r.URL.Query()
	subtests, _ := shared.ParseBooleanParam(urlQuery, "subtests")
	interop, _ := shared.ParseBooleanParam(urlQuery, "interop")
	diff, _ := shared.ParseBooleanParam(urlQuery, "diff")
	diffFilter, _, err := shared.ParseDiffFilterParams(urlQuery)
	if err != nil {
		return &searchErr{
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
		return &searchErr{
			Detail:  err,
			Message: "Failed to create query plan",
			Code:    http.StatusInternalServerError,
		}
	}

	results := plan.Execute(runs, opts)
	res, ok := results.([]shared.SearchResult)
	if !ok {
		return &searchErr{
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
	resp := shared.SearchResponse{
		Runs:    runs,
		Results: res,
	}
	if len(missing) != 0 {
		resp.IgnoredRuns = missing
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		return &searchErr{
			Detail:  err,
			Message: "Failed to marshal results to JSON",
			Code:    http.StatusInternalServerError,
		}
	}
	if len(missing) != 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}

	w.Write(respData)
	return nil
}
