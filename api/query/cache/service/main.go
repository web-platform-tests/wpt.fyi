// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/sirupsen/logrus"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/backfill"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/monitor"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/poll"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/api/option"
)

// nolint:gochecknoglobals // TODO: Fix gochecknoglobals lint error
var (
	port      = flag.Int("port", 8080, "Port to listen on")
	projectID = flag.String("project_id", "",
		"Google Cloud Platform project ID, used for connecting to Datastore")
	gcpCredentialsFile = flag.String("gcp_credentials_file", "",
		"Path to Google Cloud Platform credentials file, if necessary")
	numShards              = flag.Int("num_shards", runtime.NumCPU(), "Number of shards for parallelizing query execution")
	monitorInterval        = flag.Duration("monitor_interval", time.Second*5, "Polling interval for memory usage monitor")
	monitorMaxIngestedRuns = flag.Uint("monitor_max_ingested_runs", 10,
		"Maximum number of runs that can be ingested before memory monitor must run")
	maxHeapBytes = flag.Uint64("max_heap_bytes", 0,
		"Soft limit on heap-allocated bytes before evicting test runs from memory")
	evictRunsPercent = flag.Float64("evict_runs_percent", 0.1,
		"Decimal percentage indicating what fraction of runs to evict when soft memory limit is reached")
	updateInterval = flag.Duration("update_interval", time.Second*10,
		"Update interval for polling for new runs")
	updateMaxRuns = flag.Int("update_max_runs", 10,
		"The maximum number of latest runs to lookup in attempts to update indexes via polling")
	maxRunsPerRequest = flag.Int("max_runs_per_request", 16,
		"Maximum number of runs that may be queried per request")

	// User-facing message for when runs in a request exceeds maxRunsPerRequest.
	// Set in init() after parsing flags.
	maxRunsPerRequestMsg string

	idx index.Index
	mon monitor.Monitor
)

func livenessCheckHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Alive"))
	if err != nil {
		logger := shared.GetLogger(r.Context())
		logger.Warningf("Failed to write data in liveness check handler: %s", err.Error())
	}
}

func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	if idx == nil || mon == nil {
		http.Error(w, "Cache not yet ready", http.StatusServiceUnavailable)

		return
	}

	_, err := w.Write([]byte("Ready"))
	if err != nil {
		logger := shared.GetLogger(r.Context())
		logger.Warningf("Failed to write data in readiness check handler: %s", err.Error())
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	err := searchHandlerImpl(w, r)
	if err != nil {
		log := shared.GetLogger(r.Context())
		log.Errorf(err.Error())
		http.Error(w, err.Message, err.Code)
	}
}

// nolint:ireturn // TODO: Fix ireturn lint error
func getDatastore(ctx context.Context) (shared.Datastore, error) {
	var client *datastore.Client
	var err error
	if gcpCredentialsFile != nil && *gcpCredentialsFile != "" {
		client, err = datastore.NewClient(
			ctx, *projectID, option.WithAuthCredentialsFile(option.ServiceAccount, *gcpCredentialsFile))
	} else {
		client, err = datastore.NewClient(ctx, *projectID)
	}
	if err != nil {
		return nil, err
	}
	d := shared.NewCloudDatastore(ctx, client)

	return d, nil
}

func init() {
	flag.Parse()

	if *maxHeapBytes == 0 {
		var sysinfo syscall.Sysinfo_t
		if err := syscall.Sysinfo(&sysinfo); err != nil {
			logrus.Fatalf("Unable to get total system memory: %s", err.Error())
		}
		sysmem := float64(sysinfo.Totalram) * float64(sysinfo.Unit)
		// Reserve 2GB or 50% of the total memory for system (whichever is smaller).
		if sysmem-2e9 > sysmem*0.5 {
			*maxHeapBytes = uint64(sysmem - 2e9)
		} else {
			*maxHeapBytes = uint64(sysmem * 0.5)
		}
		logrus.Infof("Detected total system memory: %d; setting max heap size to %d", uint64(sysmem), *maxHeapBytes)
	}

	maxRunsPerRequestMsg = fmt.Sprintf("Too many runs specified; maximum is %d.", *maxRunsPerRequest)

	autoProjectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if autoProjectID == "" {
		logrus.Warningf("Failed to get project ID from environment")
	} else {
		if *projectID == "" {
			logrus.Infof("Using project ID from environment: %s", autoProjectID)
			*projectID = autoProjectID
		} else if *projectID != autoProjectID {
			logrus.Warningf(
				"Using project ID from flag: %s, even though environment reports project ID: %s",
				*projectID,
				autoProjectID,
			)
		} else {
			logrus.Infof("Using project ID: %s", *projectID)
		}
	}
}

func main() {
	logrus.Infof("Serving index with %d shards", *numShards)
	// nolint:godox // TODO: Use different field configurations for index, backfiller, monitor?
	logger := logrus.StandardLogger()

	var err error
	idx, err = index.NewShardedWPTIndex(index.HTTPReportLoader{}, *numShards)
	if err != nil {
		logrus.Fatalf("Failed to instantiate index: %v", err)
	}

	store, err := backfill.GetDatastore(*projectID, gcpCredentialsFile, logger)
	if err != nil {
		logrus.Fatalf("Failed to get datastore: %s", err)
	}
	mon, err = backfill.FillIndex(
		store,
		logger,
		monitor.GoRuntime{},
		*monitorInterval,
		*monitorMaxIngestedRuns,
		*maxHeapBytes,
		*evictRunsPercent,
		idx,
	)
	if err != nil {
		logrus.Fatalf("Failed to initiate index backkfill: %v", err)
	}

	// Index, backfiller, monitor now in place. Start polling to load runs added
	// after backfilling was started.
	go poll.KeepRunsUpdated(store, logger, *updateInterval, *updateMaxRuns, idx)

	// Initializes clients.
	if err = shared.Clients.Init(context.Background()); err != nil {
		logrus.Fatalf("Failed to initialize Google Cloud clients: %v", err)
	}
	defer shared.Clients.Close()

	// Polls Metadata update every 10 minutes.
	go poll.StartMetadataPollingService(context.Background(), logger, time.Minute*10)

	// Polls Web Feature Manifest update every 30 minutes.
	go poll.StartWebFeaturesManifestPollingService(context.Background(), logger, time.Minute*30)

	http.HandleFunc("/_ah/liveness_check", livenessCheckHandler)
	http.HandleFunc("/_ah/readiness_check", readinessCheckHandler)
	http.HandleFunc("/api/search/cache", shared.HandleWithLogging(searchHandler))
	logrus.Infof("Listening on port %d", *port)
	// nolint:gosec // TODO: Fix gosec lint error (G114).
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
