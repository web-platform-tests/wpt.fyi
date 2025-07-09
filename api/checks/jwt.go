// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt"
	"github.com/google/go-github/v73/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/oauth2"
)

func getGitHubClient(ctx context.Context, appID, installationID int64) (*github.Client, error) {
	jwtClient, err := getJWTClient(ctx, appID, installationID)
	if err != nil {
		return nil, err
	}

	return github.NewClient(jwtClient), nil
}

// NOTE(lukebjerring): oauth2/jwt has incorrect field-names, and doesn't allow
// passing in an http.Client (for GitHub's Authorization header flow), so we
// are forced to copy a little code here :(.
func getJWTClient(ctx context.Context, appID, installation int64) (*http.Client, error) {
	ss, err := getSignedJWT(ctx, appID)
	if err != nil {
		return nil, err
	}
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ss}, // nolint:exhaustruct // WONTFIX: AccessToken only required.
	)
	oauthClient := oauth2.NewClient(ctx, tokenSource)

	tokenURL := fmt.Sprintf("https://api.github.com/app/installations/%v/access_tokens", installation)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, tokenURL, nil)
	req.Header.Set("Accept", "application/vnd.github.machine-man-preview+json")
	resp, err := oauthClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch installation token: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("cannot fetch installation token: %w", err)
	}
	if c := resp.StatusCode; c < 200 || c > 299 {
		// nolint:exhaustruct // TODO: Fix exhaustruct lint error.
		// Investigate which error code should be returned here.
		return nil, &oauth2.RetrieveError{
			Response: resp,
			Body:     body,
		}
	}
	// tokenResponse is the JSON response body.
	var tokenResponse struct {
		AccessToken string    `json:"token"`
		ExpiresAt   time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %w", err)
	}
	// nolint:exhaustruct // WONTFIX: AccessToken only required.
	token := &oauth2.Token{
		AccessToken: tokenResponse.AccessToken,
		Expiry:      tokenResponse.ExpiresAt,
	}

	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(token)), nil
}

// nolint:lll // Keep hyperlink
// https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/#authenticating-as-a-github-app
func getSignedJWT(ctx context.Context, appID int64) (string, error) {
	ds := shared.NewAppEngineDatastore(ctx, false)
	secret, err := shared.GetSecret(ds, fmt.Sprintf("github-app-private-key-%v", appID))
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode([]byte(secret))
	if block == nil {
		return "", errors.New("failed to decode private key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	/* Create the jwt token */
	now := time.Now()
	claims := &jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(time.Minute * 10).Unix(),
		Issuer:    fmt.Sprintf("%v", appID),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return jwtToken.SignedString(key)
}
