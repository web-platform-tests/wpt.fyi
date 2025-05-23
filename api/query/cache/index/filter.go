// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

// This file implements filtering of tests (or subtests) and their results for
// a set of test runs using a search tree (see api/query/atoms.go).
//
// Each search atom in the tree must define a Filter method, which is called
// for each TestID (test/subtest) to determine whether or not the TestID meets
// the search criteria. Atoms are responsible for recursing into their children.
//
// Before being filtered, search atoms are bound to an in-memory index giving
// them access to the full set of tests being filtered and a results map for
// each run.

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set"
	"github.com/sirupsen/logrus"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// True is a query.True equivalent, bound to an in-memory index.
type True struct {
	index
}

// False is a query.False equivalent, bound to an in-memory index.
type False struct {
	index
}

// TestNamePattern is a query.TestNamePattern bound to an in-memory index.
type TestNamePattern struct {
	index
	q query.TestNamePattern
}

// SubtestNamePattern is a query.SubtestNamePattern bound to an in-memory index.
type SubtestNamePattern struct {
	index
	q query.SubtestNamePattern
}

// TestPath is a query.TestPath bound to an in-memory index.
type TestPath struct {
	index
	q query.TestPath
}

// runTestStatusEq is a query.RunTestStatusEq bound to an
// in-memory index.
type runTestStatusEq struct {
	index
	q query.RunTestStatusEq
}

// runTestStatusNeq is a query.RunTestStatusNeq bound to an
// in-memory index.
type runTestStatusNeq struct {
	index
	q query.RunTestStatusNeq
}

// Count is a query.Count bound to an in-memory index.
type Count struct {
	index
	count int
	args  []filter
}

// LessThan is a query.LessThan bound to an in-memory index.
type LessThan Count

// MoreThan is a query.MoreThan bound to an in-memory index.
type MoreThan Count

// Link is a query.Link bound to an in-memory index and MetadataResults.
type Link struct {
	index
	pattern  string
	metadata map[string][]string
}

// Triaged is a query.Triaged bound to an in-memory index and MetadataResults of a single browser.
type Triaged struct {
	index
	metadata map[string][]string
}

// TestLabel is a query.TestLabel bound to an in-memory index and MetadataResults.
type TestLabel struct {
	index
	label    string
	metadata map[string][]string
}

// TestWebFeature is a query.TestWebFeature bound to an in-memory index and WebFeaturesData.
type TestWebFeature struct {
	index
	webFeature      string
	webFeaturesData shared.WebFeaturesData
}

// MetadataQuality is a query.MetadataQuality bound to an in-memory index.
type MetadataQuality struct {
	index
	quality query.MetadataQuality
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

func (i index) idx() index { return i }

// Filter always returns true for true.
func (True) Filter(_ TestID) bool {
	return true
}

// Filter always returns false for false.
func (False) Filter(_ TestID) bool {
	return false
}

// Filter interprets a TestNamePattern as a filter function over TestIDs.
func (tnp TestNamePattern) Filter(t TestID) bool {
	name, _, err := tnp.tests.GetName(t)
	if err != nil {
		return false
	}

	return strings.Contains(name, tnp.q.Pattern)
}

// Filter interprets a SubtestNamePattern as a filter function over TestIDs.
func (tnp SubtestNamePattern) Filter(t TestID) bool {
	_, subtest, err := tnp.tests.GetName(t)
	if err != nil || subtest == nil {
		return false
	}

	return strings.Contains(
		strings.ToLower(*subtest),
		strings.ToLower(tnp.q.Subtest),
	)
}

// Filter interprets a TestPath as a filter function over TestIDs.
func (tp TestPath) Filter(t TestID) bool {
	name, _, err := tp.tests.GetName(t)
	if err != nil {
		return false
	}

	return strings.HasPrefix(name, tp.q.Path)
}

// Filter interprets a runTestStatusEq as a filter function over TestIDs.
func (rtse runTestStatusEq) Filter(t TestID) bool {
	return rtse.runResults[RunID(rtse.q.Run)].GetResult(t) == ResultID(rtse.q.Status)
}

// Filter interprets a runTestStatusNeq as a filter function over TestIDs.
func (rtsn runTestStatusNeq) Filter(t TestID) bool {
	return rtsn.runResults[RunID(rtsn.q.Run)].GetResult(t) != ResultID(rtsn.q.Status)
}

// Filter interprets a Count as a filter function over TestIDs.
func (c Count) Filter(t TestID) bool {
	args := c.args
	matches := 0
	for _, arg := range args {
		if arg.Filter(t) {
			matches++
		}
	}

	return matches == c.count
}

// Filter interprets a LessThan as a filter function over TestIDs.
func (c LessThan) Filter(t TestID) bool {
	args := c.args
	matches := 0
	for _, arg := range args {
		if arg.Filter(t) {
			matches++
		}
	}

	return matches < c.count
}

// Filter interprets a MoreThan as a filter function over TestIDs.
func (c MoreThan) Filter(t TestID) bool {
	args := c.args
	matches := 0
	for _, arg := range args {
		if arg.Filter(t) {
			matches++
		}
	}

	return matches > c.count
}

// Filter interprets a Link as a filter function over TestIDs.
func (l Link) Filter(t TestID) bool {
	name, _, err := l.tests.GetName(t)
	if err != nil {
		return false
	}

	// WPT metadata can contain wildcards that match arbitrary
	// subdirectories, so if we fail to lookup the map we keep stripping
	// directories and try again.
	// nolint:godox // TODO: Verify whether this is too slow; if so, try building a trie
	// from the wildcards only and match to that as a fallback.
	urls, ok := l.metadata[name]
	dir := filepath.Dir(name)
	// Dir terminates with either '.' (when the top-level is a file) or '/'
	// (when the top-level is a directory).
	for !ok && len(dir) > 1 {
		urls, ok = l.metadata[dir+"/*"]
		if ok {
			break
		}

		dir = filepath.Dir(dir)
	}
	if !ok {
		return false
	}

	for _, url := range urls {
		if strings.Contains(url, l.pattern) {
			return true
		}
	}

	return false
}

// Filter interprets a Triaged as a filter function over TestIDs.
func (tr Triaged) Filter(t TestID) bool {
	name, _, err := tr.tests.GetName(t)
	if err != nil {
		return false
	}

	// WPT metadata can contain wildcards that match arbitrary
	// subdirectories, so if we fail to lookup the map we keep stripping
	// directories and try again.
	// nolint:godox // TODO: Verify whether this is too slow; if so, try building a trie
	// from the wildcards only and match to that as a fallback.
	val, ok := tr.metadata[name]
	dir := filepath.Dir(name)
	// Dir terminates with either '.' (when the top-level is a file) or '/'
	// (when the top-level is a directory).
	for !ok && len(dir) > 1 {
		val, ok = tr.metadata[dir+"/*"]
		if ok {
			break
		}

		dir = filepath.Dir(dir)
	}

	if !ok {
		return false
	}

	if len(val) == 0 {
		return false
	}

	for _, url := range val {
		if url != "" {
			return true
		}
	}

	return false
}

// Filter interprets a TestLabel as a filter function over TestIDs.
func (tl TestLabel) Filter(t TestID) bool {
	name, _, err := tl.tests.GetName(t)
	if err != nil {
		return false
	}

	labels := tl.metadata[name]
	dir := filepath.Dir(name)
	// Dir terminates with either '.' (when the top-level is a file) or '/'
	// (when the top-level is a directory).
	for len(dir) > 1 {
		lbs := tl.metadata[dir+"/*"]
		labels = append(labels, lbs...)
		dir = filepath.Dir(dir)
	}

	for _, label := range labels {
		if strings.EqualFold(label, tl.label) {
			return true
		}
	}

	return false
}

// Filter interprets a TestWebFeature as a filter function over TestIDs.
func (twf TestWebFeature) Filter(t TestID) bool {
	name, _, err := twf.tests.GetName(t)
	if err != nil {
		return false
	}
	// Check if there's any data.
	if twf.webFeaturesData == nil {
		return false
	}
	// Get the Web Features for that exact test path.
	return twf.webFeaturesData.TestMatchesWithWebFeature(name, twf.webFeature)
}

// Filter interprets a MetadataQuality as a filter function over TestIDs.
func (q MetadataQuality) Filter(t TestID) bool {
	switch q.quality {
	case query.MetadataQualityDifferent:
		// is:different only returns subtest rows where the result
		// differs between the runs we are comparing. To detect this,
		// put them into a set and then check the size.
		set := mapset.NewSet()
		for _, result := range q.runResults {
			set.Add(result.GetResult(t))
		}

		return set.Cardinality() > 1
	case query.MetadataQualityTentative:
		// is:tentative only returns rows from tests with .tentative.
		// in their name. See
		// https://web-platform-tests.org/writing-tests/file-names.html
		name, _, err := q.tests.GetName(t)
		if err != nil {
			return false
		}

		return strings.Contains(name, ".tentative.") || strings.Contains(name, "/tentative/")
	case query.MetadataQualityOptional:
		// is:optional only returns rows from tests with .optional.
		// in their name. See
		// https://web-platform-tests.org/writing-tests/file-names.html
		// nolint:godox // TODO(gh-1619): Handle the CSS meta flags; see
		// https://web-platform-tests.org/writing-tests/css-metadata.html#requirement-flags
		name, _, err := q.tests.GetName(t)
		if err != nil {
			return false
		}

		return strings.Contains(name, ".optional.")
	case query.MetadataQualityUnknown:
		return false
	default:
		return false
	}
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

// nolint:ireturn // TODO: Fix ireturn lint error
func newFilter(idx index, q query.ConcreteQuery) (filter, error) {
	if q == nil {
		return nil, errors.New("nil ConcreteQuery provided")
	}
	switch v := q.(type) {
	case query.True:
		return True{idx}, nil
	case query.False:
		return False{idx}, nil
	case query.TestNamePattern:
		return TestNamePattern{idx, v}, nil
	case query.SubtestNamePattern:
		return SubtestNamePattern{idx, v}, nil
	case query.TestPath:
		return TestPath{idx, v}, nil
	case query.RunTestStatusEq:
		return runTestStatusEq{idx, v}, nil
	case query.RunTestStatusNeq:
		return runTestStatusNeq{idx, v}, nil
	case query.Count:
		fs, err := filters(idx, v.Args)
		if err != nil {
			return nil, err
		}

		return Count{idx, v.Count, fs}, nil
	case query.LessThan:
		fs, err := filters(idx, v.Args)
		if err != nil {
			return nil, err
		}

		return LessThan{idx, v.Count.Count, fs}, nil
	case query.MoreThan:
		fs, err := filters(idx, v.Args)
		if err != nil {
			return nil, err
		}

		return MoreThan{idx, v.Count.Count, fs}, nil
	case query.Link:
		return Link{idx, v.Pattern, v.Metadata}, nil
	case query.Triaged:
		return Triaged{idx, v.Metadata}, nil
	case query.TestLabel:
		return TestLabel{idx, v.Label, v.Metadata}, nil
	case query.TestWebFeature:
		return TestWebFeature{idx, v.WebFeature, v.WebFeaturesData}, nil
	case query.MetadataQuality:
		return MetadataQuality{idx, v}, nil
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
		return nil, fmt.Errorf("unknown ConcreteQuery type %s", reflect.TypeOf(q))
	}
}

// Execute runs each filter in a ShardedFilter in parallel, returning a slice of
// TestIDs as the result. Note that TestIDs are not deduplicated; the assumption
// is that each filter is bound to a different shard, sharded by TestID.
func (fs ShardedFilter) Execute(runs []shared.TestRun, opts query.AggregationOpts) interface{} {
	rus := make([]RunID, len(runs))
	for i := range runs {
		rus[i] = RunID(runs[i].ID)
	}
	res := make(chan []shared.SearchResult, len(fs))
	errs := make(chan error)
	for _, f := range fs {
		go syncRunFilter(rus, f, opts, res, errs)
	}

	ret := make([]shared.SearchResult, 0)
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
				// nolint:godox // TODO: Should this use a context-based logger?
				logrus.Errorf("Error executing filter query: %v: %v", fs, err)
			}
		}()
	}

	return ret
}

func syncRunFilter(rus []RunID, f filter, opts query.AggregationOpts, res chan []shared.SearchResult, errs chan error) {
	idx := f.idx()
	idx.m.RLock()
	defer idx.m.RUnlock()

	agg := newIndexAggregator(idx, rus, opts)
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
