// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"errors"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// TestNamePattern is a query.TestNamePattern bound to an in-memory index.
type TestNamePattern struct {
	index
	q query.TestNamePattern
}

// RunTestStatusConstraint is a query.RunTestStatusConstraint bound to an
// in-memory index.
type RunTestStatusConstraint struct {
	index
	q query.RunTestStatusConstraint
}

// And is a query.And bound to an in-memory index.
type And struct {
	index
	args []filter
}

// Or is a query.Or bound to an in-memory index.
type Or struct {
	index
	args []filter
}

// Not is a query.Not bound to an in-memory index.
type Not struct {
	index
	arg filter
}

// ShardedFilter is a collection of filters, each bound to a shard of in-memory
// index data.
type ShardedFilter []filter

type filter interface {
	Filter(TestID) bool
	idx() index
}

type index struct {
	tests      Tests
	runResults map[RunID]RunResults
	m          *sync.RWMutex
}

var errUnknownConcreteQuery = errors.New("Unknown ConcreteQuery type")

func (i index) idx() index { return i }

// Filter interprets a TestNamePattern as a filter function over TestIDs.
func (tnp TestNamePattern) Filter(t TestID) bool {
	name, _, err := tnp.tests.GetName(t)
	if err != nil {
		return false
	}
	return strings.Contains(name, tnp.q.Pattern)
}

// Filter interprets a RunTestStatusConstraint as a filter function over
// TestIDs.
func (rtsc RunTestStatusConstraint) Filter(t TestID) bool {
	return rtsc.runResults[RunID(rtsc.q.Run)].GetResult(t) == ResultID(rtsc.q.Status)
}

// Filter interprets an And as a filter function over TestIDs.
func (a And) Filter(t TestID) bool {
	args := a.args
	for _, arg := range args {
		if !arg.Filter(t) {
			return false
		}
	}
	return true
}

// Filter interprets an Or as a filter function over TestIDs.
func (o Or) Filter(t TestID) bool {
	args := o.args
	for _, arg := range args {
		if arg.Filter(t) {
			return true
		}
	}
	return false
}

// Filter interprets a Not as a filter function over TestID.
func (n Not) Filter(t TestID) bool {
	return !n.arg.Filter(t)
}

func newFilter(idx index, q query.ConcreteQuery) (filter, error) {
	switch v := q.(type) {
	case query.TestNamePattern:
		return TestNamePattern{idx, v}, nil
	case query.RunTestStatusConstraint:
		return RunTestStatusConstraint{idx, v}, nil
	case query.And:
		fs, err := filters(idx, v.Args)
		if err != nil {
			return nil, err
		}
		return And{idx, fs}, nil
	case query.Or:
		fs, err := filters(idx, v.Args)
		if err != nil {
			return nil, err
		}
		return Or{idx, fs}, nil
	case query.Not:
		f, err := newFilter(idx, v.Arg)
		if err != nil {
			return nil, err
		}
		return Not{idx, f}, nil
	default:
		return nil, errUnknownConcreteQuery
	}
}

// Execute runs each filter in a ShardedFilter in parallel, returning a slice of
// TestIDs as the result. Note that TestIDs are not deduplicated; the assumption
// is that each filter is bound to a different shard, sharded by TestID.
func (fs ShardedFilter) Execute(runs []shared.TestRun) interface{} {
	return fs.syncExecute(runs)
}

func (fs ShardedFilter) syncExecute(runs []shared.TestRun) interface{} {
	rus := make([]RunID, len(runs))
	for i := range runs {
		rus[i] = RunID(runs[i].ID)
	}
	res := make(chan []query.SearchResult, len(fs))
	errs := make(chan error)
	for _, f := range fs {
		go syncRunFilter(rus, f, res, errs)
	}

	ret := make([]query.SearchResult, 0)
	for i := 0; i < len(fs); i++ {
		ts := <-res
		ret = append(ret, ts...)
	}

	// To keep query execution fast, report errors in a separate goroutine and
	// return results immediately. The class of errors for query execution (as
	// apposed to binding) should be extremely rare and can be acted upon by
	// monitoring logs.
	close(errs)
	if len(errs) > 0 {
		go func() {
			for err := range errs {
				// TODO: Should this use a context-based logger?
				log.Errorf("Error executing filter query: %v: %v", fs, err)
			}
		}()
	}

	return ret
}

func syncRunFilter(rus []RunID, f filter, res chan []query.SearchResult, errs chan error) {
	idx := f.idx()
	idx.m.RLock()
	defer idx.m.RUnlock()

	agg := newIndexAggregator(idx, rus)
	idx.tests.Range(func(t TestID) bool {
		if f.Filter(t) {
			err := agg.Add(t)
			if err != nil {
				errs <- err
			}
		}
		return true
	})
	res <- agg.Done()
}

func filters(idx index, qs []query.ConcreteQuery) ([]filter, error) {
	fs := make([]filter, len(qs))
	var err error
	for i := range qs {
		fs[i], err = newFilter(idx, qs[i])
		if err != nil {
			return nil, err
		}
	}
	return fs, nil
}
