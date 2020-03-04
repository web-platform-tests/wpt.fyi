// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/oauth2"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	if !aeAPI.IsFeatureEnabled("githubLogin") {
		http.Error(w, "Feature not enabled", http.StatusNotImplemented)
		return
	}

	githubOauthImp, err := shared.NewGitHubOAuth(ctx)
	if err != nil {
		http.Error(w, "Error creating githuboauthImp", http.StatusInternalServerError)
		return
	}
	handleLogin(githubOauthImp, w, r)
}

func handleLogin(g shared.GitHubOAuth, w http.ResponseWriter, r *http.Request) {
	ctx := g.Context()
	ds := g.Datastore()
	user, token := shared.GetUserFromCookie(ctx, ds, r)
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
			log.Errorf("Failed to generate a random state for OAuth: %v", err)
			http.Error(w, "Failed to generate a random state for OAuth", http.StatusInternalServerError)
			return
		}

		redirect = g.GetAuthCodeURL(state, oauth2.AccessTypeOnline)
		err = setState(ctx, ds, state, w)
		if err != nil {
			log.Errorf("Failed to set state cookie for OAuth: %v", err)
			http.Error(w, "Failed to set state cookie for OAuth", http.StatusInternalServerError)
			return
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
	githuboauthImp, err := shared.NewGitHubOAuth(ctx)
	if err != nil {
		http.Error(w, "Error creating githuboauthImp", http.StatusInternalServerError)
		return
	}
	handleOauth(githuboauthImp, w, r)
}

func handleOauth(g shared.GitHubOAuth, w http.ResponseWriter, r *http.Request) {
	ctx := g.Context()
	log := shared.GetLogger(ctx)
	ds := g.Datastore()

	encodedState := r.FormValue("state")
	if encodedState == "" {
		http.Error(w, "Missing URL param \"state\"", http.StatusBadRequest)
		return
	}

	encryptedState, err := r.Cookie("state")
	if err != nil || encryptedState == nil {
		http.Error(w, "Missing cookie \"state\"", http.StatusBadRequest)
		return
	}
	stateFromCookie, err := decodeState(ctx, ds, encryptedState)
	if err != nil {
		log.Errorf(err.Error())
		http.Error(w, "Failed to decode state from cookies", http.StatusBadRequest)
		return
	}

	if stateFromCookie == "" {
		http.Error(w, "Failed to get state cookie", http.StatusBadRequest)
		return
	}

	if encodedState != stateFromCookie {
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
		log.Errorf("Error creating GitHub client using OAuth2 token: %v", err)
		http.Error(w, "Error creating GitHub client using OAuth2 token", http.StatusBadRequest)
		return
	}

	// Passing the empty string will fetch the authenticated user.
	ghUser, err := g.GetGitHubUser(client)
	if err != nil || ghUser == nil {
		log.Errorf("Failed to get authenticated user: %v", err)
		http.Error(w, "Failed to get authenticated user", http.StatusBadRequest)
		return
	}

	user := &shared.User{
		GitHubHandle: ghUser.GetLogin(),
		GithuhEmail:  ghUser.GetEmail(),
	}
	setSession(ctx, ds, user, g.GetAccessToken(), w)
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

func setSession(ctx context.Context, ds shared.Datastore, user *shared.User, token *string, response http.ResponseWriter) error {
	var err error
	value := map[string]interface{}{
		"user":  *user,
		"token": *token,
	}

	sc, err := shared.GetSecureCookie(ctx, ds)
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
		log.Errorf("Failed to set session cookie: %v", err)
	}

	return err
}

func setState(ctx context.Context, ds shared.Datastore, state string, response http.ResponseWriter) error {
	var err error
	sc, err := shared.GetSecureCookie(ctx, ds)
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

func decodeState(ctx context.Context, ds shared.Datastore, encryptedState *http.Cookie) (string, error) {
	cookieValue := ""
	sc, err := shared.GetSecureCookie(ctx, ds)
	if err != nil {
		return "", fmt.Errorf("failed to create securecookie decoder: %v", err)
	}

	if err := sc.Decode("state", encryptedState.Value, &cookieValue); err != nil {
		return "", fmt.Errorf("failed to decode state cookie: %v", err)
	}
	return cookieValue, nil
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
