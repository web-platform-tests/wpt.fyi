// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package backfill

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/monitor"
	shared "github.com/web-platform-tests/wpt.fyi/shared"
)

type countingIndex struct {
	index.ProxyIndex

	count int
}

func (i *countingIndex) IngestRun(r shared.TestRun) error {
	err := i.ProxyIndex.IngestRun(r)
	if err != nil {
		return err
	}

	i.count++
	return nil
}

func (i *countingIndex) EvictAnyRun() error {
	err := i.ProxyIndex.EvictAnyRun()
	if err != nil {
		return err
	}

	i.count--
	return nil
}

func TestStopImmediately(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	fetcher := NewMockRunFetcher(ctrl)
	fetcher.EXPECT().FetchRuns(gomock.Any()).Return([]shared.TestRun{
		shared.TestRun{ID: 1},
		shared.TestRun{ID: 2},
		shared.TestRun{ID: 3},
		shared.TestRun{ID: 4},
	}, nil)
	rt := monitor.NewMockRuntime(ctrl)
	rt.EXPECT().GetHeapBytes().Return(uint64(0)).AnyTimes()
	mockIdx := index.NewMockIndex(ctrl)
	mockIdx.EXPECT().IngestRun(gomock.Any()).Return(nil).AnyTimes()
	idx := countingIndex{index.NewProxyIndex(mockIdx), 0}
	m, err := FillIndex(fetcher, shared.NewNilLogger(), rt, time.Millisecond*10, 1, &idx)
	assert.Nil(t, err)
	m.Stop()
	time.Sleep(time.Second)
	assert.True(t, idx.count == 0 || idx.count == 1)
}

func TestIngestSomeRuns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fetcher := NewMockRunFetcher(ctrl)
	fetcher.EXPECT().FetchRuns(gomock.Any()).Return([]shared.TestRun{
		shared.TestRun{ID: 1},
		shared.TestRun{ID: 2},
		shared.TestRun{ID: 3},
		shared.TestRun{ID: 4},
	}, nil)

	freq := time.Millisecond * 10
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

	mockIdx.EXPECT().EvictAnyRun().Return(nil).AnyTimes()

	m, err := FillIndex(fetcher, shared.NewNilLogger(), rt, freq, maxBytes, &idx)
	assert.Nil(t, err)
	defer m.Stop()

	// Wait for runs to be ingested.
	time.Sleep(time.Second)
	assert.Equal(t, 2, idx.count)
}
