// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/lukebjerring/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/api/checks/summaries"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/taskqueue"
)

// API abstracts all the API calls used externally.
type API interface {
	Context() context.Context
	ScheduleResultsProcessing(sha string, browser shared.ProductSpec) error
	PendingCheckRun(checkSuite shared.CheckSuite, browser shared.ProductSpec) (bool, error)
	GetSuitesForSHA(sha string) ([]shared.CheckSuite, error)
	IgnoreFailure(sender, owner, repo string, run *github.CheckRun, installation *github.Installation) error
}

type checksAPIImpl struct {
	ctx   context.Context
	queue string
}

// NewAPI returns a real implementation of the API
func NewAPI(ctx context.Context) API {
	return checksAPIImpl{
		ctx:   ctx,
		queue: CheckProcessingQueue,
	}
}

func (s checksAPIImpl) Context() context.Context {
	return s.ctx
}

// ScheduleResultsProcessing adds a URL for callback to TaskQueue for the given sha and
// product, which will actually interpret the results and summarize the outcome.
func (s checksAPIImpl) ScheduleResultsProcessing(sha string, product shared.ProductSpec) error {
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

// PendingCheckRun posts an in_progress check run for the given CheckSuite/Product.
// Returns true if any check_runs were created (i.e. the create succeeded).
func (s checksAPIImpl) PendingCheckRun(suite shared.CheckSuite, product shared.ProductSpec) (bool, error) {
	host := shared.NewAppEngineAPI(s.ctx).GetHostname()
	pending := summaries.Pending{
		CheckState: summaries.CheckState{
			Product:    product,
			HeadSHA:    suite.SHA,
			Title:      getCheckTitle(product),
			DetailsURL: shared.NewDiffAPI(s.ctx).GetMasterDiffURL(suite.SHA, product),
			Status:     "in_progress",
		},
		HostName: host,
		RunsURL:  fmt.Sprintf("https://%s/runs", host),
	}
	return updateCheckRun(s.ctx, pending, suite)
}

// GetSuitesForSHA gets all existing check suites for the given Head SHA
func (s checksAPIImpl) GetSuitesForSHA(sha string) ([]shared.CheckSuite, error) {
	var suites []shared.CheckSuite
	_, err := datastore.NewQuery("CheckSuite").Filter("SHA =", sha).GetAll(s.ctx, &suites)
	return suites, err
}

// IgnoreFailure updates the given CheckRun's outcome to success, even if it failed.
func (s checksAPIImpl) IgnoreFailure(sender, owner, repo string, run *github.CheckRun, installation *github.Installation) error {
	client, err := getGitHubClient(s.ctx, run.GetApp().GetID(), installation.GetID())
	if err != nil {
		return err
	}

	// Keep the previous output, if applicable, but prefix it with an indication that
	// somebody ignored the failure.
	output := run.GetOutput()
	if output == nil {
		output = &github.CheckRunOutput{}
	}
	summary := fmt.Sprintf(`This check was marked as a success by @%s via the _Ignore_ action.

	`, sender) + output.GetSummary()
	output.Summary = &summary

	success := "success"
	opts := github.UpdateCheckRunOptions{
		Output:      output,
		Conclusion:  &success,
		CompletedAt: &github.Timestamp{Time: time.Now()},
	}
	_, _, err = client.Checks.UpdateCheckRun(s.ctx, owner, repo, run.GetID(), opts)
	return err
}

func getCheckTitle(product shared.ProductSpec) string {
	return fmt.Sprintf("wpt.fyi - %s results", product.DisplayName())
}
