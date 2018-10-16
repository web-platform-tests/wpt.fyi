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
	"math"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/spanner"
	retry "github.com/avast/retry-go"
	mapset "github.com/deckarep/golang-set"
	farm "github.com/dgryski/go-farm"
	log "github.com/sirupsen/logrus"
	"github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/sync/semaphore"
	"google.golang.org/api/iterator"
)

const (
	numConcurrentBatches = 1000
	batchSize            = 1000
	testsTableName       = "tests"
	resultsTableName     = "results"
	countStmt            = "SELECT COUNT(*) FROM results WHERE run_id = @run_id"
)

// PushID is a unique identifier for a request to push a test run to
// Cloud Spanner.
type PushID struct {
	Time  time.Time `json:"time"`
	RunID int64     `json:"run_id"`
}

// HandlePushRun handles a request to push a test run to Cloud Spanner.
func HandlePushRun(ctx context.Context, api API, w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Only PUT is supported", http.StatusMethodNotAllowed)
		return
	}

	if !api.Authenticate(ctx, r) {
		http.Error(w, "Authentication error", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(r.URL.Query().Get("run_id"), 10, 0)
	if err != nil {
		http.Error(w, `Missing or invalid query param: "run_id"`, http.StatusBadRequest)
		return
	}

	t := time.Now().UTC()
	pushID := PushID{t, id}
	pushCtx := context.WithValue(context.Background(), shared.DefaultLoggerCtxKey(), log.WithFields(log.Fields{
		"spanner_push_run_id": pushID,
	}))

	go pushRun(pushCtx, api, id)

	data, err := json.Marshal(PushID{t, id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(data)
}

func pushRun(ctx context.Context, api API, id int64) {
	var (
		dsClient *datastore.Client
		sClient  *spanner.Client
		err      error
		logger   = shared.GetLogger(ctx)
	)
	dsClient, err = api.DatastoreConnect(ctx)
	if err != nil {
		logger.Errorf("Spanner push run failed connecting to datastore: %v", err)
		return
	}

	run, err := loadRun(ctx, dsClient, id)
	if err != nil {
		logger.Errorf("Spanner push run failed loading run: %v", err)
		return
	}

	report, err := loadRunReport(ctx, run)
	if err != nil {
		logger.Errorf("Spanner push run failed loading run report: %v", err)
		return
	}

	sClient, err = api.SpannerConnect(ctx)
	if err != nil {
		logger.Errorf("Spanner push run failed connecting to spanner: %v", err)
		return
	}

	// Determine number of rows missing in spanner.
	// Retries: 5.
	var n int64
	retry.Do(func() error {
		n, err = numRowsToUpload(ctx, sClient, run.ID, report)
		return err
	}, retry.Attempts(5), retry.OnRetry(func(n uint, err error) {
		logger.Warningf("Attempt #%d to get number of rows to push failed: %v", n, err)
	}))
	if err != nil {
		logger.Errorf("Spanner push run failed calculating number of rows to upload: %v", err)
		return
	}
	if n < 0 {
		logger.Warningf("Spanner contains %d more rows than WPT report", -n)
		return
	}
	if n == 0 {
		logger.Infof("Run already uploaded to spanner")
		return
	}

	// Generate mutations for complete run.
	logger.Infof("Queuing rows for run")
	tests, results := getMutationsForRun(run.ID, report)
	logger.Infof("Queued %d+%d rows", len(tests), len(results))

	// Write rows to spanner in batches.
	// Retries: 5 per batch.
	writeMutations := func(muts []*spanner.Mutation) []error {
		logger.Infof("Writing %d-row batches", len(muts))
		s := semaphore.NewWeighted(numConcurrentBatches)
		ec := make(chan error, int64(math.Ceil(float64(len(muts))/float64(batchSize))))
		writeBatch := func(m, n int) error {
			batch := muts[m:n]

			logger.Infof("Writing batch: [%d,%d)", m, n)
			_, err := sClient.Apply(ctx, batch)
			if err != nil {
				return err
			}
			logger.Infof("Wrote batch: [%d,%d)", m, n)
			return nil
		}
		retryableWriteBatch := func(m, n int) {
			defer s.Release(1)
			err := retry.Do(func() error {
				return writeBatch(m, n)
			}, retry.Attempts(5), retry.OnRetry(func(num uint, err error) {
				logger.Warningf("Attempt #%d to write batch [%d,%d) failed: %v", num, m, n, err)
			}))
			if err != nil {
				ec <- err
			}
		}
		var end int
		for end = batchSize; end <= len(muts); end += batchSize {
			s.Acquire(ctx, 1)
			go retryableWriteBatch(end-batchSize, end)
		}
		// Corner case: Leftover rows when len(muts) % batchSize != 0.
		if end != len(muts) {
			s.Acquire(ctx, 1)
			logger.Infof("Writing small batch: [%d,%d)", end-batchSize, len(muts))
			go retryableWriteBatch(end-batchSize, len(muts))
			logger.Infof("Wrote small batch: [%d,%d)", end-batchSize, len(muts))
		}
		s.Acquire(ctx, numConcurrentBatches)

		errs := make([]error, 0)
		close(ec)
		for err := range ec {
			logger.Errorf("Failed to write batch: %v", err)
			errs = append(errs, err)
		}
		return errs
	}

	// Write tests (parent table), then results (child table).
	errs := writeMutations(tests)
	if len(errs) != 0 {
		logger.Errorf("Spanner push run failed with %d test batch write errors", len(errs))
		return
	}
	errs = writeMutations(results)
	if len(errs) != 0 {
		logger.Errorf("Spanner push run failed with %d result batch write errors", len(errs))
		return
	}

	logger.Infof("Spanner push run succeeded: Wrote batches for %d+%d rows", len(tests), len(results))
}

// loadRun loads shared.TestRun data from datastore, given an integral ID
// (Datastore key).
func loadRun(ctx context.Context, client *datastore.Client, id int64) (*shared.TestRun, error) {
	logger := shared.GetLogger(ctx)

	logger.Infof("Loading TestRun entity with integral key %d", id)

	var run shared.TestRun
	err := client.Get(ctx, &datastore.Key{
		Kind: "TestRun",
		ID:   id,
	}, &run)
	if err != nil {
		logger.Errorf("Failed to load TestRun entity with integral key %d", id)
		return nil, err
	}
	run.ID = id
	return &run, nil
}

// loadRunReport loads a metrics.TestResultsReport using the URL specified in
// run.
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

// countReportResults counts the number of meaningfully distinct test results
// detailed in report.
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

// countSpannerResults counts the number of test results bound to the given
// runID in Cloud Spanner.
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

// numRowsToUpload delegates to internal counting functions to compare the
// number of test results in a run report to the number of results stored in
// Cloud Spanner. The return value is the number of results in the report that
// do not appear in Cloud Spanner.
func numRowsToUpload(ctx context.Context, client *spanner.Client, runID int64, report *metrics.TestResultsReport) (int64, error) {
	totalRows := countReportResults(ctx, report)
	existingRows, err := countSpannerResults(ctx, client, runID)
	if err != nil {
		return 0, err
	}

	shared.GetLogger(ctx).Infof("Run %d contains %d results (according to GCS); %d found in Spanner", runID, totalRows, existingRows)

	return totalRows - existingRows, nil
}

// getMutationsForRun returns (test table mutations, results table mutations)
// associated with tests and test results contained in report.
func getMutationsForRun(runID int64, report *metrics.TestResultsReport) ([]*spanner.Mutation, []*spanner.Mutation) {
	testsRows := make([]*spanner.Mutation, 0)
	resultsRows := make([]*spanner.Mutation, 0)
	for _, r := range report.Results {
		if len(r.Subtests) == 0 {
			testsRows = appendTestsRow(r.Test, nil, testsRows)
			resultsRows = appendResultsRow(runID, r.Status, r.Test, nil, r.Message, resultsRows)
		} else {
			for _, s := range r.Subtests {
				testsRows = appendTestsRow(r.Test, &s.Name, testsRows)
				resultsRows = appendResultsRow(runID, s.Status, r.Test, &s.Name, s.Message, resultsRows)
			}
		}
	}
	return testsRows, resultsRows
}

// appendTestsRow appends a *spanner.Mutation to rows, containing the mutation
// required to ensure that the associated (sub)test is stored in Cloud Spanner.
func appendTestsRow(test string, subtest *string, rows []*spanner.Mutation) []*spanner.Mutation {
	testID := computeTestID(test, subtest)
	testsRowMap := map[string]interface{}{
		"test_id": testID,
		"test":    test,
	}
	if subtest != nil && *subtest != "" {
		testsRowMap["subtest"] = *subtest
	} else {
		testsRowMap["subtest"] = spanner.NullString{}
	}

	// Use InsertOrUpdate to ensure provided fields are correct without triggering
	// a Delete that could break sub-table consistency.
	return append(rows, spanner.InsertOrUpdateMap(testsTableName, testsRowMap))
}

// appendResultsRow appends a *spanner.Mutation to rows, containing the mutation
// required to ensure that the associated test result is stored in Cloud
// Spanner.
func appendResultsRow(runID int64, status, test string, subtest, message *string, rows []*spanner.Mutation) []*spanner.Mutation {
	testID := computeTestID(test, subtest)
	resultsRowMap := map[string]interface{}{
		"test_id": testID,
		"run_id":  runID,
		"result":  shared.TestStatusValueFromString(status),
	}
	if message != nil && *message != "" {
		resultsRowMap["message"] = *message
	}

	// Use Replace to clobber any existing row.
	return append(rows, spanner.ReplaceMap(resultsTableName, resultsRowMap))
}

// computeTestID computes a stable int64 ID for a test+(optional)subtest pair.
func computeTestID(test string, subtest *string) int64 {
	if subtest != nil && *subtest != "" {
		return int64(farm.Fingerprint64([]byte(test + "\x00" + *subtest)))
	}
	return int64(farm.Fingerprint64([]byte(test)))
}
