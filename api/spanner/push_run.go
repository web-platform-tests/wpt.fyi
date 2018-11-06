// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package spanner

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc/codes"

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
	numConcurrentBatches    = 1000
	batchSize               = 1000
	testsTableName          = "Tests"
	resultsTableName        = "Results"
	runsTableName           = "Runs"
	runResultsTableName     = "RunResults"
	runResultTestsTableName = "RunResultTests"
	testNameColumnName      = "TestName"
	subtestNameColumnName   = "SubtestName"
	countStmt               = "SELECT COUNT(*) FROM RunResultTests WHERE RunID = @RunID"
)

// PushID is a unique identifier for a request to push a test run to
// Cloud Spanner.
type PushID struct {
	Time  time.Time `json:"time"`
	RunID int64     `json:"run_id"`
}

// RunKey is a wrapper type for run identifiers that match int64 IDs from
// Datastore.
type RunKey struct {
	RunID int64
}

// Run is a record in the table specified by runsTableName.
type Run struct {
	RunKey
	BrowserName     string
	BrowserVersion  string
	OSName          string
	OSVersion       string
	WPTRevisionHash []byte
	ResultsURL      spanner.NullString
	CreatedAt       time.Time
	TimeStart       time.Time
	TimeEnd         time.Time
	RawResultsURL   spanner.NullString
	Labels          []string
}

// NewRun constructs a Run record from a shared.TestRun object.
func NewRun(r *shared.TestRun) (*Run, error) {
	hash, err := hex.DecodeString(r.FullRevisionHash)
	if err != nil {
		return nil, err
	}
	return &Run{
		RunKey:          RunKey{r.ID},
		BrowserName:     r.BrowserName,
		BrowserVersion:  r.BrowserVersion,
		OSName:          r.OSName,
		OSVersion:       r.OSVersion,
		WPTRevisionHash: hash,
		ResultsURL:      toNullString(&r.ResultsURL),
		CreatedAt:       r.CreatedAt,
		TimeStart:       r.TimeStart,
		TimeEnd:         r.TimeEnd,
		RawResultsURL:   toNullString(&r.RawResultsURL),
		Labels:          r.Labels,
	}, nil
}

// ResultKey is a wrapper for result identifiers detailed in "TestStatus..."
// symbols in the shared package.
type ResultKey struct {
	ResultID int64
}

// Result is a record in the table specified by resultsTableName.
type Result struct {
	ResultKey
	Name        string
	Description spanner.NullString
}

// NewResult constructs a Result record from a name and description.
func NewResult(name string, desc *string) *Result {
	id := shared.TestStatusValueFromString(name)
	return &Result{
		ResultKey:   ResultKey{id},
		Name:        name,
		Description: toNullString(desc),
	}
}

// TestKey is wrapper for test identifiers. These are derived from test and
// subtest names using the farm.Fingerprint64() function.
type TestKey struct {
	TestID    int64
	SubtestID spanner.NullInt64
}

// Test is a record in the table specified by testsTableName.
type Test struct {
	TestKey
	TestName    string
	SubtestName spanner.NullString
}

// NewTest constructs a Test record from a test and subtest name pair.
func NewTest(name string, sub *string) *Test {
	key := computeTestKey(name, sub)
	subName := toNullString(sub)
	return &Test{key, name, subName}
}

// RunResult is a record in the table specified by runResultsTableName.
type RunResult struct {
	RunKey
	ResultKey
}

// RunResultTest is a record in the table specified by runResultTestsTableName.
type RunResultTest struct {
	RunKey
	ResultKey
	TestKey
	Message spanner.NullString
}

// Structs is a collection of data that constitutes the records for a single
// test run, structured for easy lookup for the "write run to Cloud Spanner"
// use case.
type Structs struct {
	Runs           map[RunKey]*Run
	Results        map[ResultKey]*Result
	Tests          map[TestKey]*Test
	RunResults     map[RunKey]map[ResultKey]*RunResult
	RunResultTests map[RunKey]map[ResultKey]map[TestKey]*RunResultTest
}

// NewStructs constructs an empty, usable Structs.
func NewStructs() *Structs {
	return &Structs{
		make(map[RunKey]*Run),
		make(map[ResultKey]*Result),
		make(map[TestKey]*Test),
		make(map[RunKey]map[ResultKey]*RunResult),
		make(map[RunKey]map[ResultKey]map[TestKey]*RunResultTest),
	}
}

// AddRun adds a Run to a Structs.
func (s *Structs) AddRun(r *Run) {
	s.Runs[r.RunKey] = r
}

// AddResult adds a result to a Structs.
func (s *Structs) AddResult(r *Result) {
	s.Results[r.ResultKey] = r
}

// AddTest adds a test to a Structs.
func (s *Structs) AddTest(t *Test) error {
	if _, ok := s.Tests[t.TestKey]; ok {
		return fmt.Errorf(`Duplicate add of test with key %v`, t.TestKey)
	}
	s.Tests[t.TestKey] = t
	return nil
}

// AddRunResult adds a RunResult to a Structs.
func (s *Structs) AddRunResult(run *Run, res *Result) {
	if _, ok := s.RunResults[run.RunKey]; !ok {
		s.RunResults[run.RunKey] = make(map[ResultKey]*RunResult)
	}
	s.RunResults[run.RunKey][res.ResultKey] = &RunResult{
		RunKey:    run.RunKey,
		ResultKey: res.ResultKey,
	}
}

// AddRunResultTest adds a RunResultTest to a Structs.
func (s *Structs) AddRunResultTest(run *Run, res *Result, t *Test, message *string) {
	msg := toNullString(message)
	if _, ok := s.RunResultTests[run.RunKey]; !ok {
		s.RunResultTests[run.RunKey] = make(map[ResultKey]map[TestKey]*RunResultTest)
	}
	if _, ok := s.RunResultTests[run.RunKey][res.ResultKey]; !ok {
		s.RunResultTests[run.RunKey][res.ResultKey] = make(map[TestKey]*RunResultTest)
	}
	s.RunResultTests[run.RunKey][res.ResultKey][t.TestKey] = &RunResultTest{
		TestKey:   t.TestKey,
		RunKey:    run.RunKey,
		ResultKey: res.ResultKey,
		Message:   msg,
	}
}

// RemoveTestData removes all data associated with a test, cascading removals of
// nested/interleaved data when the removed test was the last entry associated
// with the nested/interleaved run and/or result.
func (s *Structs) RemoveTestData(t *Test) {
	delete(s.Tests, t.TestKey)
	for k1 := range s.RunResultTests {
		for k2 := range s.RunResultTests[k1] {
			if _, ok := s.RunResultTests[k1][k2][t.TestKey]; ok {
				delete(s.RunResultTests[k1][k2], t.TestKey)
				if len(s.RunResultTests[k1][k2]) == 0 {
					delete(s.RunResultTests[k1], k2)
					delete(s.RunResults[k1], k2)
				}
				if len(s.RunResultTests[k1]) == 0 {
					delete(s.RunResultTests, k1)
					delete(s.RunResults, k1)
					delete(s.Runs, k1)
				}
			}
		}
	}
}

// ToMutations unpacks a Structs into three collections of Cloud Spanner
// mutations. Due to table interleaving, the collections must be applied in
// order; i.e., all mutations in the first collection must be applied before
// any mutations in the second collection, and so on.
func (s *Structs) ToMutations() ([]*spanner.Mutation, []*spanner.Mutation, []*spanner.Mutation, error) {
	m1s := make([]*spanner.Mutation, 0, len(s.Runs)+len(s.Results))
	m2s := make([]*spanner.Mutation, 0, len(s.RunResults))
	m3s := make([]*spanner.Mutation, 0, len(s.RunResultTests))

	// Note: Skip s.Tests, which is handled separately t avoid has collisions.
	for _, r := range s.Runs {
		m, err := spanner.InsertOrUpdateStruct(runsTableName, r)
		if err != nil {
			return nil, nil, nil, err
		}
		m1s = append(m1s, m)
	}
	for _, r := range s.Results {
		m, err := spanner.InsertOrUpdateStruct(resultsTableName, r)
		if err != nil {
			return nil, nil, nil, err
		}
		m1s = append(m1s, m)
	}
	for _, m1 := range s.RunResults {
		for _, tr := range m1 {
			m, err := spanner.InsertOrUpdateStruct(runResultsTableName, tr)
			if err != nil {
				return nil, nil, nil, err
			}
			m2s = append(m2s, m)
		}
	}
	for _, m1 := range s.RunResultTests {
		for _, m2 := range m1 {
			for _, rrt := range m2 {
				m, err := spanner.InsertOrUpdateStruct(runResultTestsTableName, rrt)
				if err != nil {
					return nil, nil, nil, err
				}
				m3s = append(m3s, m)
			}
		}
	}

	return m1s, m2s, m3s, nil
}

// ConcurrentSlice is a slice that supports concurrent read/write.
type ConcurrentSlice struct {
	m *sync.RWMutex
	s []interface{}
}

// Append adds a single value to the underlying slice.
func (s *ConcurrentSlice) Append(v interface{}) {
	s.m.Lock()
	defer s.m.Unlock()
	s.s = append(s.s, v)
}

// Len returns the length of the underlying slice.
func (s *ConcurrentSlice) Len() int {
	s.m.RLock()
	defer s.m.RUnlock()
	return len(s.s)
}

// Slice returns a copy of the underlying slice.
func (s *ConcurrentSlice) Slice() []interface{} {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.s[0:]
}

// NewConcurrentSlice initializes a new ConcurrentSlice.
func NewConcurrentSlice() *ConcurrentSlice {
	return &ConcurrentSlice{
		m: &sync.RWMutex{},
		s: make([]interface{}, 0),
	}
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

	logger.Infof("Preparing data for run %d", run.ID)

	ss := NewStructs()

	r, err := NewRun(run)
	if err != nil {
		logger.Errorf("Spanner push failed to constructing run: %v", err)
		return
	}
	ss.AddRun(r)

	for _, result := range report.Results {
		t := NewTest(result.Test, nil)
		err := ss.AddTest(t)
		if err != nil {
			logger.Warningf("Error adding test from report: %v", err)
		} else {
			res := NewResult(result.Status, nil)
			ss.AddResult(res)
			ss.AddRunResult(r, res)
			ss.AddRunResultTest(r, res, t, result.Message)
		}

		names := mapset.NewSet()
		for _, s := range result.Subtests {
			if names.Contains(s.Name) {
				logger.Warningf("Found test \"%s\" contains duplicate subtest name \"%s\"", result.Test, s.Name)
			}
			names.Add(s.Name)

			t := NewTest(result.Test, &s.Name)
			err := ss.AddTest(t)
			if err != nil {
				logger.Warningf("Error adding test from report: %v", err)
			} else {
				res := NewResult(s.Status, nil)
				ss.AddResult(res)
				ss.AddRunResult(r, res)
				ss.AddRunResultTest(r, res, t, s.Message)
			}
		}
	}
	report = nil

	s := semaphore.NewWeighted(numConcurrentBatches)

	logger.Infof("Updating tests from %d-test run %d", len(ss.Tests), run.ID)

	fatalErrs := NewConcurrentSlice()
	hashCollisions := NewConcurrentSlice()

	writeTestBatch := func(batchSync *semaphore.Weighted, rowGroupSync *sync.WaitGroup, rows []*spanner.Mutation, tests []*Test, m, n int) {
		defer rowGroupSync.Done()
		defer batchSync.Release(1)
		batch := rows[m:n]
		ts := tests[m:n]

		err := retry.Do(func() error {
			logger.Infof("Writing test batch for %d-test run %d: [%d,%d)", len(rows), run.ID, m, n)

			for i := range batch {
				err := func(ins *spanner.Mutation, t *Test) error {
					insCtx, insCancel := context.WithTimeout(ctx, time.Second*60)
					defer insCancel()

					_, err = sClient.Apply(insCtx, []*spanner.Mutation{ins})
					if err == nil {
						logger.Infof("Wrote new test to Spanner: %v", *t)
						return nil
					}

					// Continue with hash collission check iff error was Cloud Spanner error
					// "already exists". Other errors are unexpected.
					spanErr, ok := err.(*spanner.Error)
					if !ok || spanErr.Code != codes.AlreadyExists {
						return err
					}

					readCtx, readCancel := context.WithTimeout(ctx, time.Second*60)
					defer readCancel()

					row, err := sClient.Single().ReadRow(readCtx, testsTableName, spanner.Key{t.TestID, t.SubtestID}, []string{testNameColumnName, subtestNameColumnName})
					if err != nil {
						return err
					}

					var testName string
					err = row.ColumnByName(testNameColumnName, &testName)
					if err != nil {
						return err
					}
					var subtestName spanner.NullString
					err = row.ColumnByName(subtestNameColumnName, &subtestName)
					if err != nil {
						return err
					}
					if t.TestName != testName || t.SubtestName != subtestName {
						hashCollisions.Append(t)
						logger.Errorf(`Hash collision: Test identifier <%d, %v> mapped to different test+subtest names: "%s".%v != "%s".%v`, t.TestID, t.SubtestID, t.TestName, t.SubtestName, testName, subtestName)
					}

					return nil
				}(batch[i], ts[i])

				if err != nil {
					return err
				}
			}

			logger.Infof("Wrote test batch for %d-test run %d: [%d,%d)", len(rows), run.ID, m, n)
			return nil
		}, retry.Attempts(5), retry.OnRetry(func(n uint, err error) {
			logger.Warningf("Retrying failed test batch for %d-test run %d: [%d,%d): %v", len(rows), run.ID, m, n, err)
		}))
		if err != nil {
			fatalErrs.Append(err)
		}
	}
	writeTestRows := func(rows []*spanner.Mutation, tests []*Test) *sync.WaitGroup {
		var wg sync.WaitGroup
		var end int
		for end = batchSize; end <= len(rows); end += batchSize {
			wg.Add(1)
			s.Acquire(ctx, 1)
			go writeTestBatch(s, &wg, rows[0:], tests[0:], end-batchSize, end)
		}
		if end != len(rows) {
			wg.Add(1)
			s.Acquire(ctx, 1)
			logger.Infof("Writing small test batch for %d-test run %d: [%d,%d)", len(rows), run.ID, end-batchSize, len(rows))
			go writeTestBatch(s, &wg, rows[0:], tests[0:], end-batchSize, len(rows))
			logger.Infof("Wrote small test batch for %d-test run %d: [%d,%d)", len(rows), run.ID, end-batchSize, len(rows))
		}
		return &wg
	}
	inss := make([]*spanner.Mutation, 0, len(ss.Tests))
	ts := make([]*Test, 0, len(ss.Tests))
	for _, t := range ss.Tests {
		ins, err := spanner.InsertStruct(testsTableName, t)
		if err != nil {
			logger.Errorf("Spanner push failed to generating tests mutations: %v", err)
			return
		}
		inss = append(inss, ins)
		ts = append(ts, t)
	}

	logger.Infof("Writing %d tests for run %d", len(ts), run.ID)
	writeTestRows(inss, ts).Wait()
	if fatalErrs.Len() > 0 {
		logger.Errorf("Spanner push failed to write tests for run %v: %v", run.ID, fatalErrs.Slice())
		return
	}
	logger.Infof("Wrote tests for %d-test run %d", len(ts), run.ID)
	inss = nil
	ts = nil

	for _, i := range hashCollisions.Slice() {
		t := i.(*Test)
		ss.RemoveTestData(t)
	}

	logger.Infof("Generating row-based mutations for run %d", run.ID)
	r1s, r2s, r3s, err := ss.ToMutations()
	if err != nil {
		logger.Errorf("Spanner push failed to generating mutations: %v", err)
		return
	}
	numRows := len(r1s) + len(r2s) + len(r3s)
	logger.Infof("Generated %d rows for run %d", numRows, run.ID)

	logger.Infof("Writing batches for %d-row run %d", numRows, run.ID)

	writeBatch := func(batchSync *semaphore.Weighted, rowGroupSync *sync.WaitGroup, rows []*spanner.Mutation, m, n int) {
		defer rowGroupSync.Done()
		defer batchSync.Release(1)
		batch := rows[m:n]

		err := retry.Do(func() error {
			logger.Infof("Writing batch for %d-row run %d: [%d,%d)", len(rows), run.ID, m, n)

			newCtx, cancel := context.WithTimeout(ctx, time.Second*60)
			defer cancel()

			_, err := sClient.Apply(newCtx, batch)
			if err != nil {
				logger.Errorf("Error writing batch for %d-row run %d: %v", len(rows), run.ID, err)
				return err
			}

			logger.Infof("Wrote batch for %d-row run %d: [%d,%d)", len(rows), run.ID, m, n)
			return nil
		}, retry.Attempts(5), retry.OnRetry(func(n uint, err error) {
			logger.Warningf("Retrying failed batch batch for %d-row run %d: [%d,%d): %v", len(rows), run.ID, m, n, err)
		}))
		if err != nil {
			fatalErrs.Append(err)
		}
	}
	writeRows := func(rows []*spanner.Mutation) *sync.WaitGroup {
		var wg sync.WaitGroup
		var end int
		for end = batchSize; end <= len(rows); end += batchSize {
			wg.Add(1)
			s.Acquire(ctx, 1)
			go writeBatch(s, &wg, rows[0:], end-batchSize, end)
		}
		if end != len(rows) {
			wg.Add(1)
			s.Acquire(ctx, 1)
			logger.Infof("Writing small batch for %d-row run %d: [%d,%d)", len(rows), run.ID, end-batchSize, len(rows))
			go writeBatch(s, &wg, rows[0:], end-batchSize, len(rows))
			logger.Infof("Wrote small batch for %d-row run %d: [%d,%d)", len(rows), run.ID, end-batchSize, len(rows))
		}
		return &wg
	}

	logger.Infof("Writing %d layer-1 rows run %d", len(r1s), run.ID)
	writeRows(r1s).Wait()
	r1s = nil
	if fatalErrs.Len() > 0 {
		logger.Errorf("Spanner push failed to write layer-1 rows for run %v: %v", run.ID, fatalErrs.Slice())
		return
	}

	logger.Infof("Writing %d layer-2 rows run %d", len(r2s), run.ID)
	writeRows(r2s).Wait()
	r2s = nil
	if fatalErrs.Len() > 0 {
		logger.Errorf("Spanner push failed to write layer-2 rows for run %v: %v", run.ID, fatalErrs.Slice())
		return
	}

	logger.Infof("Writing %d layer-3 rows run %d", len(r3s), run.ID)
	writeRows(r3s).Wait()
	r3s = nil
	if fatalErrs.Len() > 0 {
		logger.Errorf("Spanner push failed to write layer-3 rows for run %v: %v", run.ID, fatalErrs.Slice())
		return
	}

	logger.Infof("Wrote batches for %d-row run %d", numRows, run.ID)

	logger.Infof("Spanner push run succeeded: Wrote batches for %d rows", numRows)
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
	count := int64(len(report.Results))
	for _, r := range report.Results {
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
	return count
}

// countSpannerResults counts the number of test results bound to the given
// runID in Cloud Spanner.
func countSpannerResults(ctx context.Context, client *spanner.Client, runID int64) (int64, error) {
	params := map[string]interface{}{
		"RunID": runID,
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

// computeTestKey computes a stable int64 ID for a test+(optional)subtest pair.
func computeTestKey(test string, subtest *string) TestKey {
	key := TestKey{
		TestID: int64(farm.Fingerprint64([]byte(test))),
	}
	if subtest != nil && *subtest != "" {
		key.SubtestID = spanner.NullInt64{
			Int64: int64(farm.Fingerprint64([]byte(*subtest))),
			Valid: true,
		}
	}
	return key
}

// toNullString converts a string pointer to a spanner.NullString, where
// nil is equivalent to the spanner.NullString null value.
func toNullString(s *string) spanner.NullString {
	if s != nil && *s != "" {
		return spanner.NullString{
			StringVal: *s,
			Valid:     true,
		}
	}

	return spanner.NullString{}
}

// toNullInt64 converts a int64 pointer to a spanner.NullInt64, where
// both nil and 0 are equivalent to the spanner.NullInt64 null value.
func toNullInt64(n *int64) spanner.NullInt64 {
	if n != nil && *n != 0 {
		return spanner.NullInt64{
			Int64: *n,
			Valid: true,
		}
	}
	return spanner.NullInt64{}
}
