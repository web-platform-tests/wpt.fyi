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
	RunIDs []int64
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

// TestStatusEq is a query atom that matches tests where the test status/result
// from at least one test run matches the given status value, optionally filtered
// to a specific browser name.
type TestStatusEq struct {
	BrowserName string
	Status      int64
}

// TestStatusNeq is a query atom that matches tests where the test status/result
// from at least one test run does not match the given status value, optionally
// filtered to a specific browser name.
type TestStatusNeq struct {
	BrowserName string
	Status      int64
}

// BindToRuns for TestStatusEq expands to a disjunction of RunTestStatusEq
// values.
func (tse TestStatusEq) BindToRuns(runs []shared.TestRun) ConcreteQuery {
	ids := make([]int64, 0, len(runs))
	for _, run := range runs {
		if tse.BrowserName == "" || run.BrowserName == tse.BrowserName {
			ids = append(ids, run.ID)
		}
	}
	if len(ids) == 0 {
		return True{}
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
func (tsn TestStatusNeq) BindToRuns(runs []shared.TestRun) ConcreteQuery {
	ids := make([]int64, 0, len(runs))
	for _, run := range runs {
		if tsn.BrowserName == "" || run.BrowserName == tsn.BrowserName {
			ids = append(ids, run.ID)
		}
	}
	if len(ids) == 0 {
		return True{}
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
func (n AbstractNot) BindToRuns(runs []shared.TestRun) ConcreteQuery {
	return Not{n.Arg.BindToRuns(runs)}
}

// AbstractOr is the AbstractQuery for disjunction.
type AbstractOr struct {
	Args []AbstractQuery
}

// BindToRuns for AbstractOr produces an Or with bound arguments.
func (o AbstractOr) BindToRuns(runs []shared.TestRun) ConcreteQuery {
	args := make([]ConcreteQuery, 0, len(o.Args))
	for i := range o.Args {
		sub := o.Args[i].BindToRuns(runs)
		if _, ok := sub.(True); ok {
			return True{}
		}
		if _, ok := sub.(False); ok {
			continue
		}
		args = append(args, sub)
	}
	if len(args) == 0 {
		return True{}
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
func (a AbstractAnd) BindToRuns(runs []shared.TestRun) ConcreteQuery {
	args := make([]ConcreteQuery, 0, len(a.Args))
	for i := range a.Args {
		sub := a.Args[i].BindToRuns(runs)
		if _, ok := sub.(False); ok {
			return False{}
		}
		if _, ok := sub.(True); ok {
			continue
		}
		args = append(args, sub)
	}
	if len(args) == 0 {
		return True{}
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

	rq.RunIDs = data.RunIDs
	rq.AbstractQuery = q
	return nil
}

// UnmarshalJSON for TestNamePattern attempts to interpret a query atom as
// {"pattern":<test name pattern string>}.
func (tnp *TestNamePattern) UnmarshalJSON(b []byte) error {
	var data map[string]*json.RawMessage
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	patternMsg, ok := data["pattern"]
	if !ok {
		return errors.New(`Missing test name pattern property: "pattern"`)
	}
	var pattern string
	if err := json.Unmarshal(*patternMsg, &pattern); err != nil {
		return errors.New(`Missing test name pattern property "pattern" is not a string`)
	}

	tnp.Pattern = pattern
	return nil
}

// UnmarshalJSON for TestStatusEq attempts to interpret a query atom as
// {"browser_name": <browser name>, "status": <status string>}.
func (tse *TestStatusEq) UnmarshalJSON(b []byte) error {
	var data struct {
		BrowserName string `json:"browser_name"`
		Status      string `json:"status"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
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

	tse.BrowserName = browserName
	tse.Status = status
	return nil
}

// UnmarshalJSON for TestStatusNeq attempts to interpret a query atom as
// {"browser_name": <browser name>, "status": {"not": <status string>}}.
func (tsn *TestStatusNeq) UnmarshalJSON(b []byte) error {
	var data struct {
		BrowserName string `json:"browser_name"`
		Status      struct {
			Not string `json:"not"`
		} `json:"status"`
	}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if len(data.Status.Not) == 0 {
		return errors.New(`Missing test status constraint property: "status.not"`)
	}

	browserName := strings.ToLower(data.BrowserName)
	browserNameOK := false
	for _, name := range browsers {
		browserNameOK = browserNameOK || browserName == name
	}
	if !browserNameOK {
		return fmt.Errorf(`Invalid browser name: "%s"`, data.BrowserName)
	}

	statusStr := strings.ToUpper(data.Status.Not)
	status := shared.TestStatusValueFromString(statusStr)
	statusStr2 := shared.TestStatusStringFromValue(status)
	if statusStr != statusStr2 {
		return fmt.Errorf(`Invalid test status: "%s"`, data.Status)
	}

	tsn.BrowserName = browserName
	tsn.Status = status
	return nil
}

// UnmarshalJSON for AbstractNot attempts to interpret a query atom as
// {"not": <abstract query>}.
func (n *AbstractNot) UnmarshalJSON(b []byte) error {
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
	n.Arg = q
	return err
}

// UnmarshalJSON for AbstractOr attempts to interpret a query atom as
// {"or": [<abstract queries>]}.
func (o *AbstractOr) UnmarshalJSON(b []byte) error {
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
	o.Args = qs
	return nil
}

// UnmarshalJSON for AbstractAnd attempts to interpret a query atom as
// {"and": [<abstract queries>]}.
func (a *AbstractAnd) UnmarshalJSON(b []byte) error {
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
	a.Args = qs
	return nil
}

func unmarshalQ(b []byte) (AbstractQuery, error) {
	var tnp TestNamePattern
	err := json.Unmarshal(b, &tnp)
	if err == nil {
		return tnp, nil
	}
	var tse TestStatusEq
	err = json.Unmarshal(b, &tse)
	if err == nil {
		return tse, nil
	}
	var tsn TestStatusNeq
	err = json.Unmarshal(b, &tsn)
	if err == nil {
		return tsn, nil
	}
	var n AbstractNot
	err = json.Unmarshal(b, &n)
	if err == nil {
		return n, nil
	}
	var o AbstractOr
	err = json.Unmarshal(b, &o)
	if err == nil {
		return o, nil
	}
	var a AbstractAnd
	err = json.Unmarshal(b, &a)
	if err == nil {
		return a, nil
	}

	return nil, errors.New(`Failed to parse query fragment as test name pattern, test status constraint, negation, disjunction, or conjunction`)
}
