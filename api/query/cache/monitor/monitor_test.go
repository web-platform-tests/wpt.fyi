//go:build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package monitor

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

const testMaxHeapBytes uint64 = 10

func getTestHarness(t *testing.T) (*gomock.Controller, *index.MockIndex, *MockRuntime, Monitor) {
	ctrl := gomock.NewController(t)
	idx := index.NewMockIndex(ctrl)
	rt := NewMockRuntime(ctrl)
	mon, err := NewIndexMonitor(shared.NewNilLogger(), rt, time.Microsecond, 10, testMaxHeapBytes, 0.0, idx)
	assert.Nil(t, err)
	return ctrl, idx, rt, mon
}

func TestStopErr(t *testing.T) {
	ctrl, _, rt, mon := getTestHarness(t)
	defer ctrl.Finish()
	rt.EXPECT().GetHeapBytes().Return(uint64(0)).AnyTimes()
	err := mon.Stop()
	assert.Equal(t, errStopped, err)
}

func TestStartStop(t *testing.T) {
	ctrl, idx, rt, mon := getTestHarness(t)
	defer ctrl.Finish()
	rt.EXPECT().GetHeapBytes().Return(uint64(0)).AnyTimes()
	var startErr, stopErr error
	go func() {
		time.Sleep(time.Microsecond * 10)
		stopErr = mon.Stop()
	}()
	idx.EXPECT().SetIngestChan(gomock.Any())
	startErr = mon.Start()
	assert.Nil(t, stopErr)
	assert.Equal(t, errStopped, startErr)
}

func TestDoubleStart(t *testing.T) {
	ctrl, idx, rt, mon := getTestHarness(t)
	defer ctrl.Finish()
	rt.EXPECT().GetHeapBytes().Return(uint64(0)).AnyTimes()
	idx.EXPECT().SetIngestChan(gomock.Any())
	var err1, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err1 = mon.Start()
	}()
	time.Sleep(time.Microsecond * 100)
	go func() {
		defer wg.Done()
		err2 = mon.Start()
	}()
	time.Sleep(time.Microsecond * 100)
	mon.Stop()
	wg.Wait()
	assert.True(t, (err1 == errStopped && err2 == errRunning) || (err1 == errRunning && err2 == errStopped))
}

func TestOOM(t *testing.T) {
	ctrl, idx, rt, mon := getTestHarness(t)
	defer ctrl.Finish()
	rt.EXPECT().GetHeapBytes().Return(uint64(0))
	rt.EXPECT().GetHeapBytes().Return(uint64(0))
	rt.EXPECT().GetHeapBytes().Return(uint64(testMaxHeapBytes + 1))
	rt.EXPECT().GetHeapBytes().Return(uint64(0)).AnyTimes()
	idx.EXPECT().EvictRuns(gomock.Any()).DoAndReturn(func(float64) (int, error) {
		go func() { mon.Stop() }()
		return 0, nil
	})
	idx.EXPECT().SetIngestChan(gomock.Any())
	err := mon.Start()
	assert.Equal(t, errStopped, err)
}

type syncingIndex struct {
	index.ProxyIndex

	IngestsStarted   int
	IngestsCompleted int
	c                chan bool
}

func (i *syncingIndex) SetIngestChan(c chan bool) {
	i.c = c
}

func (i *syncingIndex) IngestRun(r shared.TestRun) error {
	i.IngestsStarted++
	if i.c != nil {
		i.c <- true
	}
	i.IngestsCompleted++

	return nil
}

func TestIngestTriggered(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	idx := index.NewMockIndex(ctrl)
	rt := NewMockRuntime(ctrl)

	// Long timeout of 1 minute and low "max ingestions before running monitor" of
	// 2.
	mon, err := NewIndexMonitor(shared.NewNilLogger(), rt, time.Minute, 2, testMaxHeapBytes, 0.0, idx)
	assert.Nil(t, err)

	// Send to done when all goroutines expectations should have been checked.
	done := make(chan bool)

	// Wait for monitor to set index's ingest run chan, then ingest two runs,
	// triggering monitor check (and call to rt.GetHeapBytes()).
	var c chan bool
	idx.EXPECT().SetIngestChan(gomock.Any()).DoAndReturn(func(ch chan bool) {
		c = ch

		go func() {
			idx.IngestRun(shared.TestRun{})
			idx.IngestRun(shared.TestRun{})

			// Second IngestRun() should have already triggered monitor (and
			// rt.GetHeapBytes()) to run.
			done <- true
		}()
	})

	// Track number of ingest requests started and finished. Sync over c.
	ingestStarted := 0
	ingestFinished := 0
	idx.EXPECT().IngestRun(gomock.Any()).DoAndReturn(func(shared.TestRun) error {
		ingestStarted++
		c <- true
		ingestFinished++
		return nil
	}).AnyTimes()

	// Start the monitor (triggering idx.SetIngestChan()).
	go mon.Start()

	// GetHeapBytes() is called to monitor memory usage when limit=2 IngestRun
	// calls have started and monitor hasn't run yet. (Monitor should not run due
	// to timeout set to 1 minute.)
	rt.EXPECT().GetHeapBytes().DoAndReturn(func() uint64 {
		assert.Equal(t, 2, ingestStarted)
		assert.Equal(t, 1, ingestFinished)
		return uint64(0)
	})

	<-done
}
