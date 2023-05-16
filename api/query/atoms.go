// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

// This file defines search atoms on the backend. All wpt.fyi search queries
// are broken down into a tree of search atoms, which is then traversed by the
// searchcache to find matching tests to include.
//
// All search atoms must define two methods:
//   i. BindToRuns
//   ii. UnmarshalJSON
//
// These are best understood in reverse, as that is also the order they are
// called in. UnmarshalJSON is used to convert from the original JSON search
// query into the tree of abstract search atoms. The atoms are referred to as
// abstract as they do not yet relate to any underlying data (i.e. any test
// runs). Many types of atom (such as AbstractExists) perform this
// unmarshalling recursively, which is how we end up with a tree.
//
// Once we have an abstract search tree, BindToRuns will convert it to a
// concrete search tree (that is, a tree of ConcreteQuery atoms). This gives
// the search atoms access to the specific runs that are being searched over,
// to pull any specific information needed. For example, this allows
// TestStatusEq to only produce results for test runs that match the specified
// product (and short-circuit entirely if no test runs match).
//
// Some abstract search atoms may produce more than one concrete search atom
// (e.g. AbstractExists, which produces a disjunction), whilst others may
// ignore the test runs entirely if they aren't relevant (e.g.
// TestNamePattern, which only cares about the test name and not the results).
//
// Note that this file does not perform the actual filtering of tests from the
// test runs to produce the search response; for that see the `filter` type in
// api/query/cache/index/filter.go

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

var browsers = shared.GetDefaultBrowserNames()

// AbstractQuery is an intermetidate representation of a test results query that
//  has not been bound to specific shared.TestRun specs for processing.
type AbstractQuery interface {
	BindToRuns(runs ...shared.TestRun) ConcreteQuery
}

// RunQuery is the internal representation of a query recieved from an HTTP
// client, including the IDs of the test runs to query, and the structured query
// to run.
type RunQuery struct {
	RunIDs []int64
	AbstractQuery
}

// True is a true-valued ConcreteQuery.
type True struct{}

// BindToRuns for True is a no-op; it is independent of test runs.
func (t True) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	return t
}

// False is a false-valued ConcreteQuery.
type False struct{}

// BindToRuns for False is a no-op; it is independent of test runs.
func (f False) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	return f
}

// TestNamePattern is a query atom that matches test names to a pattern string.
type TestNamePattern struct {
	Pattern string
}

// BindToRuns for TestNamePattern is a no-op; it is independent of test runs.
func (tnp TestNamePattern) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	return tnp
}

// SubtestNamePattern is a query atom that matches subtest names to a pattern string.
type SubtestNamePattern struct {
	Subtest string
}

// BindToRuns for SubtestNamePattern is a no-op; it is independent of test runs.
func (tnp SubtestNamePattern) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	return tnp
}

// TestPath is a query atom that matches exact test path prefixes.
// It is an inflexible equivalent of TestNamePattern.
type TestPath struct {
	Path string
}

// BindToRuns for TestNamePattern is a no-op; it is independent of test runs.
func (tp TestPath) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	return tp
}

// AbstractExists represents an array of abstract queries, each of which must be
// satifisfied by some run. It represents the root of a structured query.
type AbstractExists struct {
	Args []AbstractQuery
}

// BindToRuns binds each abstract query to an or-combo of that query against
// each specific/individual run.
func (e AbstractExists) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	queries := make([]ConcreteQuery, len(e.Args))
	// When the nested query is a single query, e.g. And/Or, bind that query directly.
	if len(e.Args) == 1 {
		return e.Args[0].BindToRuns(runs...)
	}

	for i, arg := range e.Args {
		var query ConcreteQuery
		// Exists queries are split; one run must satisfy the whole tree.
		byRun := make([]ConcreteQuery, 0, len(runs))
		for _, run := range runs {
			bound := arg.BindToRuns(run)
			if _, ok := bound.(False); !ok {
				byRun = append(byRun, bound)
			}
		}
		query = Or{Args: byRun}
		queries[i] = query
	}
	// And the overall node is true if all its exists queries are true.
	return And{
		Args: queries,
	}
}

// AbstractAll represents an array of abstract queries, each of which must be
// satifisfied by all runs. It represents the root of a structured query.
type AbstractAll struct {
	Args []AbstractQuery
}

// BindToRuns binds each abstract query to an and-combo of that query against
// each specific/individual run.
func (e AbstractAll) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	queries := make([]ConcreteQuery, len(e.Args))
	for i, arg := range e.Args {
		var query ConcreteQuery
		byRun := make([]ConcreteQuery, 0, len(runs))
		for _, run := range runs {
			bound := arg.BindToRuns(run)
			if _, ok := bound.(True); !ok { // And with True is pointless.
				byRun = append(byRun, bound)
			}
		}
		query = And{Args: byRun}
		queries[i] = query
	}
	// And the overall node is true if all its exists queries are true.
	return And{
		Args: queries,
	}
}

// AbstractNone represents an array of abstract queries, each of which must not be
// satifisfied by any run. It represents the root of a structured query.
type AbstractNone struct {
	Args []AbstractQuery
}

// BindToRuns binds to a not-exists for the same query(s).
func (e AbstractNone) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	return Not{
		AbstractExists{
			Args: e.Args,
		}.BindToRuns(runs...),
	}
}

// AbstractSequential represents the root of a sequential queries, where the first
// query must be satisfied by some run such that the next run, sequentially, also
// satisfies the next query, and so on.
type AbstractSequential struct {
	Args []AbstractQuery
}

// BindToRuns binds each sequential query to an and-combo of those queries against
// specific sequential runs, for each combination of sequential runs.
func (e AbstractSequential) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	numSeqQueries := len(e.Args)
	byRuns := []ConcreteQuery{}
	for i := 0; i+numSeqQueries-1 < len(runs); i++ {
		all := And{}
		for j, arg := range e.Args {
			all.Args = append(all.Args, arg.BindToRuns(runs[i+j]))
		}
		byRuns = append(byRuns, all)
	}
	return Or{
		Args: byRuns,
	}
}

// AbstractCount represents the root of a count query, where the exact number of
// runs that satisfy the query must match the expected count.
type AbstractCount struct {
	Count int
	Where AbstractQuery
}

// BindToRuns binds each count query to all of the runs, so that it can count the
// number of runs that match the criteria.
func (c AbstractCount) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	byRun := []ConcreteQuery{}
	for _, run := range runs {
		byRun = append(byRun, c.Where.BindToRuns(run))
	}
	return Count{
		Count: c.Count,
		Args:  byRun,
	}
}

// AbstractMoreThan is the root of a moreThan query, where the number of runs
// that satisfy the query must be more than the given count.
type AbstractMoreThan struct {
	AbstractCount
}

// BindToRuns binds each count query to all of the runs, so that it can count the
// number of runs that match the criteria.
func (m AbstractMoreThan) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	c := m.AbstractCount.BindToRuns(runs...).(Count)
	return MoreThan{c}
}

// AbstractLessThan is the root of a lessThan query, where the number of runs
// that satisfy the query must be less than the given count.
type AbstractLessThan struct {
	AbstractCount
}

// BindToRuns binds each count query to all of the runs, so that it can count the
// number of runs that match the criteria.
func (l AbstractLessThan) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	c := l.AbstractCount.BindToRuns(runs...).(Count)
	return LessThan{c}
}

// AbstractLink is represents the root of a link query, which matches Metadata
// URLs to a pattern string.
type AbstractLink struct {
	Pattern         string
	metadataFetcher shared.MetadataFetcher
}

// BindToRuns for AbstractLink fetches metadata for either test-level issues or
// issues associated with the given runs. It does not filter the metadata by
// the pattern yet.
func (l AbstractLink) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	if l.metadataFetcher == nil {
		l.metadataFetcher = searchcacheMetadataFetcher{}
	}
	includeTestLevel := true
	metadata, _ := shared.GetMetadataResponse(runs, includeTestLevel, logrus.StandardLogger(), l.metadataFetcher)
	metadataMap := shared.PrepareLinkFilter(metadata)

	return Link{
		Pattern:  l.Pattern,
		Metadata: metadataMap,
	}
}

// AbstractTriaged represents the root of a triaged query that matches
// tests where the test of a specific browser has been triaged through Metadata
type AbstractTriaged struct {
	Product         *shared.ProductSpec
	metadataFetcher shared.MetadataFetcher
}

// BindToRuns for AbstractTriaged binds each run matching the AbstractTriaged
// ProductSpec to a triaged object.
func (t AbstractTriaged) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	cq := make([]ConcreteQuery, 0)

	if t.metadataFetcher == nil {
		t.metadataFetcher = searchcacheMetadataFetcher{}
	}
	for _, run := range runs {
		if t.Product == nil || t.Product.Matches(run) {
			// We only want to fetch metadata for this specific run (or for no runs, if
			// the search is for test-level issues).
			includeTestLevel := false
			metadataRuns := []shared.TestRun{run}

			// Product being nil means that we want test-level issues.
			if t.Product == nil {
				includeTestLevel = true
				metadataRuns = []shared.TestRun{}
			}
			metadata, _ := shared.GetMetadataResponse(metadataRuns, includeTestLevel, logrus.StandardLogger(), t.metadataFetcher)
			metadataMap := shared.PrepareLinkFilter(metadata)
			if len(metadataMap) > 0 {
				cq = append(cq, Triaged{run.ID, metadataMap})
			}
		}
	}

	if len(cq) == 0 {
		return False{}
	}

	return Or{cq}
}

// AbstractTestLabel represents the root of a testlabel query, which matches test-level metadata
// labels to a searched label.
type AbstractTestLabel struct {
	Label           string
	metadataFetcher shared.MetadataFetcher
}

// BindToRuns for AbstractTestLabel fetches test-level metadata; it is independent of test runs.
func (t AbstractTestLabel) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	if t.metadataFetcher == nil {
		t.metadataFetcher = searchcacheMetadataFetcher{}
	}

	includeTestLevel := true
	// Passing []shared.TestRun{} means that we want test-level issues.
	metadata, _ := shared.GetMetadataResponse([]shared.TestRun{}, includeTestLevel, logrus.StandardLogger(), t.metadataFetcher)
	metadataMap := shared.PrepareTestLabelFilter(metadata)

	return TestLabel{
		Label:    t.Label,
		Metadata: metadataMap,
	}
}

// MetadataQuality represents the root of an "is" query, which asserts known
// metadata qualities to the results
type MetadataQuality int

const (
	// MetadataQualityUnknown is a placeholder for unrecognized values.
	MetadataQualityUnknown MetadataQuality = iota
	// MetadataQualityDifferent represents an is:different atom.
	// "different" ensures that one or more results differs from the other results.
	MetadataQualityDifferent
	// MetadataQualityTentative represents an is:tentative atom.
	// "tentative" ensures that the results are from a tentative test.
	MetadataQualityTentative
	// MetadataQualityOptional represents an is:optional atom.
	// "optional" ensures that the results are from an optional test.
	MetadataQualityOptional
)

// BindToRuns for MetadataQuality is a no-op; it is independent of test runs.
func (q MetadataQuality) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	return q
}

// TestStatusEq is a query atom that matches tests where the test status/result
// from at least one test run matches the given status value, optionally filtered
// to a specific browser name.
type TestStatusEq struct {
	Product *shared.ProductSpec
	Status  shared.TestStatus
}

// TestStatusNeq is a query atom that matches tests where the test status/result
// from at least one test run does not match the given status value, optionally
// filtered to a specific browser name.
type TestStatusNeq struct {
	Product *shared.ProductSpec
	Status  shared.TestStatus
}

// BindToRuns for TestStatusEq expands to a disjunction of RunTestStatusEq
// values.
func (tse TestStatusEq) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	ids := make([]int64, 0, len(runs))
	for _, run := range runs {
		if tse.Product == nil || tse.Product.Matches(run) {
			ids = append(ids, run.ID)
		}
	}
	if len(ids) == 0 {
		return False{}
	}
	if len(ids) == 1 {
		return RunTestStatusEq{ids[0], tse.Status}
	}

	q := Or{make([]ConcreteQuery, len(ids))}
	for i := range ids {
		q.Args[i] = RunTestStatusEq{ids[i], tse.Status}
	}
	return q
}

// BindToRuns for TestStatusNeq expands to a disjunction of RunTestStatusNeq
// values.
func (tsn TestStatusNeq) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	ids := make([]int64, 0, len(runs))
	for _, run := range runs {
		if tsn.Product == nil || tsn.Product.Matches(run) {
			ids = append(ids, run.ID)
		}
	}
	if len(ids) == 0 {
		return False{}
	}
	if len(ids) == 1 {
		return RunTestStatusNeq{ids[0], tsn.Status}
	}

	q := Or{make([]ConcreteQuery, len(ids))}
	for i := range ids {
		q.Args[i] = RunTestStatusNeq{ids[i], tsn.Status}
	}
	return q
}

// AbstractNot is the AbstractQuery for negation.
type AbstractNot struct {
	Arg AbstractQuery
}

// BindToRuns for AbstractNot produces a Not with a bound argument.
func (n AbstractNot) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	return Not{n.Arg.BindToRuns(runs...)}
}

// AbstractOr is the AbstractQuery for disjunction.
type AbstractOr struct {
	Args []AbstractQuery
}

// BindToRuns for AbstractOr produces an Or with bound arguments.
func (o AbstractOr) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	args := make([]ConcreteQuery, 0, len(o.Args))
	for i := range o.Args {
		sub := o.Args[i].BindToRuns(runs...)
		if t, ok := sub.(True); ok {
			return t
		}
		if _, ok := sub.(False); ok {
			continue
		}
		args = append(args, sub)
	}
	if len(args) == 0 {
		return False{}
	}
	if len(args) == 1 {
		return args[0]
	}
	return Or{
		Args: args,
	}
}

// AbstractAnd is the AbstractQuery for conjunction.
type AbstractAnd struct {
	Args []AbstractQuery
}

// BindToRuns for AbstractAnd produces an And with bound arguments.
func (a AbstractAnd) BindToRuns(runs ...shared.TestRun) ConcreteQuery {
	args := make([]ConcreteQuery, 0, len(a.Args))
	for i := range a.Args {
		sub := a.Args[i].BindToRuns(runs...)
		if _, ok := sub.(False); ok {
			return False{}
		}
		if _, ok := sub.(True); ok {
			continue
		}
		args = append(args, sub)
	}
	if len(args) == 0 {
		return False{}
	}
	if len(args) == 1 {
		return args[0]
	}
	return And{
		Args: args,
	}
}

// UnmarshalJSON interprets the JSON representation of a RunQuery, instantiating
// (an) appropriate Query implementation(s) according to the JSON structure.
func (rq *RunQuery) UnmarshalJSON(b []byte) error {
	var data struct {
		RunIDs []int64         `json:"run_ids"`
		Query  json.RawMessage `json:"query"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if len(data.RunIDs) == 0 {
		return errors.New(`Missing run query property: "run_ids"`)
	}
	rq.RunIDs = data.RunIDs

	if len(data.Query) > 0 {
		q, err := unmarshalQ(data.Query)
		if err != nil {
			return err
		}
		rq.AbstractQuery = q
	} else {
		rq.AbstractQuery = True{}
	}
	return nil
}

// UnmarshalJSON for TestNamePattern attempts to interpret a query atom as
// {"pattern":<test name pattern string>}.
func (tnp *TestNamePattern) UnmarshalJSON(b []byte) error {
	var data map[string]*json.RawMessage
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	patternMsg, ok := data["pattern"]
	if !ok {
		return errors.New(`Missing test name pattern property: "pattern"`)
	}
	var pattern string
	if err := json.Unmarshal(*patternMsg, &pattern); err != nil {
		return errors.New(`test name pattern property "pattern" is not a string`)
	}

	tnp.Pattern = pattern
	return nil
}

// UnmarshalJSON for SubtestNamePattern attempts to interpret a query atom as
// {"subtest":<subtest name pattern string>}.
func (tnp *SubtestNamePattern) UnmarshalJSON(b []byte) error {
	var data map[string]*json.RawMessage
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	subtestMsg, ok := data["subtest"]
	if !ok {
		return errors.New(`Missing subtest name pattern property: "subtest"`)
	}
	var subtest string
	if err := json.Unmarshal(*subtestMsg, &subtest); err != nil {
		return errors.New(`Subtest name property "subtest" is not a string`)
	}

	tnp.Subtest = subtest
	return nil
}

// UnmarshalJSON for TestPath attempts to interpret a query atom as
// {"path":<test name pattern string>}.
func (tp *TestPath) UnmarshalJSON(b []byte) error {
	var data map[string]*json.RawMessage
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	pathMsg, ok := data["path"]
	if !ok {
		return errors.New(`Missing test name path property: "path"`)
	}
	var path string
	if err := json.Unmarshal(*pathMsg, &path); err != nil {
		return errors.New(`Missing test name path property "path" is not a string`)
	}

	tp.Path = path
	return nil
}

// UnmarshalJSON for TestStatusEq attempts to interpret a query atom as
// {"product": <browser name>, "status": <status string>}.
func (tse *TestStatusEq) UnmarshalJSON(b []byte) error {
	var data struct {
		BrowserName string `json:"browser_name"` // Legacy
		Product     string `json:"product"`
		Status      string `json:"status"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if data.Product == "" && data.BrowserName != "" {
		data.Product = data.BrowserName
	}
	if len(data.Status) == 0 {
		return errors.New(`Missing test status constraint property: "status"`)
	}

	var product *shared.ProductSpec
	if data.Product != "" {
		p, err := shared.ParseProductSpec(data.Product)
		if err != nil {
			return err
		}
		product = &p
	}

	statusStr := strings.ToUpper(data.Status)
	status := shared.TestStatusValueFromString(statusStr)
	statusStr2 := status.String()
	if statusStr != statusStr2 {
		return fmt.Errorf(`Invalid test status: "%s"`, data.Status)
	}

	tse.Product = product
	tse.Status = status
	return nil
}

// UnmarshalJSON for TestStatusNeq attempts to interpret a query atom as
// {"product": <browser name>, "status": {"not": <status string>}}.
func (tsn *TestStatusNeq) UnmarshalJSON(b []byte) error {
	var data struct {
		BrowserName string `json:"browser_name"` // Legacy
		Product     string `json:"product"`
		Status      struct {
			Not string `json:"not"`
		} `json:"status"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if data.Product == "" && data.BrowserName != "" {
		data.Product = data.BrowserName
	}
	if len(data.Status.Not) == 0 {
		return errors.New(`Missing test status constraint property: "status.not"`)
	}

	var product *shared.ProductSpec
	if data.Product != "" {
		p, err := shared.ParseProductSpec(data.Product)
		if err != nil {
			return err
		}
		product = &p
	}

	statusStr := strings.ToUpper(data.Status.Not)
	status := shared.TestStatusValueFromString(statusStr)
	statusStr2 := status.String()
	if statusStr != statusStr2 {
		return fmt.Errorf(`Invalid test status: "%s"`, data.Status)
	}

	tsn.Product = product
	tsn.Status = status
	return nil
}

// UnmarshalJSON for AbstractNot attempts to interpret a query atom as
// {"not": <abstract query>}.
func (n *AbstractNot) UnmarshalJSON(b []byte) error {
	var data struct {
		Not json.RawMessage `json:"not"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if len(data.Not) == 0 {
		return errors.New(`Missing negation property: "not"`)
	}

	q, err := unmarshalQ(data.Not)
	n.Arg = q
	return err
}

// UnmarshalJSON for AbstractOr attempts to interpret a query atom as
// {"or": [<abstract queries>]}.
func (o *AbstractOr) UnmarshalJSON(b []byte) error {
	var data struct {
		Or []json.RawMessage `json:"or"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if len(data.Or) == 0 {
		return errors.New(`Missing disjunction property: "or"`)
	}

	qs := make([]AbstractQuery, 0, len(data.Or))
	for _, msg := range data.Or {
		q, err := unmarshalQ(msg)
		if err != nil {
			return err
		}
		qs = append(qs, q)
	}
	o.Args = qs
	return nil
}

// UnmarshalJSON for AbstractAnd attempts to interpret a query atom as
// {"and": [<abstract queries>]}.
func (a *AbstractAnd) UnmarshalJSON(b []byte) error {
	var data struct {
		And []json.RawMessage `json:"and"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if len(data.And) == 0 {
		return errors.New(`Missing conjunction property: "and"`)
	}

	qs := make([]AbstractQuery, 0, len(data.And))
	for _, msg := range data.And {
		q, err := unmarshalQ(msg)
		if err != nil {
			return err
		}
		qs = append(qs, q)
	}
	a.Args = qs
	return nil
}

// UnmarshalJSON for AbstractExists attempts to interpret a query atom as
// {"exists": [<abstract queries>]}.
func (e *AbstractExists) UnmarshalJSON(b []byte) error {
	var data struct {
		Exists []json.RawMessage `json:"exists"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if len(data.Exists) == 0 {
		return errors.New(`Missing conjunction property: "exists"`)
	}

	qs := make([]AbstractQuery, 0, len(data.Exists))
	for _, msg := range data.Exists {
		q, err := unmarshalQ(msg)
		if err != nil {
			return err
		}
		qs = append(qs, q)
	}
	e.Args = qs
	return nil
}

// UnmarshalJSON for AbstractAll attempts to interpret a query atom as
// {"all": [<abstract query>]}.
func (e *AbstractAll) UnmarshalJSON(b []byte) error {
	var data struct {
		All []json.RawMessage `json:"all"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if len(data.All) == 0 {
		return errors.New(`Missing conjunction property: "all"`)
	}

	qs := make([]AbstractQuery, 0, len(data.All))
	for _, msg := range data.All {
		q, err := unmarshalQ(msg)
		if err != nil {
			return err
		}
		qs = append(qs, q)
	}
	e.Args = qs
	return nil
}

// UnmarshalJSON for AbstractNone attempts to interpret a query atom as
// {"none": [<abstract query>]}.
func (e *AbstractNone) UnmarshalJSON(b []byte) error {
	var data struct {
		None []json.RawMessage `json:"none"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if len(data.None) == 0 {
		return errors.New(`Missing conjunction property: "none"`)
	}

	qs := make([]AbstractQuery, 0, len(data.None))
	for _, msg := range data.None {
		q, err := unmarshalQ(msg)
		if err != nil {
			return err
		}
		qs = append(qs, q)
	}
	e.Args = qs
	return nil
}

// UnmarshalJSON for AbstractSequential attempts to interpret a query atom as
// {"exists": [<abstract queries>]}.
func (e *AbstractSequential) UnmarshalJSON(b []byte) error {
	var data struct {
		Sequential []json.RawMessage `json:"sequential"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if len(data.Sequential) == 0 {
		return errors.New(`Missing conjunction property: "sequential"`)
	}

	qs := make([]AbstractQuery, 0, len(data.Sequential))
	for _, msg := range data.Sequential {
		q, err := unmarshalQ(msg)
		if err != nil {
			return err
		}
		qs = append(qs, q)
	}
	e.Args = qs
	return nil
}

// UnmarshalJSON for AbstractCount attempts to interpret a query atom as
// {"count": int, "where": query}.
func (c *AbstractCount) UnmarshalJSON(b []byte) (err error) {
	var data struct {
		Count json.RawMessage `json:"count"`
		Where json.RawMessage `json:"where"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if len(data.Count) == 0 {
		return errors.New(`Missing count property: "count"`)
	}
	if len(data.Where) == 0 {
		return errors.New(`Missing count property: "where"`)
	}

	if err := json.Unmarshal(data.Count, &c.Count); err != nil {
		return err
	}
	c.Where, err = unmarshalQ(data.Where)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON for AbstractLessThan attempts to interpret a query atom as
// {"count": int, "where": query}.
func (l *AbstractLessThan) UnmarshalJSON(b []byte) error {
	var data struct {
		Count json.RawMessage `json:"lessThan"`
		Where json.RawMessage `json:"where"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if len(data.Count) == 0 {
		return errors.New(`Missing lessThan property: "lessThan"`)
	}
	if len(data.Where) == 0 {
		return errors.New(`Missing count property: "where"`)
	}

	err := json.Unmarshal(data.Count, &l.Count)
	if err != nil {
		return err
	}
	l.Where, err = unmarshalQ(data.Where)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON for AbstractMoreThan attempts to interpret a query atom as
// {"count": int, "where": query}.
func (m *AbstractMoreThan) UnmarshalJSON(b []byte) (err error) {
	var data struct {
		Count json.RawMessage `json:"moreThan"`
		Where json.RawMessage `json:"where"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if len(data.Count) == 0 {
		return errors.New(`Missing moreThan property: "moreThan"`)
	}
	if len(data.Where) == 0 {
		return errors.New(`Missing count property: "where"`)
	}

	if err := json.Unmarshal(data.Count, &m.Count); err != nil {
		return err
	}
	m.Where, err = unmarshalQ(data.Where)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON for AbstractLink attempts to interpret a query atom as
// {"link":<metadata url pattern string>}.
func (l *AbstractLink) UnmarshalJSON(b []byte) error {
	var data map[string]*json.RawMessage
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	patternMsg, ok := data["link"]
	if !ok {
		return errors.New(`Missing Link pattern property: "link"`)
	}
	var pattern string
	if err := json.Unmarshal(*patternMsg, &pattern); err != nil {
		return errors.New(`Missing link pattern property "pattern" is not a string`)
	}

	l.Pattern = pattern
	return nil
}

// UnmarshalJSON for AbstractTestLabel attempts to interpret a query atom as
// {"label":<label string>}.
func (t *AbstractTestLabel) UnmarshalJSON(b []byte) error {
	var data map[string]*json.RawMessage
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	labelMsg, ok := data["label"]
	if !ok {
		return errors.New(`Missing label pattern property: "label"`)
	}
	var label string
	if err := json.Unmarshal(*labelMsg, &label); err != nil {
		return errors.New(`Property "label" is not a string`)
	}

	t.Label = label
	return nil
}

// UnmarshalJSON for AbstractTriaged attempts to interpret a query atom as
// {"triaged":<browser name>}.
func (t *AbstractTriaged) UnmarshalJSON(b []byte) error {
	var data map[string]*json.RawMessage
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	browserNameMsg, ok := data["triaged"]
	if !ok {
		return errors.New(`Missing Triaged property: "triaged"`)
	}

	var browserName string
	if err := json.Unmarshal(*browserNameMsg, &browserName); err != nil {
		return errors.New(`Triaged property "triaged" is not a string`)
	}

	var product *shared.ProductSpec
	if browserName != "" {
		p, err := shared.ParseProductSpec(browserName)
		if err != nil {
			return err
		}
		product = &p
	}

	t.Product = product
	return nil
}

// UnmarshalJSON for MetadataQuality attempts to interpret a query atom as
// {"is":<metadata quality>}.
func (q *MetadataQuality) UnmarshalJSON(b []byte) (err error) {
	var data map[string]*json.RawMessage
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	is, ok := data["is"]
	if !ok {
		return errors.New(`Missing "is" pattern property: "is"`)
	}
	var quality string
	if err := json.Unmarshal(*is, &quality); err != nil {
		return errors.New(`"is" property is not a string`)
	}

	*q, err = MetadataQualityFromString(quality)
	return err
}

// MetadataQualityFromString returns the enum value for the given string.
func MetadataQualityFromString(quality string) (MetadataQuality, error) {
	switch quality {
	case "different":
		return MetadataQualityDifferent, nil
	case "tentative":
		return MetadataQualityTentative, nil
	case "optional":
		return MetadataQualityOptional, nil
	}
	return MetadataQualityUnknown, fmt.Errorf(`Unknown "is" quality "%s"`, quality)
}

func unmarshalQ(b []byte) (AbstractQuery, error) {
	{
		var tnp TestNamePattern
		if err := json.Unmarshal(b, &tnp); err == nil {
			return tnp, nil
		}
	}
	{
		var stnp SubtestNamePattern
		if err := json.Unmarshal(b, &stnp); err == nil {
			return stnp, nil
		}
	}
	{
		var tp TestPath
		if err := json.Unmarshal(b, &tp); err == nil {
			return tp, nil
		}
	}
	{
		var tse TestStatusEq
		if err := json.Unmarshal(b, &tse); err == nil {
			return tse, nil
		}
	}
	{
		var tsn TestStatusNeq
		if err := json.Unmarshal(b, &tsn); err == nil {
			return tsn, nil
		}
	}
	{
		var n AbstractNot
		if err := json.Unmarshal(b, &n); err == nil {
			return n, nil
		}
	}
	{
		var o AbstractOr
		if err := json.Unmarshal(b, &o); err == nil {
			return o, nil
		}
	}
	{
		var a AbstractAnd
		if err := json.Unmarshal(b, &a); err == nil {
			return a, nil
		}
	}
	{
		var e AbstractExists
		if err := json.Unmarshal(b, &e); err == nil {
			return e, nil
		}
	}
	{
		var a AbstractAll
		if err := json.Unmarshal(b, &a); err == nil {
			return a, nil
		}
	}
	{
		var n AbstractNone
		if err := json.Unmarshal(b, &n); err == nil {
			return n, nil
		}
	}
	{
		var s AbstractSequential
		if err := json.Unmarshal(b, &s); err == nil {
			return s, nil
		}
	}
	{
		var c AbstractCount
		if err := json.Unmarshal(b, &c); err == nil {
			return c, nil
		}
	}
	{
		var c AbstractLessThan
		if err := json.Unmarshal(b, &c); err == nil {
			return c, nil
		}
	}
	{
		var c AbstractMoreThan
		if err := json.Unmarshal(b, &c); err == nil {
			return c, nil
		}
	}
	{
		var l AbstractLink
		if err := json.Unmarshal(b, &l); err == nil {
			return l, nil
		}
	}
	{
		var i MetadataQuality
		if err := json.Unmarshal(b, &i); err == nil {
			return i, nil
		}
	}
	{
		var t AbstractTriaged
		if err := json.Unmarshal(b, &t); err == nil {
			return t, nil
		}
	}
	{
		var t AbstractTestLabel
		if err := json.Unmarshal(b, &t); err == nil {
			return t, nil
		}
	}
	return nil, errors.New(`Failed to parse query fragment as any of the exisiting search atoms in wpt.fyi/api/query/README.md`)
}
