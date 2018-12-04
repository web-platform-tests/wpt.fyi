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

func updateCheckRun(ctx context.Context, summary summaries.Summary, suites ...shared.CheckSuite) (bool, error) {
	log := shared.GetLogger(ctx)
	state := summary.GetCheckState()
	actions := summary.GetActions()

	summaryStr, err := summary.GetSummary()
	if err != nil {
		log.Warningf("Failed to generate summary for %s: %s", state.HeadSHA, err.Error())
		return false, err
	}

	detailsURLStr := state.DetailsURL.String()
	opts := github.CreateCheckRunOptions{
		Name:       state.Product.String(),
		HeadSHA:    state.HeadSHA,
		DetailsURL: &detailsURLStr,
		Status:     &state.Status,
		Conclusion: state.Conclusion,
		Output: &github.CheckRunOutput{
			Title:   &state.Title,
			Summary: &summaryStr,
		},
		Actions: actions,
	}
	if state.Conclusion != nil {
		opts.CompletedAt = &github.Timestamp{Time: time.Now()}
	}

	for _, suite := range suites {
		created, err := createCheckRun(ctx, suite, opts)
		if !created || err != nil {
			return false, err
		}
		log.Debugf("Check for %s/%s @ %s (%s) updated", suite.Owner, suite.Repo, suite.SHA[:7], state.Product.String())
	}
	return true, nil
}
