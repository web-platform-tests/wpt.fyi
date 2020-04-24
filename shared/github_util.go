// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/github_util_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared GitHubUtil

package shared

import (
	"context"
	"net/http"

	"github.com/google/go-github/v31/github"
)

// GitHubUtil encapsulates implementation details of accessing the wpt-metadata repository.
type GitHubUtil interface {
	getWPTMetadataArchiveLink() (*http.Response, *string, error)
}

type gitHubUtilImpl struct {
	ctx    context.Context
	client *github.Client
}

func (g gitHubUtilImpl) getWPTMetadataArchiveLink() (*http.Response, *string, error) {
	client := tm.githubClient
	var baseRef *github.Reference
	if baseRef, _, err := client.Git.GetRef(g.ctx, g.sourceOwner, g.sourceRepo, "refs/heads/"+tm.baseBranch); err != nil {
		return nil, nil, err
	}

	// Checks redirect.
	_, resp, err := client.Git.GetArchiveLink(g.ctx, g.sourceOwner, g.sourceRepo, g.format, &RepositoryContentGetOptions{Ref: baseRef}, True)
	if err != nil {
		return nil, nil, err
	}

	return resp, baseRef.SHA, nil
}

// NewGitHubUtil returns an instance of GitHubUtil.
func NewGitHubUtil(ctx context.Context) (GitHubUtil, error) {
	return gitHubUtilImpl{}
}
