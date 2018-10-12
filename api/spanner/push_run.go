// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package spanner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/spanner"
	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/api/iterator"
)

const (
	countStmt = "SELECT COUNT(*) FROM results-apep WHERE run_id = @run_id"
)

// PushID is a unique identifier for a request to push a test run to
// Cloud Spanner.
type PushID struct {
	Time  time.Time `json:"time"`
	RunID int64     `json:"run_id"`
}

// HandlePushRun handles a request to push a test run to Cloud Spanner.
func HandlePushRun(ctx context.Context, auth Authenticator, w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Only PUT is supported", http.StatusMethodNotAllowed)
		return
	}

	if !auth.Authenticate(ctx, r) {
		http.Error(w, "Authentication error", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(r.URL.Query().Get("run_id"), 10, 0)
	if err != nil {
		http.Error(w, `Missing or invalid query param: "run_id"`, http.StatusBadRequest)
		return
	}

	t := time.Now().UTC()

	// TODO(mdittmer): Load run report URL from datastore, load report from GCS,
	// write results to Cloud Spanner.

	data, err := json.Marshal(PushID{t, id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(data)
}

func loadRun(ctx context.Context, client *datastore.Client, id int64) (*shared.TestRun, error) {
	logger := shared.GetLogger(ctx)

	logger.Infof("Loading TestRun entity with integra key %d", id)

	var run shared.TestRun
	err := client.Get(ctx, &datastore.Key{
		Kind: "TestRun",
		ID:   id,
	}, &run)
	if err != nil {
		logger.Errorf("Failed to load TestRun entity with integral key %d", id)
		return nil, err
	}
	return &run, nil
}

func loadRunReport(ctx context.Context, run *shared.TestRun) (*metrics.TestResultsReport, error) {
	logger := shared.GetLogger(ctx)

	if run.RawResultsURL == "" {
		str := fmt.Sprintf("TestRun entity ID=%d has no RawResultsURL", run.ID)
		logger.Errorf(str)
		return nil, errors.New(str)
	}

	logger.Infof("Reading report from %s", run.RawResultsURL)

	resp, err := http.Get(run.RawResultsURL)
	if err != nil {
		logger.Warningf("Failed to load raw results from \"%s\" for run ID=%d", run.RawResultsURL, run.ID)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		str := fmt.Sprintf("Non-OK HTTP status code of %d from \"%s\" for run ID=%d", resp.StatusCode, run.RawResultsURL, run.ID)
		logger.Warningf(str)
		return nil, errors.New(str)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warningf("Failed to read contents of \"%s\" for run ID=%d", run.RawResultsURL, run.ID)
		return nil, err
	}
	var report metrics.TestResultsReport
	err = json.Unmarshal(data, &report)
	if err != nil {
		logger.Warningf("Failed to unmarshal JSON from \"%s\" for run ID=%d", run.RawResultsURL, run.ID)
		return nil, err
	}
	if len(report.Results) == 0 {
		str := fmt.Sprintf("Empty report from run ID=%d (%s)", run.ID, run.RawResultsURL)
		logger.Warningf(str)
		return nil, errors.New(str)
	}

	logger.Infof("Read report for run ID=%d", run.ID)

	return &report, nil
}

func countReportResults(ctx context.Context, report *metrics.TestResultsReport) int64 {
	count := int64(0)
	for _, r := range report.Results {
		if len(r.Subtests) == 0 {
			count++
		} else {
			set := mapset.NewSet()
			for _, s := range r.Subtests {
				if set.Contains(s.Name) {
					shared.GetLogger(ctx).Warningf("Found test \"%s\" contains duplicate subtest name \"%s\"", r.Test, s.Name)
				} else {
					set.Add(s.Name)
				}
			}
			count += int64(set.Cardinality())
		}
	}
	return count
}

func countSpannerResults(ctx context.Context, client *spanner.Client, runID int64) (int64, error) {
	params := map[string]interface{}{
		"run_id": runID,
	}
	s := spanner.Statement{
		SQL:    countStmt,
		Params: params,
	}

	shared.GetLogger(ctx).Infof("Spanner query: \"%s\" with %v", countStmt, params)

	itr := client.Single().WithTimestampBound(spanner.MaxStaleness(1*time.Minute)).Query(ctx, s)
	defer itr.Stop()
	var count int64
	for {
		row, err := itr.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, err
		}

		err = row.Column(0, &count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

func numRowsToUpload(ctx context.Context, client *spanner.Client, runID int64, report *metrics.TestResultsReport) (int64, error) {
	totalRows := countReportResults(ctx, report)
	existingRows, err := countSpannerResults(ctx, client, runID)
	if err != nil {
		return 0, err
	}

	shared.GetLogger(ctx).Infof("Run %d contains %d rows (according to GCS); %d found in Spanner", runID, totalRows, existingRows)

	return totalRows - existingRows, nil
}
