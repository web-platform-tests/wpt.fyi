// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/github_oauth_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared GitHubOAuth

package shared

import (
	"context"
	"encoding/gob"
	"net/http"

	"github.com/google/go-github/v28/github"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
	ghOAuth "golang.org/x/oauth2/github"
	"google.golang.org/appengine"
)

func init() {
	// All custom types stored in securecookie need to be registered.
	gob.RegisterName("User", User{})
}

// User represents an authenticated GitHub user.
type User struct {
	GitHubHandle string
	GithuhEmail  string
}

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

// GetSecureCookie returns the securecookie instance for wpt.fyi. This instance can
// be used to encode and decode cookies set by wpt.fyi.
func GetSecureCookie(ctx context.Context, store Datastore) (*securecookie.SecureCookie, error) {
	log := GetLogger(ctx)
	hashKey, err := GetSecret(store, "secure-cookie-hashkey")
	if err != nil {
		log.Errorf("Failed to get secure-cookie-hashkey secret: %s", err.Error())
		return nil, err
	}

	blockKey, err := GetSecret(store, "secure-cookie-blockkey")
	if err != nil {
		log.Errorf("Failed to get secure-cookie-blockkey secret: %s", err.Error())
		return nil, err
	}

	secureCookie := securecookie.New([]byte(hashKey), []byte(blockKey))
	return secureCookie, nil
}

// GetUserFromCookie extracts the User and GitHub OAuth token from a request's
// session cookie, if it exists. If the cookie does not exist or cannot be decoded, nil
// is returned for both.
func GetUserFromCookie(ctx context.Context, ds Datastore, r *http.Request) (*User, *string) {
	log := GetLogger(ctx)
	if cookie, err := r.Cookie("session"); err == nil && cookie != nil {
		cookieValue := make(map[string]interface{})
		sc, err := GetSecureCookie(ctx, ds)
		if err != nil {
			return nil, nil
		}

		if err = sc.Decode("session", cookie.Value, &cookieValue); err == nil {
			decodedUser, okUser := cookieValue["user"].(User)
			decodedToken, okToken := cookieValue["token"].(string)
			if okUser && okToken {
				return &decodedUser, &decodedToken
			} else if appengine.IsDevAppServer() {
				log.Errorf("Failed to cast user or token")
			}
		} else if appengine.IsDevAppServer() {
			log.Errorf("Failed to Decode cookie: %s", err.Error())
		}
	}
	return nil, nil
}
