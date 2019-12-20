// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/github_oauth_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared GitHubOAuth

package shared

import (
	"context"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

// GitHubOAuth encapsulates implementation details of GitHub OAuth flow.
type GitHubOAuth interface {
	Context() context.Context
	Datastore() Datastore
	GetAccessToken() *string
	SetRedirectURL(url string)
	GetAuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	GetNewClient(oauthToken string) (*github.Client, error)
	GetGitHubUser(client *github.Client) (*github.User, error)
}
