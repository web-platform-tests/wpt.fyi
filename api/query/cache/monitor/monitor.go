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
	errStopped         = errors.New("monitor stopped")
	errRunning         = errors.New("monitor running")
	errNegativePercent = errors.New("invalid percentage (negative)")
	errPercentTooLarge = errors.New("invalid percentage (greater than 1.00)")
)

// Runtime is a wrapper for the go runtime package. It allows tests to mock
// runtime characteristics that code under test may monitor.
type Runtime interface {
	// GetHeapBytes reports the current number of heap allocated bytes.
	GetHeapBytes() uint64
}

// Monitor is an interface responsible for monitoring runtime conditions.
type Monitor interface {
	// Start starts the monitor, blocking until the monitor stops.
	Start() error
	// Stop stops the monitor.
	Stop() error
	// SetInterval sets the interval at which the monitor polls runtime state.
	SetInterval(time.Duration) error
	// SetMaxHeapBytes sets the limit on heap allocated bytes before attempting to
	// relieve memory pressure.
	SetMaxHeapBytes(uint64) error
	// SetEvictionPercent sets the percentage of runs to be evicted when the soft
	// memory limit (max heap bytes) is reached.
	SetEvictionPercent(float64) error
}

// ProxyMonitor is a proxy implementation of the Monitor interface. This type is
// generally used in type embeddings that wish to override the behaviour of some
// (but not all) methods, deferring to the delegate for all other behaviours.
type ProxyMonitor struct {
	delegate Monitor
}

// Start initiates monitoring by deferring to the proxy's delegate.
func (m *ProxyMonitor) Start() error {
	return m.delegate.Start()
}

// Stop halts monitoring by deferring to the proxy's delegate.
func (m *ProxyMonitor) Stop() error {
	return m.delegate.Stop()
}

// SetInterval changes the interval at which monitoring operations are performed
// by deferring to the proxy's delegated.
func (m *ProxyMonitor) SetInterval(i time.Duration) error {
	return m.delegate.SetInterval(i)
}

// SetMaxHeapBytes sets the soft limit on heap memory usage by deferring to the
// proxy's delegated.
func (m *ProxyMonitor) SetMaxHeapBytes(b uint64) error {
	return m.delegate.SetMaxHeapBytes(b)
}

// SetEvictionPercent sets the percentage of runs to be evicted when the soft
// memory limit (max heap bytes) is reached by deferring to the proxy's
// delegate.
func (m *ProxyMonitor) SetEvictionPercent(percent float64) error {
	return m.delegate.SetEvictionPercent(percent)
}

// NewProxyMonitor instantiates a new proxy monitor bound to the given delegate.
func NewProxyMonitor(m Monitor) ProxyMonitor {
	return ProxyMonitor{m}
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
	logger          shared.Logger
	rt              Runtime
	interval        time.Duration
	maxIngestedRuns uint
	maxHeapBytes    uint64
	percent         float64

	isRunning   bool
	mutex       *sync.Mutex
	runIngested chan bool

	idx index.Index
}

func (m *indexMonitor) Start() error {
	err := m.start()
	if err != nil {
		return err
	}

	m.idx.SetIngestChan(m.runIngested)

	reqs := uint(0)
	for {
		if !m.isRunning {
			return errStopped
		}

		timer := make(chan bool, 1)
		go func() {
			time.Sleep(m.interval)
			timer <- true
		}()

		for {
			select {
			case <-timer:
				reqs = 0
			case <-m.runIngested:
				reqs++
				if reqs == m.maxIngestedRuns {
					reqs = 0
				}
			}
			if reqs == 0 {
				break
			}
		}

		m.check()
	}
}

func (m *indexMonitor) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if !m.isRunning {
		return errStopped
	}
	m.isRunning = false
	m.runIngested = make(chan bool)

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

func (m *indexMonitor) SetEvictionPercent(percent float64) error {
	if percent < 0 {
		return errNegativePercent
	} else if percent > 1.0 {
		return errPercentTooLarge
	}

	m.percent = percent

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

func (m *indexMonitor) check() {
	heapBytes := m.rt.GetHeapBytes()
	if heapBytes > m.maxHeapBytes {
		m.logger.Warningf("Monitor %d bytes allocated, exceeding threshold of %d bytes", heapBytes, m.maxHeapBytes)
		if _, err := m.idx.EvictRuns(m.percent); err != nil {
			m.logger.Warningf("Error occurred while evicting %f%% of current runs: %w", m.percent, err)
		}
	} else {
		m.logger.Debugf("Monitor: %d heap-allocated bytes OK", heapBytes)
	}
}

// NewIndexMonitor instantiates a new index.Index monitor.
// nolint:ireturn // TODO: Fix ireturn lint error
func NewIndexMonitor(
	logger shared.Logger,
	rt Runtime, interval time.Duration,
	maxIngestedRuns uint,
	maxHeapBytes uint64,
	percent float64,
	idx index.Index,
) (Monitor, error) {
	if percent < 0 {
		return nil, errNegativePercent
	} else if percent > 1.0 {
		return nil, errPercentTooLarge
	}

	return &indexMonitor{
		logger,
		rt,
		interval,
		maxIngestedRuns,
		maxHeapBytes,
		percent,
		false,
		&sync.Mutex{},
		make(chan bool),
		idx,
	}, nil
}
