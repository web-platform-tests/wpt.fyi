package webapp

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"net/http"
	"net/url"

	"github.com/google/go-github/github"
	"github.com/gorilla/securecookie"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/oauth2"
	ghOAuth "golang.org/x/oauth2/github"
	"google.golang.org/appengine"
)

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	if !aeAPI.IsFeatureEnabled("githubLogin") {
		http.Error(w, "Feature not implemented", http.StatusNotImplemented)
		return
	}

	user, token := getUserFromCookie(r)
	returnURL := r.FormValue("return")
	redirect := returnURL
	log := shared.GetLogger(ctx)
	if user == nil || token == nil {
		log.Infof("Initiating a new user login.")
		conf := getGithubOAuthConfig(ctx)
		conf.RedirectURL = getCallbackURI(returnURL, r)
		state, err := generateRandomState(32)
		if err != nil {
			log.Errorf("Failed to generate random state")
			http.Error(w, "Error creating a random state for login", http.StatusInternalServerError)
			return
		}
		redirect = conf.AuthCodeURL(state, oauth2.AccessTypeOnline)
		setState(ctx, state, w)
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
	oauthToken := r.FormValue("code")
	if oauthToken == "" {
		http.Error(w, "No token or username provided", http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	conf := getGithubOAuthConfig(ctx)
	token, err := conf.Exchange(ctx, oauthToken)
	log := shared.GetLogger(ctx)
	if err != nil {
		log.Errorf("Invalid OAuth2 token: %s", err.Error())
		http.Error(w, "Invalid OAuth2 token", http.StatusBadRequest)
		return
	}

	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client := github.NewClient(oauthClient)
	if err != nil {
		log.Errorf("Failed to get GitHub client")
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}
	ghUser, _, err := client.Users.Get(ctx, "") // Empty string => Authenticated user.
	if err != nil || ghUser == nil {
		log.Errorf("Failed to get authenticated user")
		http.Error(w, "Failed to get authenticated user", http.StatusBadRequest)
		return
	}
	user := &User{
		GitHubHandle: ghUser.GetLogin(),
		GithuhEmail:  ghUser.GetEmail(),
	}
	setSession(ctx, user, &token.AccessToken, w)
	log.Infof("User %s logged in", user.GitHubHandle)

	ret := r.FormValue("return")
	encodedState := r.FormValue("state")
	stateFromCookie := getState(r)
	if encodedState != stateFromCookie {
		log.Errorf("Failed to verify encoded state")
		http.Error(w, "Failed to verify encoded state", http.StatusBadRequest)
		return
	}

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
		if err = getSecureCookie(ctx).Decode("session", cookie.Value, &cookieValue); err == nil {
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

func setSession(ctx context.Context, user *User, token *string, response http.ResponseWriter) {
	value := map[string]interface{}{
		"user":  *user,
		"token": *token,
	}
	if encoded, err := getSecureCookie(ctx).Encode("session", value); err == nil {
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
}

func setState(ctx context.Context, state string, response http.ResponseWriter) {
	if encoded, err := getSecureCookie(ctx).Encode("state", state); err == nil {
		cookie := &http.Cookie{
			Name:     "state",
			Value:    encoded,
			Path:     "/",
			MaxAge:   600,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(response, cookie)
	} else {
		log := shared.GetLogger(ctx)
		log.Errorf("Failed to set state cookie: %s", err.Error())
	}
}

func getState(r *http.Request) string {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)
	cookieValue := ""
	if cookie, err := r.Cookie("state"); err == nil && cookie != nil {
		if err = getSecureCookie(ctx).Decode("state", cookie.Value, &cookieValue); err != nil {
			log.Errorf("Failed to Decode cookie for state: %s", err.Error())
		}
	}
	return cookieValue
}

var secureCookie *securecookie.SecureCookie

func getSecureCookie(ctx context.Context) *securecookie.SecureCookie {
	if secureCookie == nil {
		store := shared.NewAppEngineDatastore(ctx, false)
		hashKey, _ := shared.GetSecret(store, "secure-cookie-hashkey")
		blockKey, _ := shared.GetSecret(store, "secure-cookie-blockkey")
		secureCookie = securecookie.New([]byte(hashKey), []byte(blockKey))
	}
	return secureCookie
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

func getGithubOAuthConfig(ctx context.Context) *oauth2.Config {
	store := shared.NewAppEngineDatastore(ctx, false)
	clientID, _ := shared.GetSecret(store, "github-oauth-client-id")
	secret, _ := shared.GetSecret(store, "github-oauth-client-secret")
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		// (no scope) - see https://developer.github.com/apps/building-oauth-apps/understanding-scopes-for-oauth-apps/#available-scopes
		Scopes:   []string{},
		Endpoint: ghOAuth.Endpoint,
	}
}

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
