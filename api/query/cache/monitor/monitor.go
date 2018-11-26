// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package monitor

import (
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

var (
	errStopped = errors.New("Monitor stopped")
	errRunning = errors.New("Monitor running")
)

// Runtime is a wrapper for the go runtime package. It allows tests to mock
// runtime characteristics that code under test may monitor.
type Runtime interface {
	// GetHeapBytes reports the current number of heap allocated bytes.
	GetHeapBytes() uint64
}

// Monitor is an interface responsible for monitoring runtime conditions.
type Monitor interface {
	// Start the monitor; block until the monitor stops.
	Start() error
	// Stop the monitor.
	Stop() error
	// Set the interval at which the monitor polls runtime state.
	SetInterval(time.Duration) error
	// Set the limit on heap allocated bytes before attempting to relieve memory
	// pressure.
	SetMaxHeapBytes(uint64) error
}

// GoRuntime is the live go runtime implementation of Runtime.
type GoRuntime struct{}

// GetHeapBytes reports the current number of heap allocated bytes.
func (r GoRuntime) GetHeapBytes() uint64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.HeapAlloc
}

type indexMonitor struct {
	logger       shared.Logger
	rt           Runtime
	interval     time.Duration
	maxHeapBytes uint64

	isRunning bool
	mutex     *sync.Mutex

	idx index.Index
}

func (m *indexMonitor) Start() error {
	err := m.start()
	if err != nil {
		return err
	}

	for {
		if !m.isRunning {
			return errStopped
		}

		heapBytes := m.rt.GetHeapBytes()
		if heapBytes > m.maxHeapBytes {
			m.logger.Errorf("Out of memory: %d > %d", heapBytes, m.maxHeapBytes)
			m.idx.EvictAnyRun()
		} else {
			m.logger.Debugf("Monitor: %d heap-allocated bytes OK", heapBytes)
		}
		time.Sleep(m.interval)
	}
}

func (m *indexMonitor) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if !m.isRunning {
		return errStopped
	}
	m.isRunning = false
	return nil
}

func (m *indexMonitor) SetInterval(interval time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.interval = interval
	return nil
}

func (m *indexMonitor) SetMaxHeapBytes(maxHeapBytes uint64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.maxHeapBytes = maxHeapBytes
	return nil
}

func (m *indexMonitor) start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.isRunning {
		return errRunning
	}
	m.isRunning = true
	return nil
}

// NewIndexMonitor instantiates a new index.Index monitor.
func NewIndexMonitor(logger shared.Logger, rt Runtime, interval time.Duration, maxHeapBytes uint64, idx index.Index) Monitor {
	return &indexMonitor{logger, rt, interval, maxHeapBytes, false, &sync.Mutex{}, idx}
}
