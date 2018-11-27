// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package monitor

import (
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func getTestHarness(t *testing.T) (*gomock.Controller, *index.MockIndex, *MockRuntime, Monitor) {
	ctrl := gomock.NewController(t)
	idx := index.NewMockIndex(ctrl)
	rt := NewMockRuntime(ctrl)
	mon := NewIndexMonitor(shared.NewNilLogger(), rt, time.Microsecond, 1, idx)
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
	ctrl, _, rt, mon := getTestHarness(t)
	defer ctrl.Finish()
	rt.EXPECT().GetHeapBytes().Return(uint64(0)).AnyTimes()
	var startErr, stopErr error
	go func() {
		time.Sleep(time.Microsecond * 10)
		stopErr = mon.Stop()
	}()
	startErr = mon.Start()
	assert.Nil(t, stopErr)
	assert.Equal(t, errStopped, startErr)
}

func TestDoubleStart(t *testing.T) {
	ctrl, _, rt, mon := getTestHarness(t)
	defer ctrl.Finish()
	rt.EXPECT().GetHeapBytes().Return(uint64(0)).AnyTimes()
	var err1, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err1 = mon.Start()
	}()
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
	rt.EXPECT().GetHeapBytes().Return(uint64(10))
	rt.EXPECT().GetHeapBytes().Return(uint64(0)).AnyTimes()
	idx.EXPECT().EvictAnyRun().DoAndReturn(func() interface{} {
		go func() { mon.Stop() }()
		return nil
	})
	err := mon.Start()
	assert.Equal(t, errStopped, err)
}
