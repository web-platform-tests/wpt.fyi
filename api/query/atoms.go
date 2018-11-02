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

type query interface {
	toPlan(runIDs []int64) plan
}

type runQuery struct {
	runIDs []int64
	query
}

func (rq runQuery) toPlan() plan {
	return rq.query.toPlan(rq.runIDs)
}

type testNamePattern struct {
	pattern string
}

func (tnp testNamePattern) toPlan(runIDs []int64) plan {
	return nil
}

type testStatusConstraint struct {
	browserName string
	status      int64
}

func (tsc testStatusConstraint) toPlan(runIDs []int64) plan {
	return nil
}

type not struct {
	not query
}

func (n not) toPlan(runIDs []int64) plan {
	return nil
}

type or struct {
	or []query
}

func (o or) toPlan(runIDs []int64) plan {
	return nil
}

type and struct {
	and []query
}

func (a and) toPlan(runIDs []int64) plan {
	return nil
}

func (rq *runQuery) UnmarshalJSON(b []byte) error {
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
	rq.query = q
	return nil
}

func (tnp *testNamePattern) UnmarshalJSON(b []byte) error {
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

	tnp.pattern = data.Pattern
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

	qs := make([]query, 0, len(data.Or))
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

	qs := make([]query, 0, len(data.And))
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

func unmarshalQ(b []byte) (query, error) {
	var tnp testNamePattern
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
