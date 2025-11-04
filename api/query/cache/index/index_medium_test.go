//go:build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	metrics "github.com/web-platform-tests/wpt.fyi/shared/metrics"
	"go.uber.org/mock/gomock"
)

const (
	testEvictionNumResults = 10000
	// Each result should require at least one byte for its name + 4 bytes for
	// "PASS".
	testEvictionMinBytes = testEvictionNumResults * 5
)

func TestEvictAnyRunRelievesMemoryPressure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	loader := NewMockReportLoader(ctrl)
	i, err := NewShardedWPTIndex(loader, 1)
	assert.Nil(t, err)
	run := shared.TestRun{
		ID:            1,
		RawResultsURL: "http://example.com/results.json",
	}
	results := make([]*metrics.TestResults, 0, testEvictionNumResults)
	for j := 0; j < testEvictionNumResults; j++ {
		str := strconv.Itoa(j)
		results = append(results, &metrics.TestResults{
			Test:   str,
			Status: "PASS",
		})
	}
	report := &metrics.TestResultsReport{Results: results}
	loader.EXPECT().Load(run).Return(report, nil)

	assert.Nil(t, i.IngestRun(run))
	results = nil
	runtime.GC()

	var stats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&stats)
	baseline := stats.HeapAlloc

	n, err := i.EvictRuns(0.0)
	assert.Nil(t, err)
	assert.Equal(t, 1, n)

	runtime.GC()
	runtime.ReadMemStats(&stats)
	assert.True(t, baseline-testEvictionMinBytes > stats.HeapAlloc)
}
