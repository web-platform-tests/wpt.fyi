// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"

	"github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared"

	log "github.com/sirupsen/logrus"
)

var (
	errNoRuns             = errors.New("No runs")
	errNilRun             = errors.New("Test run is nil")
	errRunExists          = errors.New("Run already exists in index")
	errRunLoading         = errors.New("Run currently being loaded into index")
	errSomeShardsRequired = errors.New("Index must have at least one shard")
	errZeroRun            = errors.New("Cannot ingest run with ID of 0")
)

// Index is an index of test run results that can ingest and evict runs.
// FUTURE: Index will also be able to service queries.
type Index interface {
	// IngestRun loads the test run results associated with the input test run
	// into the index.
	IngestRun(shared.TestRun) error
	// EvictAnyRun reduces memory pressure by evicting the cache's choice of run
	// from memory.
	EvictAnyRun() error
}

// ProxyIndex is a proxy implementation of the Index interface.
type ProxyIndex struct {
	delegate Index
}

// IngestRun is delegated.
func (i *ProxyIndex) IngestRun(r shared.TestRun) error {
	return i.delegate.IngestRun(r)
}

// EvictAnyRun is delegated.
func (i *ProxyIndex) EvictAnyRun() error {
	return i.delegate.EvictAnyRun()
}

func NewProxyIndex(idx Index) ProxyIndex {
	return ProxyIndex{idx}
}

// ReportLoader handles loading a WPT test results report based on metadata in
// a shared.TestRun.
type ReportLoader interface {
	Load(shared.TestRun) (*metrics.TestResultsReport, error)
}

type shardedWPTIndex struct {
	runs     sync.Map
	inFlight sync.Map
	loader   ReportLoader
	tests    Tests
	shards   []*wptIndex
}

type wptIndex struct {
	tests   Tests
	results Results
}

type httpReportLoader struct{}

func (i *shardedWPTIndex) IngestRun(r shared.TestRun) error {
	// Error cases: ID cannot be 0, run cannot be loaded or loading-in-progress.
	if r.ID == 0 {
		return errZeroRun
	}
	id := RunID(r.ID)
	_, wasLoaded := i.runs.Load(id)
	if wasLoaded {
		return errRunExists
	}
	// LoadOrStore will mark this run as in-flight if it is not already marked as
	// such.
	_, isLoading := i.inFlight.LoadOrStore(id, r)
	defer i.inFlight.Delete(id)
	if isLoading {
		return errRunLoading
	}

	// Delegate loader to construct complete run report.
	report, err := i.loader.Load(r)
	if err != nil {
		return err
	}

	// Results of different tests will be stored in different shards, based on the
	// top-level test (i.e., not subtests) integral ID of each test in the report.
	//
	// Create RunResults for each shard's partition of this run's results.
	numShards := uint64(len(i.shards))
	runResults := make([]RunResults, numShards)
	for j := uint64(0); j < numShards; j++ {
		runResults[j] = NewRunResults()
	}

	for _, res := range report.Results {
		// Add top-level test (i.e., not subtest) result to appropriate shard.
		re := ResultID(metrics.TestStatusFromString(res.Status))
		t, err := i.tests.Add(res.Test, nil)
		if err != nil {
			return err
		}
		runResultsForShard := runResults[int(t.testID%numShards)]
		runResultsForShard.Add(re, t)

		// Dedup subtests, warning when subtest names are duplicated.
		subs := make(map[string]metrics.SubTest)
		for _, sub := range res.Subtests {
			if _, ok := subs[sub.Name]; ok {
				log.Warningf("Duplicate subtests with the same name: %s %s", res.Test, sub.Name)
				continue
			}
			subs[sub.Name] = sub
		}

		// Add each subtests' result to the appropriate shard (same shard as
		// top-level test).
		for _, sub := range subs {
			re := ResultID(metrics.SubTestStatusFromString(sub.Status))
			t, err := i.tests.Add(res.Test, &sub.Name)
			if err != nil {
				return err
			}

			runResultsForShard.Add(re, t)
		}
	}

	// Add accumulated RunResults to shards.
	for j := range runResults {
		err := i.shards[j].results.Add(id, runResults[j])
		if err != nil {
			return err
		}
	}

	// Mark run as successfully stored in index.
	i.runs.Store(id, r)

	return nil
}

func (i *shardedWPTIndex) EvictAnyRun() error {
	// Accumulate runs into sortable collection.
	runs := make(shared.TestRuns, 0)
	i.runs.Range(func(key, value interface{}) bool {
		run := value.(shared.TestRun)
		runs = append(runs, run)
		return true
	})
	if len(runs) == 0 {
		return errNoRuns
	}

	// Sort and mark and select oldest run for eviction.
	sort.Sort(runs)
	runIDToEvict := RunID(runs[0].ID)

	// Delete run from runs-in-index collection.
	i.runs.Delete(runIDToEvict)

	// Delete run results from each shard.
	for j, shard := range i.shards {
		if err := shard.results.Delete(runIDToEvict); err != nil {
			log.Warningf(`Error while evicting run from shard %d: %v`, j, err)
		}
	}

	return nil
}

func (l httpReportLoader) Load(run shared.TestRun) (*metrics.TestResultsReport, error) {
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
		return nil, fmt.Errorf("Empty report from run ID=%d (%s)", run.ID, run.RawResultsURL)
	}
	return &report, nil
}

// NewShardedWPTIndex creates a new empty Index for WPT test run results.
func NewShardedWPTIndex(loader ReportLoader, numShards int) (Index, error) {
	if numShards <= 0 {
		return nil, errSomeShardsRequired
	}

	shards := make([]*wptIndex, 0, numShards)
	tests := NewTests()
	for i := 0; i < numShards; i++ {
		shards = append(shards, newWPTIndex(tests))
	}
	return &shardedWPTIndex{
		runs:     sync.Map{},
		inFlight: sync.Map{},
		loader:   loader,
		tests:    tests,
		shards:   shards,
	}, nil
}

// NewReportLoader constructs a loader that loads result reports over HTTP from
// a shared.TestRun.RawResultsURL.
func NewReportLoader() ReportLoader {
	return httpReportLoader{}
}

func newWPTIndex(tests Tests) *wptIndex {
	return &wptIndex{
		tests:   tests,
		results: NewResults(),
	}
}
