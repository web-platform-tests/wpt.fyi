// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/github_util_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared GitHubUtil

package shared

import (
	"context"

	"github.com/google/go-github/v31/github"
)

const sourceOwner string = "web-platform-tests"
const sourceRepo string = "wpt-metadata"
const baseBranch string = "master"

// GitHubUtil encapsulates the implementation details of accessing a wpt-metadata archive.
type GitHubUtil interface {
	getWPTMetadataArchiveLink() (*github.Response, *string, error)
}

type gitHubUtilImpl struct {
	ctx    context.Context
	client *github.Client
}

func (g gitHubUtilImpl) getWPTMetadataArchiveLink() (*github.Response, *string, error) {
	baseRef, _, err := g.client.Git.GetRef(g.ctx, sourceOwner, sourceRepo, "refs/heads/"+baseBranch)
	if err != nil {
		return nil, nil, err
	}

	sha := baseRef.Object.SHA
	_, resp, err := g.client.Repositories.GetArchiveLink(g.ctx, sourceOwner, sourceRepo, "tarball", &github.RepositoryContentGetOptions{Ref: *sha}, true)
	if err != nil {
		return nil, nil, err
	}

	return resp, sha, nil
}

// NewGitHubUtil returns an instance of GitHubUtil.
func NewGitHubUtil(ctx context.Context, client *github.Client) GitHubUtil {
	return gitHubUtilImpl{ctx, client}
}
