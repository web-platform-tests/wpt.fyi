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

// GitHubUtil encapsulates the implementation of accessing the information in the wpt-metadata repo.
type GitHubUtil interface {
	GetWPTMetadataMasterSHA() (*string, error)
}

type gitHubUtilImpl struct {
	ctx    context.Context
	client *github.Client
}

func (g gitHubUtilImpl) GetWPTMetadataMasterSHA() (*string, error) {
	baseRef, _, err := g.client.Git.GetRef(g.ctx, sourceOwner, sourceRepo, "refs/heads/"+baseBranch)
	if err != nil {
		return nil, err
	}

	return baseRef.Object.SHA, nil
}

// GetGitHubUtil returns an instance of the GitHubUtil interface.
func GetGitHubUtil(ctx context.Context, client *github.Client) GitHubUtil {
	return gitHubUtilImpl{ctx, client}
}
