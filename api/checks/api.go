// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"fmt"
	"net/url"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/taskqueue"
)

// SuitesAPI abstracts all the API calls used externally.
type SuitesAPI interface {
	Context() context.Context

	// PendingCheckRun loads the CheckSuite(s), if any, for the given SHA, and creates
	// a pending check_run for the given browser name for each CheckSuite.
	// Returns true if any check_runs were created (i.e. any CheckSuite entities were
	// found, and the create succeeded).
	PendingCheckRun(sha string, browser shared.ProductSpec) (bool, error)

	// CompleteCheckRun loads the CheckSuite(s), if any, for the given SHA, and creates
	// a complete check_run for the given browser on GitHub.
	// Returns true if any check_runs were created (i.e. any CheckSuite entities were
	// found, and the create succeeded).
	CompleteCheckRun(sha string, browser shared.ProductSpec) (bool, error)

	// ScheduleResultsProcessing adds a URL for callback to TaskQueue for the given sha and
	// product, which will actually interpret the results and summarize the outcome.
	ScheduleResultsProcessing(sha string, browser shared.ProductSpec) error
}

type suitesAPIImpl struct {
	ctx   context.Context
	queue string
}

// NewSuitesAPI returns a real implementation of the SuitesAPI
func NewSuitesAPI(ctx context.Context) SuitesAPI {
	return suitesAPIImpl{
		ctx:   ctx,
		queue: CheckProcessingQueue,
	}
}

func (s suitesAPIImpl) Context() context.Context {
	return s.ctx
}

func (s suitesAPIImpl) PendingCheckRun(sha string, product shared.ProductSpec) (bool, error) {
	return pendingCheckRun(s.ctx, sha, product)
}

func (s suitesAPIImpl) CompleteCheckRun(sha string, product shared.ProductSpec) (bool, error) {
	return completeCheckRun(s.ctx, sha, product)
}

func (s suitesAPIImpl) ScheduleResultsProcessing(sha string, product shared.ProductSpec) error {
	log := shared.GetLogger(s.ctx)
	target := fmt.Sprintf("/api/checks/%s", sha)
	q := url.Values{}
	q.Set("product", product.String())
	t := taskqueue.NewPOSTTask(target, q)
	t, err := taskqueue.Add(s.ctx, t, s.queue)
	if err != nil {
		log.Warningf("Failed to queue %s @ %s: %s", product.String(), sha[:7], err.Error())
	} else {
		log.Infof("Added %s @ %s to checks processing queue", product.String(), sha[:7])
	}
	return err
}
