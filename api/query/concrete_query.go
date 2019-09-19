// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// Average as of Aug 20, 2019
const averageNumberOfSubtests = int(1655263 / 34236)

// AggregationOpts are options for the aggregation format used when collecting
// the results.
type AggregationOpts struct {
	IncludeSubtests         bool
	InteropFormat           bool
	IncludeDiff             bool
	IgnoreTestHarnessResult bool // Don't +1 the "OK" status for testharness tests.
	DiffFilter              shared.DiffFilterParam
}

// Binder is a mechanism for binding a query over a slice of test runs to
// a particular query service mechanism.
type Binder interface {
	// Bind produces an query execution Plan and/or error after binding its inputs
	// to a query service mechanism. E.g., an in-memory cache may verify that the
	// given runs are in the cache and extract results data subsets that pertain
	// to the runs before producing a Plan implementation that can operate over
	// the subsets directly.
	Bind([]shared.TestRun, ConcreteQuery) (Plan, error)
}

// Plan a query execution plan that returns results.
type Plan interface {
	// Execute runs the query execution plan. The result set type depends on the
	// underlying query service mechanism that the Plan was bound with.
	Execute([]shared.TestRun, AggregationOpts) interface{}
}

// ConcreteQuery is an AbstractQuery that has been bound to specific test runs.
type ConcreteQuery interface {
	Size() int
}

// Count constrains search results to include only test results where the number
// of runs that match the given criteria is exactly the expected count.
type Count struct {
	Count int
	Args  []ConcreteQuery
}

// MoreThan constrains search results to include only test results where the number
// of runs that match the given criteria is more than the given count.
type MoreThan struct {
	Count
}

// LessThan constrains search results to include only test results where the number
// of runs that match the given criteria is less than the given count.
type LessThan struct {
	Count
}

// Link is a ConcreteQuery of AbstractLink.
type Link struct {
	Pattern  string
	Metadata map[string][]string
}

// RunTestStatusEq constrains search results to include only test results from a
// particular run that have a particular test status value. Run IDs are those
// values automatically assigned to shared.TestRun instances by Datastore.
// Status IDs are those codified in shared.TestStatus* symbols.
type RunTestStatusEq struct {
	Run    int64
	Status shared.TestStatus
}

// RunTestStatusNeq constrains search results to include only test results from a
// particular run that do not have a particular test status value. Run IDs are
// those values automatically assigned to shared.TestRun instances by Datastore.
// Status IDs are those codified in shared.TestStatus* symbols.
type RunTestStatusNeq struct {
	Run    int64
	Status shared.TestStatus
}

// Or is a logical disjunction of ConcreteQuery instances.
type Or struct {
	Args []ConcreteQuery
}

// And is a logical conjunction of ConcreteQuery instances.
type And struct {
	Args []ConcreteQuery
}

// Not is a logical negation of ConcreteQuery instances.
type Not struct {
	Arg ConcreteQuery
}

// Size of TestNamePattern has a size of 1: servicing such a query requires a
// substring match per test.
func (TestNamePattern) Size() int { return 1 }

// Size of SubtestNamePattern has a size of 1: servicing such a query requires a
// substring match per subtest.
func (SubtestNamePattern) Size() int { return averageNumberOfSubtests }

// Size of TestPath has a size of 1: servicing such a query requires a
// substring match per test.
func (TestPath) Size() int { return 1 }

// Size of RunTestStatusEq is 1: servicing such a query requires a single lookup
// in a test run result mapping per test.
func (RunTestStatusEq) Size() int { return 1 }

// Size of RunTestStatusNeq is 1: servicing such a query requires a single
// lookup in a test run result mapping per test.
func (RunTestStatusNeq) Size() int { return 1 }

// Size of Link has a size of 1: servicing such a query requires a
// substring match per Metadata Link Node.
func (Link) Size() int { return 1 }

// Size of Count is the sum of the sizes of its constituent ConcretQuery instances.
func (c Count) Size() int { return size(c.Args) }

// Size of Is depends on the quality.
func (q MetadataQuality) Size() int {
	// Currently only 'Different' supported, which is one set comparison per row.
	return 1
}

// Size of Or is the sum of the sizes of its constituent ConcretQuery instances.
func (o Or) Size() int { return size(o.Args) }

// Size of And is the sum of the sizes of its constituent ConcretQuery
// instances.
func (a And) Size() int { return size(a.Args) }

// Size of Not is one unit greater than the size of its ConcreteQuery argument.
func (n Not) Size() int { return 1 + n.Arg.Size() }

// Size of True is 0: It should be optimized out of queries in practice.
func (True) Size() int { return 0 }

// Size of False is 0: It should be optimized out of queries in practice.
func (False) Size() int { return 0 }

func size(qs []ConcreteQuery) int {
	s := 0
	for _, q := range qs {
		s += q.Size()
	}
	return s
}
