// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"time"

	"github.com/google/go-github/github"
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
	if existing != nil {
		created, err = updateExistingCheckRunSummary(ctx, summary, suite, existing)
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

	runs, _, err := client.Checks.ListCheckRunsForRef(ctx, suite.Owner, suite.Repo, suite.SHA, nil)
	if err != nil {
		return nil, err
	}
	return runs.CheckRuns, nil
}

func updateExistingCheckRunSummary(ctx context.Context, summary summaries.Summary, suite shared.CheckSuite, run *github.CheckRun) (bool, error) {
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
	opts := github.UpdateCheckRunOptions{
		Name:       state.Name(),
		HeadSHA:    &state.HeadSHA,
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
