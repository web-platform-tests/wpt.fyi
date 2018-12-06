// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"errors"
	"sync"
	"testing"
	"time"

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

// TODO: Add synchronization test to check for race conditions once Bind+Execute
// are fully implemented over indexes and filters.
