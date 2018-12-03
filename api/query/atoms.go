// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

var browsers = shared.GetDefaultBrowserNames()

// AbstractQuery is an intermetidate representation of a test results query that
//  has not been bound to specific shared.TestRun specs for processing.
type AbstractQuery interface {
	BindToRuns(runs []shared.TestRun) ConcreteQuery
}

// RunQuery is the internal representation of a query recieved from an HTTP
// client, including the IDs of the test runs to query, and the structured query
// to run.
type RunQuery struct {
	runIDs []int64
	AbstractQuery
}

// TestNamePattern is a query atom that matches test names to a pattern string.
type TestNamePattern struct {
	Pattern string
}

// BindToRuns for TestNamePattern is a no-op: TestNamePattern implements both
// AbstractQuery and ConcreteQuery because it is independent of test runs.
func (tnp TestNamePattern) BindToRuns(runs []shared.TestRun) ConcreteQuery {
	return tnp
}

type testStatusConstraint struct {
	browserName string
	status      int64
}

func (tsc testStatusConstraint) BindToRuns(runs []shared.TestRun) ConcreteQuery {
	ids := make([]int64, 0, len(runs))
	for _, run := range runs {
		if run.BrowserName == tsc.browserName {
			ids = append(ids, run.ID)
		}
	}
	if len(ids) == 0 {
		return True{}
	}
	if len(ids) == 1 {
		return RunTestStatusConstraint{ids[0], tsc.status}
	}

	q := Or{make([]ConcreteQuery, len(ids))}
	for i := range ids {
		q.Args[i] = RunTestStatusConstraint{ids[i], tsc.status}
	}
	return q
}

type not struct {
	not AbstractQuery
}

func (n not) BindToRuns(runs []shared.TestRun) ConcreteQuery {
	return Not{n.not.BindToRuns(runs)}
}

type or struct {
	or []AbstractQuery
}

func (o or) BindToRuns(runs []shared.TestRun) ConcreteQuery {
	q := Or{make([]ConcreteQuery, 0, len(o.or))}
	for i := range o.or {
		sub := o.or[i].BindToRuns(runs)
		if _, ok := sub.(True); ok {
			return True{}
		}
		if _, ok := sub.(False); ok {
			continue
		}
		q.Args = append(q.Args, sub)
	}
	if len(q.Args) == 0 {
		return True{}
	}
	if len(q.Args) == 1 {
		return q.Args[0]
	}
	return q
}

type and struct {
	and []AbstractQuery
}

func (a and) BindToRuns(runs []shared.TestRun) ConcreteQuery {
	q := And{make([]ConcreteQuery, 0, len(a.and))}
	for i := range a.and {
		sub := a.and[i].BindToRuns(runs)
		if _, ok := sub.(False); ok {
			return False{}
		}
		if _, ok := sub.(True); ok {
			continue
		}
		q.Args = append(q.Args, sub)
	}
	if len(q.Args) == 0 {
		return True{}
	}
	if len(q.Args) == 1 {
		return q.Args[0]
	}
	return q
}

// UnmarshalJSON interprets the JSON representation of a RunQuery, instantiating
// (an) appropriate Query implementation(s) according to the JSON structure.
func (rq *RunQuery) UnmarshalJSON(b []byte) error {
	var data struct {
		RunIDs []int64         `json:"run_ids"`
		Query  json.RawMessage `json:"query"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if len(data.RunIDs) == 0 {
		return errors.New(`Missing run query property: "run_ids"`)
	}
	if len(data.Query) == 0 {
		return errors.New(`Missing run query property: "query"`)
	}

	q, err := unmarshalQ(data.Query)
	if err != nil {
		return err
	}

	rq.runIDs = data.RunIDs
	rq.AbstractQuery = q
	return nil
}

// UnmarshalJSON for TestNamePattern attempts to interpret a query atom as as
// {"pattern":<test name pattern string>}.
func (tnp *TestNamePattern) UnmarshalJSON(b []byte) error {
	var data struct {
		Pattern string `json:"pattern"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if len(data.Pattern) == 0 {
		return errors.New(`Missing testn mae pattern property: "pattern"`)
	}

	tnp.Pattern = data.Pattern
	return nil
}

func (tsc *testStatusConstraint) UnmarshalJSON(b []byte) error {
	var data struct {
		BrowserName string `json:"browser_name"`
		Status      string `json:"status"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if len(data.BrowserName) == 0 {
		return errors.New(`Missing test status constraint property: "browser_name"`)
	}
	if len(data.Status) == 0 {
		return errors.New(`Missing test status constraint property: "status"`)
	}

	browserName := strings.ToLower(data.BrowserName)
	browserNameOK := false
	for _, name := range browsers {
		browserNameOK = browserNameOK || browserName == name
	}
	if !browserNameOK {
		return fmt.Errorf(`Invalid browser name: "%s"`, data.BrowserName)
	}

	statusStr := strings.ToUpper(data.Status)
	status := shared.TestStatusValueFromString(statusStr)
	statusStr2 := shared.TestStatusStringFromValue(status)
	if statusStr != statusStr2 {
		return fmt.Errorf(`Invalid test status: "%s"`, data.Status)
	}

	tsc.browserName = browserName
	tsc.status = status
	return nil
}

func (n *not) UnmarshalJSON(b []byte) error {
	var data struct {
		Not json.RawMessage `json:"not"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if len(data.Not) == 0 {
		return errors.New(`Missing negation property: "not"`)
	}

	q, err := unmarshalQ(data.Not)
	n.not = q
	return err
}

func (o *or) UnmarshalJSON(b []byte) error {
	var data struct {
		Or []json.RawMessage `json:"or"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
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
	o.or = qs
	return nil
}

func (a *and) UnmarshalJSON(b []byte) error {
	var data struct {
		And []json.RawMessage `json:"and"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
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
	a.and = qs
	return nil
}

func unmarshalQ(b []byte) (AbstractQuery, error) {
	var tnp TestNamePattern
	err := json.Unmarshal(b, &tnp)
	if err == nil {
		return tnp, nil
	}
	var tsc testStatusConstraint
	err = json.Unmarshal(b, &tsc)
	if err == nil {
		return tsc, nil
	}
	var n not
	err = json.Unmarshal(b, &n)
	if err == nil {
		return n, nil
	}
	var o or
	err = json.Unmarshal(b, &o)
	if err == nil {
		return o, nil
	}
	var a and
	err = json.Unmarshal(b, &a)
	if err == nil {
		return a, nil
	}

	return nil, errors.New(`Failed to parse query fragment as test name pattern, test status constraint, negation, disjunction, or conjunction`)
}
