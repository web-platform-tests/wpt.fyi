// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -build_flags=--mod=mod -destination sharedtest/github_oauth_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared GitHubOAuth,GitHubAccessControl

package shared

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v80/github"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
	ghOAuth "golang.org/x/oauth2/github"
)

func init() {
	// All custom types stored in securecookie need to be registered.
	gob.RegisterName("User", User{})
}

// User represents an authenticated GitHub user.
type User struct {
	GitHubHandle string `json:"github_handle,omitempty"`
	GitHubEmail  string `json:"github_email,omitempty"`
}

// GitHubAccessControl encapsulates implementation details of access control for the wpt-metadata repository.
type GitHubAccessControl interface {
	// IsValid* functions also verify the access token with GitHub.
	IsValidWPTMember() (bool, error)
	IsValidAdmin() (bool, error)
}

type githubAccessControlImpl struct {
	ctx   context.Context
	ds    Datastore
	user  *User
	token string

	// This is the client for the OAuth app.
	oauthClientID string
	oauthGHClient *github.Client

	// This is the bot account client.
	botClient *github.Client
}

// GitHubOAuth encapsulates implementation details of GitHub OAuth flow.
type GitHubOAuth interface {
	Context() context.Context
	Datastore() Datastore
	GetAccessToken() string
	GetAuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	GetUser(client *github.Client) (*github.User, error)
	NewClient(oauthCode string) (*github.Client, error)
	SetRedirectURL(url string)
}

type githubOAuthImpl struct {
	ctx         context.Context
	ds          Datastore
	conf        *oauth2.Config
	accessToken string
}

func (g *githubOAuthImpl) Datastore() Datastore {
	return g.ds
}

func (g *githubOAuthImpl) Context() context.Context {
	return g.ctx
}

func (g *githubOAuthImpl) GetAccessToken() string {
	return g.accessToken
}

func (g *githubOAuthImpl) SetRedirectURL(url string) {
	g.conf.RedirectURL = url
}

func (g *githubOAuthImpl) GetAuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return g.conf.AuthCodeURL(state, opts...)
}

func (g *githubOAuthImpl) NewClient(oauthCode string) (*github.Client, error) {
	token, err := g.conf.Exchange(g.ctx, oauthCode)
	if err != nil {
		return nil, err
	}
	g.accessToken = token.AccessToken

	oauthClient := oauth2.NewClient(g.ctx, oauth2.StaticTokenSource(token))
	client := github.NewClient(oauthClient)

	return client, nil
}

func (g *githubOAuthImpl) GetUser(client *github.Client) (*github.User, error) {
	// Passing the empty string will fetch the authenticated user.
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

	clientID, secret, err := getOAuthClientIDSecret(store)
	if err != nil {
		log.Errorf("Failed to get github-oauth-client-{id,secret}: %s", err.Error())
		return nil, err
	}

	oauth := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		// (no scope) - see https://developer.github.com/apps/building-oauth-apps/understanding-scopes-for-oauth-apps/#available-scopes
		Scopes:   []string{},
		Endpoint: ghOAuth.Endpoint,
	}

	return &githubOAuthImpl{ctx: ctx, conf: oauth, ds: store}, nil
}

func (gaci githubAccessControlImpl) isValidAccessToken() (bool, error) {
	_, res, err := gaci.oauthGHClient.Authorizations.Check(gaci.ctx, gaci.oauthClientID, gaci.token)
	if err != nil {
		return false, err
	}

	return res.StatusCode == http.StatusOK, nil
}

func (gaci githubAccessControlImpl) IsValidWPTMember() (bool, error) {
	valid, err := gaci.isValidAccessToken()
	if err != nil {
		return false, err
	}
	if !valid {
		return false, errors.New("invalid access token")
	}
	isMember, _, err := gaci.botClient.Organizations.IsMember(gaci.ctx, "web-platform-tests", gaci.user.GitHubHandle)
	return isMember, err
}

func (gaci githubAccessControlImpl) IsValidAdmin() (bool, error) {
	valid, err := gaci.isValidAccessToken()
	if err != nil {
		return false, err
	}
	if !valid {
		return false, errors.New("invalid access token")
	}
	key := gaci.ds.NewNameKey("Admin", gaci.user.GitHubHandle)
	var dst struct{}
	if err := gaci.ds.Get(key, &dst); err == ErrNoSuchEntity {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// NewGitHubAccessControl returns a GitHubAccessControl for checking the
// permission of a logged-in GitHub user.
func NewGitHubAccessControl(ctx context.Context, ds Datastore, botClient *github.Client, user *User, token string) (GitHubAccessControl, error) {
	clientID, secret, err := getOAuthClientIDSecret(ds)
	if err != nil {
		return nil, err
	}
	tp := github.BasicAuthTransport{
		Username: clientID,
		Password: secret,
	}
	return githubAccessControlImpl{
		ctx:           ctx,
		ds:            ds,
		user:          user,
		token:         token,
		oauthClientID: clientID,
		oauthGHClient: github.NewClient(tp.Client()),
		botClient:     botClient,
	}, nil
}

// NewGitHubAccessControlFromRequest returns a GitHubAccessControl for checking
// the permission of a logged-in GitHub user from a request. (nil, nil) will be
// returned if the user is not logged in.
func NewGitHubAccessControlFromRequest(aeAPI AppEngineAPI, ds Datastore, r *http.Request) (GitHubAccessControl, error) {
	ctx := aeAPI.Context()
	botClient, err := aeAPI.GetGitHubClient()
	if err != nil {
		return nil, err
	}
	user, token := GetUserFromCookie(ctx, ds, r)
	if user == nil {
		return nil, nil
	}
	return NewGitHubAccessControl(ctx, ds, botClient, user, token)
}

// NewSecureCookie returns a SecureCookie instance for wpt.fyi. This instance
// can be used to encode and decode cookies set by wpt.fyi.
func NewSecureCookie(store Datastore) (*securecookie.SecureCookie, error) {
	hashKey, err := GetSecret(store, "secure-cookie-hashkey")
	if err != nil {
		return nil, fmt.Errorf("failed to get secure-cookie-hashkey secret: %v", err)
	}

	blockKey, err := GetSecret(store, "secure-cookie-blockkey")
	if err != nil {
		return nil, fmt.Errorf("failed to get secure-cookie-blockkey secret: %v", err)
	}

	secureCookie := securecookie.New([]byte(hashKey), []byte(blockKey))
	return secureCookie, nil
}

// GetUserFromCookie extracts the User and GitHub OAuth token from a request's
// session cookie, if it exists. If the cookie does not exist or cannot be
// decoded, (nil, "") will be returned.
func GetUserFromCookie(ctx context.Context, ds Datastore, r *http.Request) (*User, string) {
	log := GetLogger(ctx)
	if cookie, err := r.Cookie("session"); err == nil && cookie != nil {
		cookieValue := make(map[string]interface{})
		sc, err := NewSecureCookie(ds)
		if err != nil {
			log.Errorf("Failed to create SecureCookie: %s", err.Error())
			return nil, ""
		}

		if err = sc.Decode("session", cookie.Value, &cookieValue); err == nil {
			decodedUser, okUser := cookieValue["user"].(User)
			decodedToken, okToken := cookieValue["token"].(string)
			if okUser && okToken {
				return &decodedUser, decodedToken
			}
			log.Errorf("Failed to cast user or token")
		} else {
			log.Errorf("Failed to decode cookie: %s", err.Error())
		}
	}
	return nil, ""
}

// NewGitHubClientFromToken returns a new GitHub client from an access token.
func NewGitHubClientFromToken(ctx context.Context, token string) *github.Client {
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	}))
	return github.NewClient(oauthClient)
}

func getOAuthClientIDSecret(store Datastore) (clientID, clientSecret string, err error) {
	clientID, err = GetSecret(store, "github-oauth-client-id")
	if err != nil {
		return "", "", err
	}
	clientSecret, err = GetSecret(store, "github-oauth-client-secret")
	if err != nil {
		return "", "", err
	}
	return clientID, clientSecret, nil
}
