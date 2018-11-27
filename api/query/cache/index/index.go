// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"errors"
	"sort"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

var (
	errNoRuns = errors.New("No runs")
	errNilRun = errors.New("Test run is nil")
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

type wptIndex struct {
	runs shared.TestRuns
}

func (i *wptIndex) IngestRun(r shared.TestRun) error {
	i.runs = append(i.runs, r)
	sort.Sort(i.runs)
	return nil
}

func (i *wptIndex) EvictAnyRun() error {
	if len(i.runs) == 0 {
		return errNoRuns
	}
	i.runs = i.runs[1:]
	return nil
}

// NewWPTIndex creates a new empty Index for WPT test run results.
func NewWPTIndex() Index {
	return &wptIndex{make(shared.TestRuns, 0)}
}
