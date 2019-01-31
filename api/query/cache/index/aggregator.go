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

	rus             []RunID
	agg             map[uint64]query.SearchResult
	includeSubtests bool
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

		r = query.SearchResult{Test: name, LegacyStatus: nil}
	}

	rus := r.LegacyStatus
	if rus == nil {
		rus = make([]query.LegacySearchRunResult, len(a.rus))
	}

	for i, ru := range a.rus {
		res := int64(a.runResults[ru].GetResult(t))
		// TODO: Switch to a consistent value for Total across all runs.
		//
		// Only include tests with non-UNKNOWN status for this run's total.
		if res != shared.TestStatusUnknown {
			rus[i].Total++
			if res == shared.TestStatusPass || res == shared.TestStatusOK {
				rus[i].Passes++
			}
		}

		if a.includeSubtests {
			if _, name, err := ts.GetName(t); err == nil && name != nil {
				r.Subtests.Add(*name)
			}
		}
	}
	r.LegacyStatus = rus
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

func newIndexAggregator(idx index, rus []RunID, opts query.AggregationOpts) aggregator {
	return &indexAggregator{
		index:           idx,
		rus:             rus,
		agg:             make(map[uint64]query.SearchResult),
		includeSubtests: opts.IncludeSubtests,
	}
}
