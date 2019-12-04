package webapp

import (
	"context"
	"encoding/base64"
	"encoding/gob"
	"net/http"
	"crypto/rand"

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
}

var returnURLMap = make(map[string]string)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	if !aeAPI.IsFeatureEnabled("githubLogin") {
		http.Error(w, "Feature not implemented", http.StatusNotImplemented)
	}

	user := getUserFromCookie(r)
	returnURL := r.FormValue("return")
	redirect := returnURL
	log := shared.GetLogger(ctx)
	if user == nil {
		conf := getGithubOAuthConfig(ctx)
		state, err := generateRandomState(32)
		if err != nil {
			log.Errorf("Failed to generate random state")
			http.Error(w, "Error creating a random state for login", http.StatusInternalServerError)
			return
		}
		redirect = conf.AuthCodeURL(state, oauth2.AccessTypeOnline)
		setState(ctx, state, w)
		returnURLMap[state] = returnURL
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
	}
	setSession(ctx, user, w)
	log.Infof("User %s logged in", user.GitHubHandle)

	ret := "/"
	encodedState := r.FormValue("state")
	stateFromCookie := getState(r)
	if encodedState != stateFromCookie {
		log.Errorf("Failed to verify encoded state")
		http.Error(w, "Failed to verify encoded state", http.StatusBadRequest)
		return
	}

	if val, ok := returnURLMap[encodedState]; ok {
		ret = val
	}
	http.Redirect(w, r, ret, http.StatusTemporaryRedirect)
}

func logoutHandler(response http.ResponseWriter, ruest *http.Request) {
	clearSession(response)
	http.Redirect(response, ruest, "/", http.StatusFound)
}

func getUserFromCookie(r *http.Request) (user *User) {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)
	if cookie, err := r.Cookie("session"); err == nil && cookie != nil {
		cookieValue := make(map[string]interface{})
		if err = getSecureCookie(ctx).Decode("session", cookie.Value, &cookieValue); err == nil {
			if decoded, ok := cookieValue["user"].(User); ok {
				user = &decoded
			} else if appengine.IsDevAppServer() {
				log.Errorf("Failed to cast user: %s", err.Error())
			}
		} else if appengine.IsDevAppServer() {
			log.Errorf("Failed to Decode cookie: %s", err.Error())
		}
	}
	return user
}

func setSession(ctx context.Context, user *User, response http.ResponseWriter) {
	value := map[string]interface{}{
		"user": *user,
	}
	if encoded, err := getSecureCookie(ctx).Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	} else {
		log := shared.GetLogger(ctx)
		log.Errorf("Failed to set session cookie: %s", err.Error())
	}
}

func setState(ctx context.Context, state string, response http.ResponseWriter) {
	if encoded, err := getSecureCookie(ctx).Encode("state", state); err == nil {
		cookie := &http.Cookie{
			Name:  "state",
			Value: encoded,
			Path:  "/",
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
			log.Errorf("Failed to Decode cookie: %s", err.Error())
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

func extractString(object map[string]interface{}, field string) string {
	if value, ok := object[field]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
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
