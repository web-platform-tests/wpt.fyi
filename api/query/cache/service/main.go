// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"

	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/backfill"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/poll"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/api/option"

	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/monitor"
	cq "github.com/web-platform-tests/wpt.fyi/api/query/cache/query"

	"cloud.google.com/go/datastore"
	log "github.com/sirupsen/logrus"
)

var (
	port                   = flag.Int("port", 8080, "Port to listen on")
	numShards              = flag.Int("num_shards", runtime.NumCPU(), "Number of shards for parallelizing query execution")
	monitorInterval        = flag.Duration("monitor_interval", time.Second*5, "Polling interval for memory usage monitor")
	monitorMaxIngestedRuns = flag.Uint("monitor_max_ingested_runs", uint(10), "Maximum number of runs that can be ingested before memory monitor must run")
	maxHeapBytes           = flag.Uint64("max_heap_bytes", uint64(2e+11), "Soft limit on heap-allocated bytes before evicting test runs from memory")
	evictRunsPercent       = flag.Float64("evict_runs_percent", 0.1, "Decimal percentage indicating what fraction of runs to evict when soft memory limit is reached")
	updateInterval         = flag.Duration("update_interval", time.Second*10, "Update interval for polling for new runs")
	updateMaxRuns          = flag.Int("update_max_runs", 10, "The maximum number of latest runs to lookup in attempts to update indexes via polling")
	maxRunsPerRequest      = flag.Int("max_runs_per_request", 16, "Maximum number of runs that may be queried per request")

	// User-facing message for when runs in a request exceeds maxRunsPerRequest.
	// Set in init() after parsing flags.
	maxRunsPerRequestMsg string

	idx index.Index
	mon monitor.Monitor
)

func livenessCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Alive"))
}

func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	if idx == nil || mon == nil {
		http.Error(w, "Cache not yet ready", http.StatusServiceUnavailable)
		return
	}

	w.Write([]byte("Ready"))
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid HTTP method", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
	}
	err = r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to finish reading request body", http.StatusInternalServerError)
	}

	var rq query.RunQuery
	err = json.Unmarshal(data, &rq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(rq.RunIDs) > *maxRunsPerRequest {
		http.Error(w, maxRunsPerRequestMsg, http.StatusBadRequest)
		return
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
	store, err := getDatastore()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open datastore: %s", err.Error()), http.StatusInternalServerError)
		return
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
			err := store.Get(store.NewIDKey("TestRun", int64(id)), runPtr)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unknown test run ID: %d", id), http.StatusBadRequest)
				return
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
		data, err = json.Marshal(query.SearchResponse{
			IgnoredRuns: missing,
		})
		if err != nil {
			http.Error(w, "Failed to marshal results to JSON", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(data)
		return
	}

	// Prepare user query based on `ids` that are (or at least were a moment ago)
	// resident in `idx`. In the unlikely event that a run in `ids`/`runs` is no
	// longer in `idx`, `idx.Bind()` below will return an error.
	q := cq.PrepareUserQuery(ids, rq.AbstractQuery.BindToRuns(runs...))

	// Configure format, from request params.
	urlQuery := r.URL.Query()
	_, subtests := urlQuery["subtests"]
	_, interop := urlQuery["interop"]
	_, diff := urlQuery["diff"]
	diffFilter, _, err := shared.ParseDiffFilterParams(urlQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	opts := query.AggregationOpts{
		IncludeSubtests:         subtests,
		InteropFormat:           interop,
		IncludeDiff:             diff,
		DiffFilter:              diffFilter,
		IgnoreTestHarnessResult: shared.IsFeatureEnabled(store, "ignoreHarnessInTotal"),
	}
	plan, err := idx.Bind(runs, q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results := plan.Execute(runs, opts)
	res, ok := results.([]query.SearchResult)
	if !ok {
		http.Error(w, "Search index returned bad results", http.StatusInternalServerError)
		return
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
	resp := query.SearchResponse{
		Runs:    runs,
		Results: res,
	}
	if len(missing) != 0 {
		resp.IgnoredRuns = missing
	}
	data, err = json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal results to JSON", http.StatusInternalServerError)
		return
	}
	if len(missing) != 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}

	w.Write(data)
}

func getDatastore() (shared.Datastore, error) {
	ctx := context.Background()
	var client *datastore.Client
	var err error
	if shared.GCPCredentialsFile != nil && *shared.GCPCredentialsFile != "" {
		client, err = datastore.NewClient(ctx, *shared.ProjectID, option.WithCredentialsFile(*shared.GCPCredentialsFile))
	} else {
		client, err = datastore.NewClient(ctx, *shared.ProjectID)
	}
	if err != nil {
		return nil, err
	}
	d := shared.NewCloudDatastore(ctx, client)
	return d, nil
}

func init() {
	flag.Parse()

	maxRunsPerRequestMsg = fmt.Sprintf("Too many runs specified; maximum is %d.", *maxRunsPerRequest)
}

func main() {
	flag.Parse()
	shared.InitProjectID()

	log.Infof("Serving index with %d shards", *numShards)
	// TODO: Use different field configurations for index, backfiller, monitor?
	logger := log.StandardLogger()

	idx, err := index.NewShardedWPTIndex(index.HTTPReportLoader{}, *numShards)
	if err != nil {
		log.Fatalf("Failed to instantiate index: %v", err)
	}

	fetcher := backfill.NewDatastoreRunFetcher(*shared.ProjectID, shared.GCPCredentialsFile, logger)
	mon, err = backfill.FillIndex(fetcher, logger, monitor.GoRuntime{}, *monitorInterval, *monitorMaxIngestedRuns, *maxHeapBytes, *evictRunsPercent, idx)
	if err != nil {
		log.Fatalf("Failed to initiate index backkfill: %v", err)
	}

	// Index, backfiller, monitor now in place. Start polling to load runs added
	// after backfilling was started.
	go poll.KeepRunsUpdated(fetcher, logger, *updateInterval, *updateMaxRuns, idx)

	http.HandleFunc("/_ah/liveness_check", livenessCheckHandler)
	http.HandleFunc("/_ah/readiness_check", readinessCheckHandler)
	http.HandleFunc("/api/search/cache", searchHandler)
	log.Infof("Listening on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
