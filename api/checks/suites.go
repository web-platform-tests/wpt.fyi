// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"time"

	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/datastore"
)

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

func pendingCheckRun(ctx context.Context, sha, browser string) (bool, error) {
	suites, err := getSuitesForSHA(ctx, sha)
	if err != nil {
		return false, err
	} else if len(suites) < 1 {
		return false, nil
	}
	detailsURL := getDetailsURL(ctx, sha, browser)
	detailsURLStr := detailsURL.String()

	status := "in_progress"
	opts := github.CreateCheckRunOptions{
		Name:       browser,
		HeadSHA:    sha,
		DetailsURL: &detailsURLStr,
		Status:     &status,
		StartedAt:  &github.Timestamp{Time: time.Now()},
	}

	for _, suite := range suites {
		created, err := createCheckRun(ctx, suite, opts)
		if !created || err != nil {
			return false, err
		}
	}
	return true, nil
}

func completeCheckRun(ctx context.Context, sha, browser string) (bool, error) {
	suites, err := getSuitesForSHA(ctx, sha)
	if err != nil {
		return false, err
	} else if len(suites) < 1 {
		return false, nil
	}
	detailsURL := getDetailsURL(ctx, sha, browser)
	detailsURLStr := detailsURL.String()

	status := "completed"
	conclusion := "success"
	opts := github.CreateCheckRunOptions{
		Name:        browser,
		HeadSHA:     sha,
		DetailsURL:  &detailsURLStr,
		Status:      &status,
		Conclusion:  &conclusion,
		CompletedAt: &github.Timestamp{Time: time.Now()},
	}
	for _, suite := range suites {
		created, err := createCheckRun(ctx, suite, opts)
		if !created || err != nil {
			return false, err
		}
	}
	return true, nil
}
