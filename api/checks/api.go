// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"fmt"
	"net/url"

	"github.com/web-platform-tests/wpt.fyi/api/checks/summaries"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/taskqueue"
)

// SuitesAPI abstracts all the API calls used externally.
type SuitesAPI interface {
	Context() context.Context
	ScheduleResultsProcessing(sha string, browser shared.ProductSpec) error
	PendingCheckRun(checkSuite shared.CheckSuite, browser shared.ProductSpec) (bool, error)
	GetSuitesForSHA(sha string) ([]shared.CheckSuite, error)
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

// ScheduleResultsProcessing adds a URL for callback to TaskQueue for the given sha and
// product, which will actually interpret the results and summarize the outcome.
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

// PendingCheckRun loads the CheckSuite(s), if any, for the given SHA, and creates
// a pending check_run for the given browser name for each CheckSuite.
// Returns true if any check_runs were created (i.e. any CheckSuite entities were
// found, and the create succeeded).
func (s suitesAPIImpl) PendingCheckRun(suite shared.CheckSuite, product shared.ProductSpec) (bool, error) {
	host := shared.NewAppEngineAPI(s.ctx).GetHostname()
	pending := summaries.Pending{
		CheckState: summaries.CheckState{
			Product:    product,
			HeadSHA:    suite.SHA,
			Title:      getCheckTitle(product),
			DetailsURL: getMasterDiffURL(s.ctx, suite.SHA, product),
			Status:     "in_progress",
		},
		HostName: host,
		RunsURL:  fmt.Sprintf("https://%s/runs", host),
	}
	return updateCheckRun(s.ctx, pending, suite)
}

// GetSuitesForSHA gets all existing check suites for the given Head SHA
func (s suitesAPIImpl) GetSuitesForSHA(sha string) ([]shared.CheckSuite, error) {
	var suites []shared.CheckSuite
	_, err := datastore.NewQuery("CheckSuite").Filter("SHA =", sha).GetAll(s.ctx, &suites)
	return suites, err
}
