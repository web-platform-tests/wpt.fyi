// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/web-platform-tests/wpt.fyi/api/checks/summaries"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func updateCheckRunSummary(ctx context.Context, summary summaries.Summary, suite shared.CheckSuite) (bool, error) {
	log := shared.GetLogger(ctx)
	product := summary.GetCheckState().Product
	testRun := summary.GetCheckState().TestRun

	// Attempt to update any existing check runs for this SHA.
	checkRuns, err := getExistingCheckRuns(ctx, suite)
	if err != nil {
		log.Warningf("Failed to load existing check runs for %s: %s", suite.SHA[:7], err.Error())
	}

	// Update, not create, if a run name matches this completed TestRun.
	var existing *github.CheckRun
	if testRun != nil {
		for _, run := range checkRuns {
			if run.GetApp().GetID() != suite.AppID {
				continue
			}
			if spec, _ := shared.ParseProductSpec(run.GetName()); spec.Matches(*testRun) {
				log.Debugf("Found existing run %v for %s @ %s", run.GetID(), run.GetName(), suite.SHA[:7])
				existing = run

				break
			}
		}
	}

	var created bool
	// nolint:nestif // TODO: Fix nestif lint error
	if existing != nil {
		created, err = updateExistingCheckRunSummary(ctx, summary, suite, existing)
		if err != nil {
			log.Warningf("Failed to update existing check run summary for %s: %s", *existing.HeadSHA, err.Error())
		}
	} else {
		state := summary.GetCheckState()
		actions := summary.GetActions()

		var summaryStr string
		summaryStr, err = summary.GetSummary()
		if err != nil {
			log.Warningf("Failed to generate summary for %s: %s", state.HeadSHA, err.Error())

			return false, err
		}

		detailsURLStr := state.DetailsURL.String()
		title := state.Title()
		// nolint:exhaustruct // WONTFIX: Name, HeadSHA only required.
		opts := github.CreateCheckRunOptions{
			Name:       state.Name(),
			HeadSHA:    state.HeadSHA,
			DetailsURL: &detailsURLStr,
			Status:     &state.Status,
			Conclusion: state.Conclusion,
			Output: &github.CheckRunOutput{
				Title:   &title,
				Summary: &summaryStr,
			},
			Actions: actions,
		}
		if state.Conclusion != nil {
			opts.CompletedAt = &github.Timestamp{Time: time.Now()}
		}
		created, err = createCheckRun(ctx, suite, opts)
		if err != nil {
			log.Warningf("Failed to create check run summary for %s: %s", suite.SHA, err.Error())
		}
	}
	if created {
		log.Debugf("Check for %s/%s @ %s (%s) updated", suite.Owner, suite.Repo, suite.SHA[:7], product.String())
	}

	return created, nil
}

func getExistingCheckRuns(ctx context.Context, suite shared.CheckSuite) ([]*github.CheckRun, error) {
	log := shared.GetLogger(ctx)
	client, err := getGitHubClient(ctx, suite.AppID, suite.InstallationID)
	if err != nil {
		log.Errorf("Failed to fetch runs for suite: %s", err.Error())

		return nil, err
	}

	var runs []*github.CheckRun
	// nolint:exhaustruct // TODO: Fix exhaustruct lint error.
	options := github.ListCheckRunsOptions{
		// nolint:exhaustruct // TODO: Fix exhaustruct lint error.
		ListOptions: github.ListOptions{
			// 100 is the maximum allowed items per page; see
			// https://developer.github.com/v3/guides/traversing-with-pagination/#changing-the-number-of-items-received
			PerPage: 100,
		},
	}

	// As a safety-check, we will not do more than 10 iterations (at 100
	// check runs per page, this gives us a 1000 run upper limit).
	for i := 0; i < 10; i++ {
		result, response, err := client.Checks.ListCheckRunsForRef(ctx, suite.Owner, suite.Repo, suite.SHA, &options)
		if err != nil {
			return nil, err
		}

		runs = append(runs, result.CheckRuns...)

		// GitHub APIs indicate being on the last page by not returning any
		// value for NextPage, which go-github translates into zero.
		// See https://gowalker.org/github.com/google/go-github/github#Response
		if response.NextPage == 0 {
			return runs, nil
		}

		// Setup for the next call.
		options.Page = response.NextPage
	}

	return nil, fmt.Errorf("more than 10 pages of CheckRuns returned for ref %s", suite.SHA)
}

func updateExistingCheckRunSummary(
	ctx context.Context,
	summary summaries.Summary,
	suite shared.CheckSuite,
	run *github.CheckRun,
) (bool, error) {
	log := shared.GetLogger(ctx)

	state := summary.GetCheckState()
	actions := summary.GetActions()

	summaryStr, err := summary.GetSummary()
	if err != nil {
		log.Warningf("Failed to generate summary for %s: %s", state.HeadSHA, err.Error())

		return false, err
	}

	detailsURLStr := state.DetailsURL.String()
	title := state.Title()
	// nolint:exhaustruct // WONTFIX: Name, HeadSHA only required.
	opts := github.UpdateCheckRunOptions{
		Name:       state.Name(),
		DetailsURL: &detailsURLStr,
		Status:     &state.Status,
		Conclusion: state.Conclusion,
		Output: &github.CheckRunOutput{
			Title:   &title,
			Summary: &summaryStr,
		},
		Actions: actions,
	}
	if state.Conclusion != nil {
		opts.CompletedAt = &github.Timestamp{Time: time.Now()}
	}

	client, err := getGitHubClient(ctx, suite.AppID, suite.InstallationID)
	if err != nil {
		return false, err
	}

	_, _, err = client.Checks.UpdateCheckRun(ctx, suite.Owner, suite.Repo, run.GetID(), opts)
	if err != nil {
		log.Errorf("Failed to update run %v: %s", run.GetID(), err.Error())

		return false, err
	}

	return true, err
}
