// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// TODO(Hexcles): Extract type RunID to another package (shared) so that Index
// can be mocked into a different package without cyclic imports.

package index

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sync"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/lru"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/metrics"

	"github.com/sirupsen/logrus"
)

var (
	errNilRun             = errors.New("Test run is nil")
	errNoQuery            = errors.New("No query provided")
	errNoRuns             = errors.New("No runs")
	errRunExists          = errors.New("Run already exists in index")
	errRunLoading         = errors.New("Run currently being loaded into index")
	errSomeShardsRequired = errors.New("Index must have at least one shard")
	errUnexpectedRuns     = errors.New("Unexpected number of runs")
	errZeroRun            = errors.New("Cannot ingest run with ID of 0")
	errEmptyReport        = errors.New("Report contains no results")
)

// ErrRunExists returns the error associated with an attempt to perform
// operations on a run currently unknown to an Index when the Index, in fact,
// already knows about the run.
func ErrRunExists() error {
	return errRunExists
}

// ErrRunLoading returns the error associated with an attempt to perform
// operations on a run currently unknown to an Index when the Index, in fact,
// is currently loading data associated with the run.
func ErrRunLoading() error {
	return errRunLoading
}

// Index is an index of test run results that can ingest and evict runs.
type Index interface {
	query.Binder

	// Run loads the metadata associated with the given RunID value. It returns
	// an error if the Index does not understand the given RunID value.
	Run(RunID) (shared.TestRun, error)
	// Runs loads the metadata associated with the given RunID values. It returns
	// an error if the Index does not understand one or more of the given RunID
	// values.
	Runs([]RunID) ([]shared.TestRun, error)
	// IngestRun loads the test run results associated with the input test run
	// into the index.
	IngestRun(shared.TestRun) error
	// EvictRuns reduces memory pressure by evicting the cache's choice of runs
	// from memory. The parameter is a percentage of current runs to evict.
	EvictRuns(float64) (int, error)
	// SetIndexChan sets the channel that synchronizes before ingesting a run.
	// This channel is used by index monitors to ensure that the monitor is
	// scheduled to run frequently enough to keep pace with any influx of ingested
	// runs.
	SetIngestChan(chan bool)
}

// ProxyIndex is a proxy implementation of the Index interface. This type is
// generally used in type embeddings that wish to override the behaviour of some
// (but not all) methods, deferring to the delegate for all other behaviours.
type ProxyIndex struct {
	delegate Index
}

// Run loads the metadata for the given run ID value by deferring to the
// proxy's delegate.
func (i *ProxyIndex) Run(id RunID) (shared.TestRun, error) {
	return i.delegate.Run(id)
}

// Runs loads the metadata for the given run ID values by deferring to the
// proxy's delegate.
func (i *ProxyIndex) Runs(ids []RunID) ([]shared.TestRun, error) {
	return i.delegate.Runs(ids)
}

// IngestRun loads the given run's results in to the index by deferring to the
// proxy's delegate.
func (i *ProxyIndex) IngestRun(r shared.TestRun) error {
	return i.delegate.IngestRun(r)
}

// EvictRuns deletes percent% runs from the index by deferring to the proxy's
// delegate.
func (i *ProxyIndex) EvictRuns(percent float64) (int, error) {
	return i.delegate.EvictRuns(percent)
}

// SetIngestChan sets the channel that synchronizes before ingesting a run by
// deferring to the proxy's delegate.
func (i *ProxyIndex) SetIngestChan(c chan bool) {
	i.delegate.SetIngestChan(c)
}

// NewProxyIndex instantiates a new proxy index bound to the given delegate.
func NewProxyIndex(idx Index) ProxyIndex {
	return ProxyIndex{idx}
}

// ReportLoader handles loading a WPT test results report based on metadata in
// a shared.TestRun.
type ReportLoader interface {
	Load(shared.TestRun) (*metrics.TestResultsReport, error)
}

// shardedWPTIndex is an Index that manages test and result data across mutually
// exclusive shards.
type shardedWPTIndex struct {
	runs     map[RunID]shared.TestRun
	lru      lru.LRU
	inFlight mapset.Set
	loader   ReportLoader
	shards   []*wptIndex
	m        *sync.RWMutex
	c        chan bool
}

// wptIndex is an index of tests and results. Multicore machines should use
// shardedWPTIndex, which embed a slice of wptIndex containing mutually
// exclusive subsets of test and result data.
type wptIndex struct {
	tests   Tests
	results Results
	m       *sync.RWMutex
}

// testData is a wrapper for a single unit of test+result data from a test run.
type testData struct {
	testName
	ResultID
}

// HTTPReportLoader loads WPT test run reports from the URL specified in test
// run metadata.
type HTTPReportLoader struct{}

func (i *shardedWPTIndex) Run(id RunID) (shared.TestRun, error) {
	return i.syncGetRun(id)
}

func (i *shardedWPTIndex) Runs(ids []RunID) ([]shared.TestRun, error) {
	return i.syncGetRuns(ids)
}

func (i *shardedWPTIndex) IngestRun(r shared.TestRun) error {
	// Error cases: ID cannot be 0, run cannot be loaded or loading-in-progress.
	if r.ID == 0 {
		return errZeroRun
	}

	// Synchronize with anything that may be monitoring run ingestion. Do this
	// before any i.sync* routines to avoid deadlock.
	if i.c != nil {
		i.c <- true
	}

	if err := i.syncMarkInProgress(r); err != nil {
		return err
	}
	defer i.syncClearInProgress(r)

	// Delegate loader to construct complete run report.
	report, err := i.loader.Load(r)
	if err != nil && err != errEmptyReport {
		return err
	}

	// Results of different tests will be stored in different shards, based on the
	// top-level test (i.e., not subtests) integral ID of each test in the report.
	//
	// Create RunResults for each shard's partition of this run's results.
	numShards := len(i.shards)
	numShardsU64 := uint64(numShards)
	shardData := make([]map[TestID]testData, numShards)
	for j := 0; j < numShards; j++ {
		shardData[j] = make(map[TestID]testData)
	}

	for _, res := range report.Results {
		// Add top-level test (i.e., not subtest) result to appropriate shard.
		t, err := computeTestID(res.Test, nil)
		if err != nil {
			return err
		}

		shardIdx := int(t.testID % numShardsU64)
		dataForShard := shardData[shardIdx]
		re := ResultID(shared.TestStatusValueFromString(res.Status))
		dataForShard[t] = testData{
			testName: testName{
				name:    res.Test,
				subName: nil,
			},
			ResultID: re,
		}

		// Dedup subtests, warning when subtest names are duplicated.
		subs := make(map[string]metrics.SubTest)
		for _, sub := range res.Subtests {
			if _, ok := subs[sub.Name]; ok {
				logrus.Warningf("Duplicate subtests with the same name: %s %s", res.Test, sub.Name)
				continue
			}
			subs[sub.Name] = sub
		}

		// Add each subtests' result to the appropriate shard (same shard as
		// top-level test).
		for i := range subs {
			name := subs[i].Name
			t, err := computeTestID(res.Test, &name)
			if err != nil {
				return err
			}

			re := ResultID(shared.TestStatusValueFromString(subs[i].Status))
			dataForShard[t] = testData{
				testName: testName{
					name:    res.Test,
					subName: &name,
				},
				ResultID: re,
			}
		}
	}

	i.syncStoreRun(r, shardData)

	return nil
}

func (i *shardedWPTIndex) EvictRuns(percent float64) (int, error) {
	return i.syncEvictRuns(math.Max(0.0, math.Min(1.0, percent)))
}

func (i *shardedWPTIndex) Bind(runs []shared.TestRun, q query.ConcreteQuery) (query.Plan, error) {
	if len(runs) == 0 {
		return nil, errNoRuns
	} else if q == nil {
		return nil, errNoQuery
	}

	ids := make([]RunID, len(runs))
	for j, run := range runs {
		ids[j] = RunID(run.ID)
	}
	idxs, err := i.syncExtractRuns(ids)
	if err != nil {
		return nil, err
	}

	fs := make(ShardedFilter, len(idxs))
	for j, idx := range idxs {
		f, err := newFilter(idx, q)
		if err != nil {
			return nil, err
		}
		fs[j] = f
	}
	return fs, nil
}

func (i *shardedWPTIndex) SetIngestChan(c chan bool) {
	i.c = c
}

// Load for HTTPReportLoader loads WPT test run reports from the URL specified
// in test run metadata.
func (l HTTPReportLoader) Load(run shared.TestRun) (*metrics.TestResultsReport, error) {
	// Attempt to fetch-and-unmarshal run from run.RawResultsURL.
	resp, err := http.Get(run.RawResultsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Non-OK HTTP status code of %d from "%s" for run ID=%d`, resp.StatusCode, run.RawResultsURL, run.ID)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var report metrics.TestResultsReport
	err = json.Unmarshal(data, &report)
	if err != nil {
		return nil, err
	}
	if len(report.Results) == 0 {
		return &report, errEmptyReport
	}
	return &report, nil
}

// NewShardedWPTIndex creates a new empty Index for WPT test run results.
func NewShardedWPTIndex(loader ReportLoader, numShards int) (Index, error) {
	if numShards <= 0 {
		return nil, errSomeShardsRequired
	}

	shards := make([]*wptIndex, 0, numShards)
	for i := 0; i < numShards; i++ {
		tests := NewTests()
		shards = append(shards, newWPTIndex(tests))
	}

	return &shardedWPTIndex{
		runs:     make(map[RunID]shared.TestRun),
		lru:      lru.NewLRU(),
		inFlight: mapset.NewSet(),
		loader:   loader,
		shards:   shards,
		m:        &sync.RWMutex{},
	}, nil
}

// NewReportLoader constructs a loader that loads result reports over HTTP from
// a shared.TestRun.RawResultsURL.
func NewReportLoader() ReportLoader {
	return HTTPReportLoader{}
}

func (i *shardedWPTIndex) syncGetRun(id RunID) (shared.TestRun, error) {
	i.m.RLock()
	defer i.m.RUnlock()

	run, loaded := i.runs[id]
	if !loaded {
		return shared.TestRun{}, fmt.Errorf("Unknown run ID: %v", id)
	}

	return run, nil
}

func (i *shardedWPTIndex) syncGetRuns(ids []RunID) ([]shared.TestRun, error) {
	i.m.RLock()
	defer i.m.RUnlock()

	runs := make([]shared.TestRun, len(ids))
	for j := range ids {
		run, ok := i.runs[ids[j]]
		if !ok {
			return nil, fmt.Errorf("Unknown run ID: %v", ids[j])
		}

		runs[j] = run
	}
	return runs, nil
}

func (i *shardedWPTIndex) syncMarkInProgress(run shared.TestRun) error {
	i.m.Lock()
	defer i.m.Unlock()

	id := RunID(run.ID)
	_, loaded := i.runs[id]
	if loaded {
		return errRunExists
	}
	if i.inFlight.Contains(id) {
		return errRunLoading
	}

	i.inFlight.Add(id)

	return nil
}

func (i *shardedWPTIndex) syncClearInProgress(run shared.TestRun) error {
	i.m.Lock()
	defer i.m.Unlock()

	id := RunID(run.ID)
	if !i.inFlight.Contains(id) {
		return errNilRun
	}

	i.inFlight.Remove(id)

	return nil
}

func (i *shardedWPTIndex) syncStoreRun(run shared.TestRun, data []map[TestID]testData) error {
	i.m.Lock()
	defer i.m.Unlock()

	id := RunID(run.ID)
	for j, shardData := range data {
		if err := syncStoreRunOnShard(i.shards[j], id, shardData); err != nil {
			return err
		}
	}
	i.runs[id] = run
	i.lru.Access(int64(id))

	return nil
}

func syncStoreRunOnShard(shard *wptIndex, id RunID, shardData map[TestID]testData) error {
	shard.m.Lock()
	defer shard.m.Unlock()

	runResults := NewRunResults()
	for t, data := range shardData {
		shard.tests.Add(t, data.testName.name, data.testName.subName)
		runResults.Add(data.ResultID, t)
	}
	return shard.results.Add(id, runResults)
}

func (i *shardedWPTIndex) syncEvictRuns(percent float64) (int, error) {
	i.m.Lock()
	defer i.m.Unlock()

	if len(i.runs) == 0 {
		return 0, errNoRuns
	}

	runIDs := i.lru.EvictLRU(percent)
	if len(runIDs) == 0 {
		return 0, errNoRuns
	}

	for _, runID := range runIDs {
		id := RunID(runID)

		// Delete data from shards, and from runs collection.
		for _, shard := range i.shards {
			if err := syncDeleteResultsFromShard(shard, id); err != nil {
				return 0, err
			}
		}
		delete(i.runs, id)
	}

	return len(runIDs), nil
}

func syncDeleteResultsFromShard(shard *wptIndex, id RunID) error {
	shard.m.Lock()
	defer shard.m.Unlock()

	return shard.results.Delete(id)
}

func (i *shardedWPTIndex) syncExtractRuns(ids []RunID) ([]index, error) {
	i.m.RLock()
	defer i.m.RUnlock()

	idxs := make([]index, len(i.shards))
	var err error
	for j, shard := range i.shards {
		idxs[j], err = syncMakeIndex(shard, ids)
		if err != nil {
			return nil, err
		}
	}

	for _, id := range ids {
		i.lru.Access(int64(id))
	}

	return idxs, nil
}

func syncMakeIndex(shard *wptIndex, ids []RunID) (index, error) {
	shard.m.RLock()
	defer shard.m.RUnlock()

	tests := shard.tests
	runResults := make(map[RunID]RunResults)
	for _, id := range ids {
		rrs := shard.results.ForRun(id)
		if rrs == nil {
			return index{}, fmt.Errorf("Run is unknown to shard: RunID=%v", id)
		}
		runResults[id] = shard.results.ForRun(id)
	}
	return index{
		tests:      tests,
		runResults: runResults,
		m:          shard.m,
	}, nil
}

func newWPTIndex(tests Tests) *wptIndex {
	return &wptIndex{
		tests:   tests,
		results: NewResults(),
		m:       &sync.RWMutex{},
	}
}
