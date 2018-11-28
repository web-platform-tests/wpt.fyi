// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/appengine/datastore"

	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/api/checks/summaries"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// CheckProcessingQueue is the name of the TaskQueue that handles processing and
// interpretation of TestRun results, in order to update the GitHub checks.
const CheckProcessingQueue = "check-processing"

func getOrCreateCheckSuite(ctx context.Context, sha, owner, repo string, installation int64) (*shared.CheckSuite, error) {
	query := datastore.NewQuery("CheckSuite").
		Filter("SHA =", sha).
		Filter("InstallationID =", installation).
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
	suite.InstallationID = installation
	_, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "CheckSuite", nil), &suite)
	if err != nil {
		log.Debugf("Created CheckSuite entity for %s/%s @ %s", owner, repo, sha)
	}
	return &suite, err
}

func getSuitesForSHA(ctx context.Context, sha string) ([]shared.CheckSuite, error) {
	var suites []shared.CheckSuite
	_, err := datastore.NewQuery("CheckSuite").Filter("SHA =", sha).GetAll(ctx, &suites)
	return suites, err
}

func pendingCheckRun(ctx context.Context, sha string, product shared.ProductSpec) (bool, error) {
	host := shared.NewAppEngineAPI(ctx).GetHostname()
	pending := summaries.Pending{
		CheckState: summaries.CheckState{
			Product:    product,
			HeadSHA:    sha,
			Title:      getCheckTitle(product),
			DetailsURL: shared.NewDiffAPI(ctx).GetMasterDiffURL(sha, product),
			Status:     "in_progress",
		},
		HostName: host,
		RunsURL:  fmt.Sprintf("https://%s/runs", host),
	}
	return updateCheckRun(ctx, pending)
}

func completeCheckRun(ctx context.Context, sha string, product shared.ProductSpec) (bool, error) {
	aeAPI := shared.NewAppEngineAPI(ctx)
	host := aeAPI.GetHostname()
	runsURL := aeAPI.GetRunsURL(shared.TestRunFilter{SHA: sha[:10]})
	diffAPI := shared.NewDiffAPI(ctx)
	diffURL := diffAPI.GetMasterDiffURL(sha, product)
	success := "success"
	completed := summaries.Completed{
		CheckState: summaries.CheckState{
			Product:    product,
			HeadSHA:    sha,
			Title:      fmt.Sprintf("wpt.fyi - %s results", product.DisplayName()),
			DetailsURL: diffURL,
			Status:     "completed",
			Conclusion: &success,
		},
		HostName: host,
		HostURL:  fmt.Sprintf("https://%s/", host),
		SHAURL:   runsURL.String(),
		DiffURL:  diffURL.String(),
	}
	return updateCheckRun(ctx, completed)
}

func updateCheckRun(ctx context.Context, summary summaries.Summary) (bool, error) {
	log := shared.GetLogger(ctx)
	state := summary.GetCheckState()
	suites, err := getSuitesForSHA(ctx, state.HeadSHA)
	if err != nil {
		log.Warningf("Failed to load CheckSuites for %s: %s", state.HeadSHA, err.Error())
		return false, err
	} else if len(suites) < 1 {
		log.Debugf("No CheckSuites found for %s", state.HeadSHA)
		return false, nil
	}

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
