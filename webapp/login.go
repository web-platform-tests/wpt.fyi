// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"net/http"
	"net/url"

	"github.com/google/go-github/v28/github"
	"github.com/gorilla/securecookie"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/oauth2"
	ghOAuth "golang.org/x/oauth2/github"
	"google.golang.org/appengine"
)

func init() {
	gob.Register(map[string]interface{}{})
	// Need RegisterName - for local packages, Register appends main[0-9]{5}.User
	gob.RegisterName("User", User{})
}

// User represents an authenticated GitHub user.
type User struct {
	GitHubHandle string
	GithuhEmail  string
}

type GithubOAuth interface {
	GetAccessToken() *string
	SetRedirectURL(url string)
	GetAuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	GetNewClient(oauthToken string) (*github.Client, error)
	GetGithubUser(client *github.Client) (*github.User, error)
}

type GithubOAuthImp struct {
	ctx         context.Context
	conf        *oauth2.Config
	accessToken *string
}

func (g GithubOAuthImp) GetAccessToken() *string {
	return g.accessToken
}

func (g GithubOAuthImp) SetRedirectURL(url string) {
	g.conf.RedirectURL = url
}

func (g GithubOAuthImp) GetAuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return g.conf.AuthCodeURL(state, opts...)
}

func (g GithubOAuthImp) GetNewClient(oauthToken string) (*github.Client, error) {
	token, err := g.conf.Exchange(g.ctx, oauthToken)
	if err != nil {
		return nil, err
	}
	g.accessToken = &token.AccessToken

	oauthClient := oauth2.NewClient(g.ctx, oauth2.StaticTokenSource(token))
	client := github.NewClient(oauthClient)

	return client, nil
}

func (g GithubOAuthImp) GetGithubUser(client *github.Client) (*github.User, error) {
	ghUser, _, err := client.Users.Get(g.ctx, "")
	if err != nil {
		return nil, err
	}

	return ghUser, nil
}

func newGithubOAuth(ctx context.Context) (GithubOAuth, error) {
	store := shared.NewAppEngineDatastore(ctx, false)
	log := shared.GetLogger(ctx)
	clientID, err := shared.GetSecret(store, "github-oauth-client-id")
	if err != nil {
		log.Errorf("Failed to get github-oauth-client-id secret: %s", err.Error())
		return GithubOAuthImp{}, err
	}

	secret, err := shared.GetSecret(store, "github-oauth-client-secret")
	if err != nil {
		log.Errorf("Failed to get github-oauth-client-secret: %s", err.Error())
		return GithubOAuthImp{}, err
	}

	oauth := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		// (no scope) - see https://developer.github.com/apps/building-oauth-apps/understanding-scopes-for-oauth-apps/#available-scopes
		Scopes:   []string{},
		Endpoint: ghOAuth.Endpoint,
	}

	return GithubOAuthImp{ctx: ctx, conf: oauth}, nil
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	if !aeAPI.IsFeatureEnabled("githubLogin") {
		http.Error(w, "Feature not enabled", http.StatusNotImplemented)
		return
	}

	githubOuathImp, err := newGithubOAuth(ctx)
	if err != nil {
		http.Error(w, "Error creating githubOuathImp", http.StatusInternalServerError)
		return
	}
	handleLogin(githubOuathImp, w, r)
}

func handleLogin(g GithubOAuth, w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	user, token := getUserFromCookie(r)
	returnURL := r.FormValue("return")
	if returnURL == "" {
		returnURL = "/"
	}

	redirect := ""
	log := shared.GetLogger(ctx)
	if user == nil || token == nil {
		log.Infof("Initiating a new user login.")
		g.SetRedirectURL(getCallbackURI(returnURL, r))
		state, err := generateRandomState(32)
		if err != nil {
			log.Errorf("Failed to generate random state: %v", err)
			http.Error(w, "Error creating a random state for login", http.StatusInternalServerError)
			return
		}

		redirect = g.GetAuthCodeURL(state, oauth2.AccessTypeOnline)
		err = setState(ctx, state, w)
		if err != nil {
			log.Errorf("Failed to set state cookie: %s", err.Error())
			http.Error(w, "Error setting state cookie for login", http.StatusInternalServerError)
		}

		log.Infof("OAuthing with github and returning to %s", returnURL)
	} else {
		if redirect == "" {
			redirect = "/"
		}
		log.Infof("User %s is logged in", user.GitHubHandle)
	}

	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

func oauthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	githubOuathImp, err := newGithubOAuth(ctx)
	if err != nil {
		http.Error(w, "Error creating githubOuathImp", http.StatusInternalServerError)
		return
	}
	handleOauth(githubOuathImp, w, r)
}

func handleOauth(g GithubOAuth, w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)

	encodedState := r.FormValue("state")
	if encodedState == "" {
		log.Errorf("Failed to get state URL param")
		http.Error(w, "Failed to get state URL param", http.StatusBadRequest)
		return
	}

	stateFromCookie := getState(r)
	if stateFromCookie == "" {
		http.Error(w, "Failed to get state cookie", http.StatusBadRequest)
		return
	}

	if encodedState != stateFromCookie {
		log.Errorf("Failed to verify encoded state")
		http.Error(w, "Failed to verify encoded state", http.StatusBadRequest)
		return
	}

	oauthToken := r.FormValue("code")
	if oauthToken == "" {
		http.Error(w, "No token or username provided", http.StatusBadRequest)
		return
	}

	client, err := g.GetNewClient(oauthToken)
	if err != nil {
		log.Errorf("Error creating Github client using OAuth2 token: %s", err.Error())
		http.Error(w, "Error creating Github client using OAuth2 token", http.StatusBadRequest)
		return
	}

	// Passing the empty string will fetch the authenticated user.
	ghUser, err := g.GetGithubUser(client)
	if err != nil || ghUser == nil {
		log.Errorf("Failed to get authenticated user: %s", err.Error())
		http.Error(w, "Failed to get authenticated user", http.StatusBadRequest)
		return
	}

	user := &User{
		GitHubHandle: ghUser.GetLogin(),
		GithuhEmail:  ghUser.GetEmail(),
	}
	setSession(ctx, user, g.GetAccessToken(), w)
	if err != nil {
		http.Error(w, "Failed to set credential cookie", http.StatusInternalServerError)
		return
	}
	log.Infof("User %s logged in", user.GitHubHandle)

	ret := r.FormValue("return")
	http.Redirect(w, r, ret, http.StatusTemporaryRedirect)
}

func logoutHandler(response http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)
	clearSession(response)

	log.Infof("User logged out")
	http.Redirect(response, r, "/", http.StatusFound)
}

func getUserFromCookie(r *http.Request) (*User, *string) {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)
	if cookie, err := r.Cookie("session"); err == nil && cookie != nil {
		cookieValue := make(map[string]interface{})
		sc, err := getSecureCookie(ctx)
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

func setSession(ctx context.Context, user *User, token *string, response http.ResponseWriter) error {
	var err error
	value := map[string]interface{}{
		"user":  *user,
		"token": *token,
	}

	sc, err := getSecureCookie(ctx)
	if err != nil {
		return err
	}

	if encoded, err := sc.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:     "session",
			Value:    encoded,
			Path:     "/",
			MaxAge:   2592000,
			HttpOnly: true,
			Secure:   true,
		}

		// SameSite=None for http.Cookie is only available in Go.113;
		// see https://github.com/golang/go/issues/32546.
		if v := cookie.String(); v != "" {
			response.Header().Add("Set-Cookie", v+"; SameSite=None")
		}
	} else {
		log := shared.GetLogger(ctx)
		log.Errorf("Failed to set session cookie: %s", err.Error())
	}

	return err
}

func setState(ctx context.Context, state string, response http.ResponseWriter) error {
	var err error
	sc, err := getSecureCookie(ctx)
	if err != nil {
		return err
	}

	if encoded, err := sc.Encode("state", state); err == nil {
		cookie := &http.Cookie{
			Name:     "state",
			Value:    encoded,
			Path:     "/",
			MaxAge:   600,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(response, cookie)
	}

	return err
}

func getState(r *http.Request) string {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)
	cookieValue := ""
	if cookie, err := r.Cookie("state"); err == nil && cookie != nil {
		sc, err := getSecureCookie(ctx)
		if err != nil {
			return ""
		}

		if err = sc.Decode("state", cookie.Value, &cookieValue); err != nil {
			log.Errorf("Failed to decode cookie for state: %s", err.Error())
		}
	} else {
		log.Errorf("Failed to get state cookie: %s", err.Error())
	}
	return cookieValue
}

var secureCookie *securecookie.SecureCookie

func getSecureCookie(ctx context.Context) (*securecookie.SecureCookie, error) {
	log := shared.GetLogger(ctx)
	if secureCookie == nil {
		store := shared.NewAppEngineDatastore(ctx, false)
		hashKey, err := shared.GetSecret(store, "secure-cookie-hashkey")
		if err != nil {
			log.Errorf("Failed to get secure-cookie-hashkey secret: %s", err.Error())
			return nil, err
		}

		blockKey, err := shared.GetSecret(store, "secure-cookie-blockkey")
		if err != nil {
			log.Errorf("Failed to get secure-cookie-blockkey secret: %s", err.Error())
			return nil, err
		}

		secureCookie = securecookie.New([]byte(hashKey), []byte(blockKey))
	}
	return secureCookie, nil
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

/*
func getGithubOAuthConfig(ctx context.Context) *oauth2.Config {
	store := shared.NewAppEngineDatastore(ctx, false)
	log := shared.GetLogger(ctx)
	clientID, err := shared.GetSecret(store, "github-oauth-client-id")
	if err != nil {
		log.Errorf("Failed to get github-oauth-client-id secret: %s", err.Error())
	}

	secret, err := shared.GetSecret(store, "github-oauth-client-secret")
	if err != nil {
		log.Errorf("Failed to get github-oauth-client-secret: %s", err.Error())
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		// (no scope) - see https://developer.github.com/apps/building-oauth-apps/understanding-scopes-for-oauth-apps/#available-scopes
		Scopes:   []string{},
		Endpoint: ghOAuth.Endpoint,
	}
}
*/

func generateRandomState(size int) (string, error) {
	byteArray := make([]byte, size)
	_, err := rand.Read(byteArray)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(byteArray), nil
}

func getCallbackURI(ret string, r *http.Request) string {
	callback := url.URL{Scheme: "https", Host: r.Host, Path: "oauth"}
	q := callback.Query()
	q.Set("return", ret)
	callback.RawQuery = q.Encode()
	return callback.String()
}
