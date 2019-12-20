// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/github_oauth_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared GitHubOAuth

package shared

import (
	"context"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	ghOAuth "golang.org/x/oauth2/github"
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

type githubOAuthImp struct {
	ctx         context.Context
	ds          Datastore
	conf        *oauth2.Config
	accessToken *string
}

func (g *githubOAuthImp) Datastore() Datastore {
	return g.ds
}

func (g *githubOAuthImp) Context() context.Context {
	return g.ctx
}

func (g *githubOAuthImp) GetAccessToken() *string {
	return g.accessToken
}

func (g *githubOAuthImp) SetRedirectURL(url string) {
	g.conf.RedirectURL = url
}

func (g *githubOAuthImp) GetAuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return g.conf.AuthCodeURL(state, opts...)
}

func (g *githubOAuthImp) GetNewClient(oauthToken string) (*github.Client, error) {
	token, err := g.conf.Exchange(g.ctx, oauthToken)
	if err != nil {
		return nil, err
	}
	g.accessToken = &token.AccessToken

	oauthClient := oauth2.NewClient(g.ctx, oauth2.StaticTokenSource(token))
	client := github.NewClient(oauthClient)

	return client, nil
}

func (g *githubOAuthImp) GetGitHubUser(client *github.Client) (*github.User, error) {
	ghUser, _, err := client.Users.Get(g.ctx, "")
	if err != nil {
		return nil, err
	}

	return ghUser, nil
}

// NewGitHubOAuth returns an instance of GitHubOAuth for loginHandler and oauthHandler.
func NewGitHubOAuth(ctx context.Context) (GitHubOAuth, error) {
	store := NewAppEngineDatastore(ctx, false)
	log := GetLogger(ctx)
	clientID, err := GetSecret(store, "github-oauth-client-id")
	if err != nil {
		log.Errorf("Failed to get github-oauth-client-id secret: %s", err.Error())
		return &githubOAuthImp{}, err
	}

	secret, err := GetSecret(store, "github-oauth-client-secret")
	if err != nil {
		log.Errorf("Failed to get github-oauth-client-secret: %s", err.Error())
		return &githubOAuthImp{}, err
	}

	oauth := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		// (no scope) - see https://developer.github.com/apps/building-oauth-apps/understanding-scopes-for-oauth-apps/#available-scopes
		Scopes:   []string{},
		Endpoint: ghOAuth.Endpoint,
	}

	return &githubOAuthImp{ctx: ctx, conf: oauth, ds: store}, nil
}
