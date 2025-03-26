//go:build medium
// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package backfill

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/monitor"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

type countingIndex struct {
	index.ProxyIndex

	count int
}

var errNotImplemented = errors.New("not implemented")

func (i *countingIndex) IngestRun(r shared.TestRun) error {
	err := i.ProxyIndex.IngestRun(r)
	if err != nil {
		return err
	}

	i.count++
	return nil
}

func (i *countingIndex) EvictRuns(percent float64) (int, error) {
	n, err := i.ProxyIndex.EvictRuns(percent)
	if err != nil {
		return n, err
	}

	i.count -= n
	return n, nil
}

func (*countingIndex) Bind([]shared.TestRun, query.ConcreteQuery) (query.Plan, error) {
	return nil, errNotImplemented
}

func TestStopImmediately(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := sharedtest.NewMockDatastore(ctrl)
	query := sharedtest.NewMockTestRunQuery(ctrl)
	product, _ := shared.ParseProductSpec("chrome")
	store.EXPECT().TestRunQuery().Return(query)
	query.EXPECT().LoadTestRuns(gomock.Any(), nil, nil, nil, nil, gomock.Any(), nil).Return(shared.TestRunsByProduct{
		shared.ProductTestRuns{Product: product, TestRuns: shared.TestRuns{
			shared.TestRun{ID: 1},
			shared.TestRun{ID: 2},
			shared.TestRun{ID: 3},
			shared.TestRun{ID: 4},
		}},
	}, nil)
	rt := monitor.NewMockRuntime(ctrl)
	rt.EXPECT().GetHeapBytes().Return(uint64(0)).AnyTimes()
	mockIdx := index.NewMockIndex(ctrl)
	mockIdx.EXPECT().IngestRun(gomock.Any()).Return(nil).AnyTimes()
	mockIdx.EXPECT().SetIngestChan(gomock.Any())
	idx := countingIndex{index.NewProxyIndex(mockIdx), 0}
	m, err := FillIndex(store, shared.NewNilLogger(), rt, time.Millisecond*10, 10, 1, 0.0, &idx)
	assert.Nil(t, err)
	m.Stop()
	time.Sleep(time.Second)
	assert.True(t, idx.count == 0 || idx.count == 1)
}

func TestIngestSomeRuns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := sharedtest.NewMockDatastore(ctrl)
	query := sharedtest.NewMockTestRunQuery(ctrl)
	product, _ := shared.ParseProductSpec("chrome")
	store.EXPECT().TestRunQuery().Return(query)
	query.EXPECT().LoadTestRuns(gomock.Any(), nil, nil, nil, nil, gomock.Any(), nil).Return(shared.TestRunsByProduct{
		shared.ProductTestRuns{
			Product: product,
			TestRuns: shared.TestRuns{
				shared.TestRun{ID: 1},
				shared.TestRun{ID: 2},
				shared.TestRun{ID: 3},
				shared.TestRun{ID: 4},
			},
		},
	}, nil)

	freq := time.Millisecond * 10
	maxIngestedRuns := uint(10)
	maxBytes := uint64(1)
	rt := monitor.NewMockRuntime(ctrl)

	mockIdx := index.NewMockIndex(ctrl)
	idx := countingIndex{index.NewProxyIndex(mockIdx), 0}

	rt.EXPECT().GetHeapBytes().DoAndReturn(func() uint64 {
		// Trigger monitor when 3 or more runs are loaded.
		if idx.count >= 3 {
			return maxBytes + 1
		} else {
			return uint64(0)
		}
	}).AnyTimes()

	mockIdx.EXPECT().IngestRun(gomock.Any()).DoAndReturn(func(shared.TestRun) error {
		// Wait 2x monitor frequency to allow monitor to halt ingesting runs,
		// if necessary.
		time.Sleep(freq * 2)
		return nil
	}).AnyTimes()

	mockIdx.EXPECT().EvictRuns(gomock.Any()).Return(1, nil).AnyTimes()

	mockIdx.EXPECT().SetIngestChan(gomock.Any())
	m, err := FillIndex(store, shared.NewNilLogger(), rt, freq, maxIngestedRuns, maxBytes, 0.0, &idx)
	assert.Nil(t, err)
	defer m.Stop()

	// Wait for runs to be ingested.
	time.Sleep(time.Second)
	assert.Equal(t, 2, idx.count)
}
