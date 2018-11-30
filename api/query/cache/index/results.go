// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"fmt"
	"sync"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// RunID is a unique identifier for a WPT test run. These IDs are generated when
// a run is committed to Datastore.
type RunID int64

// ResultID is a unique identifier for a WPT test result/status. RunID values
// are documented in github.com/web-platform-tests/wpt.fyi/shared TestStatus*
// values.
type ResultID int64

// Results is an interface for an index that stores RunID => RunResults
// mappings.
type Results interface {
	// Add stores a RunID => RunResults mapping.
	Add(RunID, RunResults) error
	// Delete deletes all result data for a particular WPT test run.
	Delete(RunID) error
	// ForRun produces a RunResults interface for a particular WPT test run, or
	// nil if the input RunID is unknown to this index.
	ForRun(RunID) RunResults

	// TODO: Add filter binding function:
	// ResultFilter(ru RunID, re ResultID) UnboundFilter
}

// RunResults is an interface for an index that stores a TestID => ResultID
// mapping.
type RunResults interface {
	// Add stores a TestID => ResultID mapping.
	Add(ResultID, TestID)
	// GetResult looks up the ResultID associated with a TestID; the
	// "status unknown" value is used if the lookup yields no ResultID.
	GetResult(TestID) ResultID
}

type resultsMap struct {
	byRunTest sync.Map
}

type runResultsMap struct {
	byTest map[TestID]ResultID
}

// NewResults generates a new empty results index.
func NewResults() Results {
	return &resultsMap{byRunTest: sync.Map{}}
}

// NewRunResults generates a new empty run results index.
func NewRunResults() RunResults {
	return &runResultsMap{make(map[TestID]ResultID)}
}

func (rs *resultsMap) Add(ru RunID, rr RunResults) error {
	_, wasLoaded := rs.byRunTest.LoadOrStore(ru, rr)
	if wasLoaded {
		return fmt.Errorf("Already loaded into results index: %v", ru)
	}

	return nil
}

func (rs *resultsMap) Delete(ru RunID) error {
	_, ok := rs.byRunTest.Load(ru)
	if !ok {
		return fmt.Errorf(`No such run in results index; run ID: %v`, ru)
	}

	rs.byRunTest.Delete(ru)
	return nil
}

func (rs *resultsMap) ForRun(ru RunID) RunResults {
	v, ok := rs.byRunTest.Load(ru)
	if !ok {
		return nil
	}
	rrm := v.(*runResultsMap)
	return rrm
}

func (rrs *runResultsMap) Add(re ResultID, t TestID) {
	rrs.byTest[t] = re
}

func (rrs *runResultsMap) GetResult(t TestID) ResultID {
	re, ok := rrs.byTest[t]
	if !ok {
		return ResultID(shared.TestStatusUnknown)
	}
	return re
}

// TODO: Add filter binding function:
// func ResultFilter(ru RunID, re ResultID) UnboundFilter {
// 	return NewResultEQFilter(ru, re)
// }
