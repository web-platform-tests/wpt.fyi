// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

type aggregator interface {
	Add(t TestID) error
	Done() []query.SearchResult
}

type indexAggregator struct {
	index

	runIDs []RunID
	agg    map[uint64]query.SearchResult
	opts   query.AggregationOpts
}

func (a *indexAggregator) Add(t TestID) error {
	id := t.testID
	ts := a.tests
	r, ok := a.agg[id]
	if !ok {
		name, _, err := ts.GetName(t)
		if err != nil {
			return err
		}

		r = query.SearchResult{
			Test:         name,
			LegacyStatus: nil,
			Interop:      nil,
			Diff:         nil,
		}
	}

	if a.opts.InteropFormat {
		if r.Interop == nil {
			r.Interop = make([]int, len(a.runIDs)+1)
		}
		passing := 0
		for _, id := range a.runIDs {
			res := shared.TestStatus(a.runResults[id].GetResult(t))
			if res.IsPassOrOK() {
				passing++
			}
		}
		r.Interop[passing]++
	}

	results := r.LegacyStatus
	if results == nil {
		results = make([]query.LegacySearchRunResult, len(a.runIDs))
	}

	for i, id := range a.runIDs {
		res := shared.TestStatus(a.runResults[id].GetResult(t))
		// TODO: Switch to a consistent value for Total across all runs.
		//
		// Only include tests with non-UNKNOWN status for this run's total.
		if res != shared.TestStatusUnknown {
			results[i].Total++
			if res.IsPassOrOK() {
				results[i].Passes++
			}
		}
	}
	if a.opts.IncludeSubtests {
		if _, subtest, err := ts.GetName(t); err == nil && subtest != nil {
			name := *subtest
			r.Subtests = append(r.Subtests, name)
		}
	}
	if a.opts.IncludeDiff && len(a.runIDs) == 2 {
		if r.Diff == nil {
			r.Diff = shared.TestDiff{0, 0, 0}
		}
		r.Diff.Append(
			shared.TestStatus(a.runResults[a.runIDs[0]].GetResult(t)),
			shared.TestStatus(a.runResults[a.runIDs[1]].GetResult(t)),
			&a.opts.DiffFilter)
	}
	r.LegacyStatus = results
	a.agg[id] = r

	return nil
}

func (a *indexAggregator) Done() []query.SearchResult {
	res := make([]query.SearchResult, 0, len(a.agg))
	for _, r := range a.agg {
		res = append(res, r)
	}
	return res
}

func newIndexAggregator(idx index, runIDs []RunID, opts query.AggregationOpts) aggregator {
	return &indexAggregator{
		index:  idx,
		runIDs: runIDs,
		agg:    make(map[uint64]query.SearchResult),
		opts:   opts,
	}
}
