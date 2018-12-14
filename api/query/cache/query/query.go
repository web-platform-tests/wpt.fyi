// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// PrepareUserQuery transforms a user query to a form suitable for planning and
// execution against an index. For example, the query may be extended to filter
// out tests that do not appear in the reports of any test run over which the
// query will be executed. PrepareUserQuery should be executed exactly once on a
// user query.
func PrepareUserQuery(runIDs []int64, q query.ConcreteQuery) query.ConcreteQuery {
	baseQuery := query.Or{
		Args: make([]query.ConcreteQuery, len(runIDs)),
	}
	for i, runID := range runIDs {
		baseQuery.Args[i] = query.Not{
			Arg: query.RunTestStatusConstraint{
				Run:    runID,
				Status: shared.TestStatusUnknown,
			},
		}
	}

	// Add baseQuery to existing AND in q=AND(...), or create AND(baseQuery, q).
	if andQ, ok := q.(query.And); ok {
		andQ.Args = append([]query.ConcreteQuery{baseQuery}, andQ.Args...)
		q = andQ
	} else {
		q = query.And{
			Args: []query.ConcreteQuery{
				baseQuery,
				q,
			},
		}
	}

	return q
}
