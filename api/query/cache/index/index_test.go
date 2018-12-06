// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/web-platform-tests/wpt.fyi/api/query"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	metrics "github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestInvalidNumShards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	_, err := NewShardedWPTIndex(loader, 0)
	assert.NotNil(t, err)
	_, err = NewShardedWPTIndex(loader, -1)
}

func TestEvictEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	i, err := NewShardedWPTIndex(loader, 1)
	assert.Nil(t, err)
	assert.NotNil(t, i.EvictAnyRun())
}

func TestIngestRun_zeroID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	i, err := NewShardedWPTIndex(loader, 1)
	assert.Nil(t, err)
	assert.NotNil(t, i.IngestRun(shared.TestRun{ID: 0}))
}

func TestIngestRun_double(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	i, err := NewShardedWPTIndex(loader, 1)
	assert.Nil(t, err)
	run := shared.TestRun{
		ID:            1,
		RawResultsURL: "http://example.com/results.json",
	}
	results := &metrics.TestResultsReport{}
	loader.EXPECT().Load(run).Return(results, nil)
	assert.Nil(t, i.IngestRun(run))
	assert.NotNil(t, i.IngestRun(run))
}

func TestIngestRun_concurrent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	i, err := NewShardedWPTIndex(loader, 1)
	assert.Nil(t, err)
	run := shared.TestRun{
		ID:            1,
		RawResultsURL: "http://example.com/results.json",
	}

	// Wait for 2 goroutines to finish. Gate second goroutine's IngestRun()
	// invocation with channel that receives value after first goroutine has
	// started to ingest the run.
	var wg sync.WaitGroup
	startSecondIngestRun := make(chan bool)
	wg.Add(2)
	go func() {
		defer wg.Done()
		loader.EXPECT().Load(run).DoAndReturn(func(shared.TestRun) (*metrics.TestResultsReport, error) {
			// Now that Load(run) has been invoked, i's implementation should have
			// already marked run as in-flight. Trigger second attempt to ingest run,
			// and pause a little to let it error-out.
			startSecondIngestRun <- true
			time.Sleep(time.Millisecond * 10)
			return &metrics.TestResultsReport{}, nil
		})
		i.IngestRun(run)
	}()
	go func() {
		defer wg.Done()
		<-startSecondIngestRun
		// Expect error during second concurrent attempt to ingest run.
		assert.NotNil(t, i.IngestRun(run))
	}()
	wg.Wait()
}

func TestIngestRun_loaderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	i, err := NewShardedWPTIndex(loader, 1)
	assert.Nil(t, err)
	run := shared.TestRun{
		ID:            1,
		RawResultsURL: "http://example.com/results.json",
	}
	loaderErr := errors.New("Failed to load test results")
	loader.EXPECT().Load(run).Return(nil, loaderErr)
	assert.Equal(t, loaderErr, i.IngestRun(run))
}

func TestEvictNonEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	i, err := NewShardedWPTIndex(loader, 1)
	assert.Nil(t, err)
	run := shared.TestRun{
		ID:            1,
		RawResultsURL: "http://example.com/results.json",
	}
	results := &metrics.TestResultsReport{
		Results: []*metrics.TestResults{
			&metrics.TestResults{
				Test:     "a",
				Status:   "PASS",
				Subtests: []metrics.SubTest{},
			},
			&metrics.TestResults{
				Test:   "b",
				Status: "OK",
				Subtests: []metrics.SubTest{
					metrics.SubTest{
						Name:   "sub",
						Status: "FAIL",
					},
				},
			},
		},
	}
	loader.EXPECT().Load(run).Return(results, nil)
	assert.Nil(t, i.IngestRun(run))
	assert.Nil(t, i.EvictAnyRun())
}

func TestSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	i, err := NewShardedWPTIndex(loader, 1)
	assert.Nil(t, err)

	// Populate data with predictable set of two results for each run.
	loader.EXPECT().Load(gomock.Any()).DoAndReturn(func(run shared.TestRun) (*metrics.TestResultsReport, error) {
		strID := strconv.FormatInt(run.ID, 10)
		strStatus := shared.TestStatusStringFromValue(run.ID % 7)
		return &metrics.TestResultsReport{
			Results: []*metrics.TestResults{
				&metrics.TestResults{
					Test:   "shared",
					Status: strStatus,
				},
				&metrics.TestResults{
					Test:   "test" + strID,
					Status: "PASS",
				},
			},
		}, nil
	}).AnyTimes()

	// Baseline before running things in parallel: Index already contains 8 runs.
	i.IngestRun(makeRun(1))
	i.IngestRun(makeRun(2))
	i.IngestRun(makeRun(3))
	i.IngestRun(makeRun(4))
	i.IngestRun(makeRun(5))
	i.IngestRun(makeRun(6))
	i.IngestRun(makeRun(7))
	i.IngestRun(makeRun(8))

	// Eight times (from run IDs 9 through 16), in parallel:
	// - Evict one run,
	// - Add one run,
	// - Attempt one query (that may fail to bind if it references an already
	//   already evicted run).
	var wg sync.WaitGroup
	for j := 9; j <= 16; j++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			i.EvictAnyRun()
		}(j)
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			i.IngestRun(makeRun(id))
		}(int64(j))
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			runs := []shared.TestRun{
				makeRun(int64(n - 1)),
				makeRun(int64(n - 2)),
				makeRun(int64(n - 3)),
				makeRun(int64(n - 4)),
			}
			plan, err := i.Bind(runs, query.TestStatusConstraint{
				BrowserName: "Chrome",
				Status:      shared.TestStatusPass,
			})
			if err != nil {
				return
			}

			plan.Execute(runs)
		}(j)
	}
	wg.Wait()

	// Number of runs should now be 8 + 8 - 8 = 8.
	// Shards, taken together, should contain data for two predictable run results
	// for each run still in the index. (See loader.EXPECT()...DoAndReturn(...)
	// callback above for predictable test names and values.)

	// TODO: Should Index have a Runs() getter for purposes such as this check?
	idx, ok := i.(*shardedWPTIndex)
	assert.True(t, ok)

	assert.Equal(t, 8, len(idx.runs))
	sharedTestID, err := computeTestID("shared", nil)
	assert.Nil(t, err)
	numResults := 0
	for _, s := range idx.shards {
		// TODO: Should Results have a getter for purposes such as this check?
		results, ok := s.results.(*resultsMap)
		assert.True(t, ok)
		numRuns := 0
		results.byRunTest.Range(func(key, value interface{}) bool {
			numRuns++
			return true
		})
		assert.Equal(t, 8, numRuns)

		for _, run := range idx.runs {
			value, ok := results.byRunTest.Load(RunID(run.ID))
			assert.True(t, ok)
			// TODO: Should Results have a getter for purposes such as this check?
			res, ok := value.(*runResultsMap)
			assert.True(t, ok)

			strID := strconv.FormatInt(run.ID, 10)
			expectedTestID, err := computeTestID("test"+strID, nil)
			assert.Nil(t, err)

			for testID, resultID := range res.byTest {
				// Either test is the "shared test" with varied result values across
				// runs or it is the "test-specific test" with name `test<test ID>` and
				// result value of "PASS".
				assert.True(t, sharedTestID == testID || (expectedTestID == testID && resultID == ResultID(shared.TestStatusPass)))
				numResults++
			}
		}
	}

	// Total number of results is 8 runs * 2 results per run = 16.
	assert.Equal(t, 16, numResults)
}

var browsers = []string{
	"Chrome",
	"Edge",
	"Firefox",
	"Safari",
}

func makeRun(id int64) shared.TestRun {
	browserName := browsers[id%int64(len(browsers))]
	return shared.TestRun{
		ID: id,
		ProductAtRevision: shared.ProductAtRevision{
			Product: shared.Product{
				BrowserName: browserName,
			},
		},
	}
}

// TODO: Add synchronization test to check for race conditions once Bind+Execute
// are fully implemented over indexes and filters.
