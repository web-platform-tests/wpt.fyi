// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -build_flags=--mod=mod -destination mock_checks/api_mock.go github.com/web-platform-tests/wpt.fyi/api/checks API

package checks

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/go-github/v82/github"
	"github.com/web-platform-tests/wpt.fyi/api/checks/summaries"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

const (
	wptfyiCheckAppID        = int64(23318) // https://github.com/apps/wpt-fyi-status-check
	wptfyiStagingCheckAppID = int64(19965) // https://github.com/apps/staging-wpt-fyi-status-check

	wptRepoInstallationID        = int64(577173)
	wptRepoStagingInstallationID = int64(449270)

	wptRepoID                = int64(3618133)
	checksForAllUsersFeature = "checksAllUsers"
)

// API abstracts all the API calls used externally.
type API interface {
	shared.AppEngineAPI

	ScheduleResultsProcessing(sha string, browser shared.ProductSpec) error
	GetSuitesForSHA(sha string) ([]shared.CheckSuite, error)
	IgnoreFailure(sender, owner, repo string, run *github.CheckRun, installation *github.Installation) error
	CancelRun(sender, owner, repo string, run *github.CheckRun, installation *github.Installation) error
	CreateWPTCheckSuite(appID, installationID int64, sha string, prNumbers ...int) (bool, error)
	GetWPTRepoAppInstallationIDs() (appID, installationID int64)
}

type checksAPIImpl struct {
	shared.AppEngineAPI

	queue string
}

// NewAPI returns a real implementation of the API.
// nolint:ireturn // TODO: Fix ireturn lint error
func NewAPI(ctx context.Context) API {
	return checksAPIImpl{
		AppEngineAPI: shared.NewAppEngineAPI(ctx),
		queue:        CheckProcessingQueue,
	}
}

// ScheduleResultsProcessing adds a URL for callback to TaskQueue for the given sha and
// product, which will actually interpret the results and summarize the outcome.
func (s checksAPIImpl) ScheduleResultsProcessing(sha string, product shared.ProductSpec) error {
	log := shared.GetLogger(s.Context())
	target := fmt.Sprintf("/api/checks/%s", sha)
	q := url.Values{}
	q.Set("product", product.String())
	_, err := s.ScheduleTask(s.queue, "", target, q)
	if err != nil {
		log.Warningf("Failed to queue %s @ %s: %s", product.String(), sha[:7], err.Error())
	} else {
		log.Infof("Added %s @ %s to checks processing queue", product.String(), sha[:7])
	}

	return err
}

// GetSuitesForSHA gets all existing check suites for the given Head SHA.
func (s checksAPIImpl) GetSuitesForSHA(sha string) ([]shared.CheckSuite, error) {
	var suites []shared.CheckSuite
	store := shared.NewAppEngineDatastore(s.Context(), false)
	_, err := store.GetAll(store.NewQuery("CheckSuite").Filter("SHA =", sha), &suites)

	return suites, err
}

// IgnoreFailure updates the given CheckRun's outcome to success, even if it failed.
func (s checksAPIImpl) IgnoreFailure(
	sender,
	owner, repo string,
	run *github.CheckRun,
	installation *github.Installation,
) error {
	client, err := getGitHubClient(s.Context(), run.GetApp().GetID(), installation.GetID())
	if err != nil {
		return err
	}

	// Keep the previous output, if applicable, but prefix it with an indication that
	// somebody ignored the failure.
	output := run.GetOutput()
	if output == nil {
		// nolint:exhaustruct // TODO: Fix exhaustruct lint error.
		output = &github.CheckRunOutput{}
	}
	prepend := fmt.Sprintf("This check was marked as a success by @%s via the _Ignore_ action.\n\n", sender)
	summary := prepend + output.GetSummary()
	output.Summary = &summary

	success := "success"
	// nolint:exhaustruct // WONTFIX: Name only required.
	opts := github.UpdateCheckRunOptions{
		Name:        run.GetName(),
		Output:      output,
		Conclusion:  &success,
		CompletedAt: &github.Timestamp{Time: time.Now()},
		Actions: []*github.CheckRunAction{
			summaries.RecomputeAction(),
		},
	}
	_, _, err = client.Checks.UpdateCheckRun(s.Context(), owner, repo, run.GetID(), opts)

	return err
}

// CancelRun updates the given CheckRun's outcome to cancelled, even if it failed.
func (s checksAPIImpl) CancelRun(
	sender,
	owner,
	repo string,
	run *github.CheckRun,
	installation *github.Installation,
) error {
	client, err := getGitHubClient(s.Context(), run.GetApp().GetID(), installation.GetID())
	if err != nil {
		return err
	}

	// Keep the previous output, if applicable, but prefix it with an indication that
	// somebody ignored the failure.
	summary := fmt.Sprintf("This check was cancelled by @%s via the _Cancel_ action.", sender)
	title := run.GetOutput().GetTitle()
	// nolint:exhaustruct // TODO: Fix exhaustruct lint error.
	output := &github.CheckRunOutput{
		Title:   &title,
		Summary: &summary,
	}

	cancelled := "cancelled"
	// nolint:exhaustruct // WONTFIX: Name only required.
	opts := github.UpdateCheckRunOptions{
		Name:        run.GetName(),
		Output:      output,
		Conclusion:  &cancelled,
		CompletedAt: &github.Timestamp{Time: time.Now()},
		Actions: []*github.CheckRunAction{
			summaries.RecomputeAction(),
			summaries.IgnoreAction(),
		},
	}
	_, _, err = client.Checks.UpdateCheckRun(s.Context(), owner, repo, run.GetID(), opts)

	return err
}

// CreateWPTCheckSuite creates a check_suite on the main wpt repo for the given
// SHA. This is needed when a PR comes from a different fork of the repo.
func (s checksAPIImpl) CreateWPTCheckSuite(appID, installationID int64, sha string, prNumbers ...int) (bool, error) {
	log := shared.GetLogger(s.Context())
	log.Debugf("Creating check_suite for web-platform-tests/wpt @ %s", sha)

	client, err := getGitHubClient(s.Context(), appID, installationID)
	if err != nil {
		return false, err
	}

	// nolint:exhaustruct // WONTFIX: HeadSHA only required.
	opts := github.CreateCheckSuiteOptions{
		HeadSHA: sha,
	}
	suite, _, err := client.Checks.CreateCheckSuite(s.Context(), shared.WPTRepoOwner, shared.WPTRepoName, opts)
	if err != nil {
		log.Errorf("Failed to create GitHub check suite: %s", err.Error())
	} else if suite != nil {
		log.Infof("check_suite %v created", suite.GetID())
		_, err = getOrCreateCheckSuite(
			s.Context(),
			sha,
			shared.WPTRepoOwner,
			shared.WPTRepoName,
			appID,
			installationID,
			prNumbers...,
		)
		if err != nil {
			log.Infof("Error while getting check suite: %s", err.Error())
		}
	}

	return suite != nil, err
}

func (s checksAPIImpl) GetWPTRepoAppInstallationIDs() (appID, installationID int64) {
	// Production
	if s.GetHostname() == "wpt.fyi" {
		return wptfyiCheckAppID, wptRepoInstallationID
	}
	// Default to staging
	return wptfyiStagingCheckAppID, wptRepoStagingInstallationID
}
