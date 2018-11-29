// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"time"

	"google.golang.org/appengine/datastore"

	"github.com/lukebjerring/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/api/checks/summaries"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// CheckProcessingQueue is the name of the TaskQueue that handles processing and
// interpretation of TestRun results, in order to update the GitHub checks.
const CheckProcessingQueue = "check-processing"

func getOrCreateCheckSuite(ctx context.Context, sha, owner, repo string, appID, installationID int64) (*shared.CheckSuite, error) {
	query := datastore.NewQuery("CheckSuite").
		Filter("SHA =", sha).
		Filter("AppID =", appID).
		Filter("InstallationID =", installationID).
		Filter("Owner =", owner).
		Filter("Repo =", repo).
		KeysOnly()
	var suite shared.CheckSuite
	if keys, err := query.GetAll(ctx, nil); err != nil {
		return nil, err
	} else if len(keys) > 0 {
		err := datastore.Get(ctx, keys[0], &suite)
		return &suite, err
	}

	log := shared.GetLogger(ctx)
	suite.SHA = sha
	suite.Owner = owner
	suite.Repo = repo
	suite.AppID = appID
	suite.InstallationID = installationID
	_, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "CheckSuite", nil), &suite)
	if err != nil {
		log.Debugf("Created CheckSuite entity for %s/%s @ %s", owner, repo, sha)
	}
	return &suite, err
}

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
