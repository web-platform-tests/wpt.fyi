// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/oauth2"
	"google.golang.org/appengine/datastore"
)

// NOTE(lukebjerring): This is https://github.com/apps/staging-wpt-fyi-status-check
const wptfyiCheckAppID = 19965
const wptRepoID = 3618133
const wptRepoInstallationID = 449270
const wptRepoOwner = "web-platform-tests"
const wptRepoName = "wpt"

// checkWebhookHandler listens for check_suite and check_run events,
// responding to requested and rerequested events.
func checkWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	event := r.Header.Get("X-GitHub-Event")
	switch event {
	case "check_suite":
	case "check_run":
	case "pull_request":
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)

	secret, err := shared.GetSecret(ctx, "github-check-webhook-secret")
	if err != nil {
		http.Error(w, "Unable to verify request: secret not found", http.StatusInternalServerError)
		return
	}

	payload, err := github.ValidatePayload(r, []byte(secret))
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debugf("GitHub Delivery: %s", r.Header.Get("X-GitHub-Delivery"))

	var processed bool
	if event == "check_suite" {
		processed, err = handleCheckSuiteEvent(ctx, payload)
	} else if event == "check_run" {
		processed, err = handleCheckRunEvent(ctx, payload)
	} else if event == "pull_request" {
		processed, err = handlePullRequestEvent(ctx, payload)
	}
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if processed {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "wpt.fyi check started successfully")
	} else {
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintln(w, "Status was ignored")
	}
	return
}

// handleCheckSuiteEvent handles a check_suite (re)requested event by ensuring
// that a check_run exists for each product that contains results for the head SHA.
func handleCheckSuiteEvent(ctx context.Context, payload []byte) (bool, error) {
	log := shared.GetLogger(ctx)
	var checkSuite github.CheckSuiteEvent
	if err := json.Unmarshal(payload, &checkSuite); err != nil {
		return false, err
	}

	appID := checkSuite.GetCheckSuite().GetApp().GetID()
	if appID != wptfyiCheckAppID {
		log.Infof("Ignoring check_suite App ID %v", appID)
		return false, nil
	}

	if !shared.IsFeatureEnabled(ctx, "checksAllUsers") {
		whitelist := []string{
			"autofoolip",
			"chromium-wpt-export-bot",
			"foolip",
			"jgraham",
			"lukebjerring",
		}
		sender := ""
		if checkSuite.Sender != nil && checkSuite.Sender.Login != nil {
			sender = *checkSuite.Sender.Login
		}
		if !shared.StringSliceContains(whitelist, sender) {
			log.Infof("Sender %s not whitelisted for wpt.fyi checks", sender)
			return false, nil
		}
	}

	action := checkSuite.GetAction()
	if action == "requested" || action == "rerequested" {
		owner := checkSuite.GetRepo().GetOwner().GetLogin()
		repo := checkSuite.GetRepo().GetName()
		sha := checkSuite.GetCheckSuite().GetHeadSHA()
		log.Debugf("Check suite %s: %s/%s @ %s", action, owner, repo, sha[:7])

		if action == "requested" {
			// For new suites, check if the pull is across forks; if so, request a suite
			// on the main repo (web-platform-tests/wpt) too.
			pullRequests := checkSuite.GetCheckSuite().PullRequests
			for _, p := range pullRequests {
				destRepoID := p.GetBase().GetRepo().GetID()
				if destRepoID == wptRepoID && p.GetHead().GetRepo().GetID() != destRepoID {
					_, err := createWPTCheckSuite(ctx, sha)
					if err != nil {
						log.Errorf("Failed to create wpt check_suite: %s", err.Error())
					}
				}
			}
		}

		installation := *checkSuite.Installation.ID
		suite, err := getOrCreateCheckSuite(ctx, sha, owner, repo, installation)
		if err != nil || suite == nil {
			return false, err
		}

		if action == "rerequested" {
			completeChecksForExistingRuns(ctx, sha)
		}
	}
	return false, nil
}

// handleCheckRunEvent handles a check_run rerequested events by updating
// the status based on whether results for the check_run's product exist.
func handleCheckRunEvent(ctx context.Context, payload []byte) (bool, error) {
	log := shared.GetLogger(ctx)
	var checkRun github.CheckRunEvent
	if err := json.Unmarshal(payload, &checkRun); err != nil {
		return false, err
	}

	appID := checkRun.GetCheckRun().GetApp().GetID()
	if appID != wptfyiCheckAppID {
		log.Infof("Ignoring check_suite App ID %v", appID)
		return false, nil
	}

	action := checkRun.GetAction()
	status := checkRun.GetCheckRun().GetStatus()
	if (action == "created" && status != "completed") || action == "rerequested" {
		name, sha := *checkRun.CheckRun.Name, *checkRun.CheckRun.HeadSHA
		log.Debugf("Check run %s @ %s %s", name, sha, action)
		spec, err := shared.ParseProductSpec(*checkRun.CheckRun.Name)
		if err != nil {
			log.Errorf("Failed to parse \"%s\" as product spec", *checkRun.CheckRun.Name)
		}
		return completeChecksForExistingRuns(ctx, sha, spec)
	}
	return false, nil
}

func handlePullRequestEvent(ctx context.Context, payload []byte) (bool, error) {
	var pullRequest github.PullRequestEvent
	if err := json.Unmarshal(payload, &pullRequest); err != nil {
		return false, err
	}

	switch pullRequest.GetAction() {
	case "opened":
	case "synchronize":
		break
	default:
		return false, nil
	}

	sha := pullRequest.GetPullRequest().GetHead().GetSHA()
	destRepoID := pullRequest.GetPullRequest().GetBase().GetRepo().GetID()
	if destRepoID == wptRepoID && pullRequest.GetPullRequest().GetHead().GetRepo().GetID() != destRepoID {
		// Pull is across forks; request a check suite on the main fork too.
		return createWPTCheckSuite(ctx, sha)
	}
	return false, nil
}

func completeChecksForExistingRuns(ctx context.Context, sha string, products ...shared.ProductSpec) (bool, error) {
	// Jump straight to completed check_run for already-present runs for the SHA.
	products = shared.ProductSpecs(products).OrDefault()
	runsByProduct, err := shared.LoadTestRuns(ctx, products, nil, sha[:10], nil, nil, nil, nil)
	if err != nil {
		return false, fmt.Errorf("Failed to load test runs: %s", err.Error())
	}
	createdSome := false
	for _, rbp := range runsByProduct {
		if len(rbp.TestRuns) > 0 {
			created, err := completeCheckRun(ctx, sha, rbp.Product)
			createdSome = createdSome || created
			if err != nil {
				return createdSome, err
			}
		}
	}
	return createdSome, nil
}

// createWPTCheckSuite creates a check_suite on the main wpt repo for the given
// SHA. This is needed when a PR comes from a different fork of the repo.
func createWPTCheckSuite(ctx context.Context, sha string) (bool, error) {
	log := shared.GetLogger(ctx)
	log.Debugf("Creating check_suite for web-platform-tests/wpt @ %s", sha)

	jwtClient, err := getJWTClient(ctx, wptRepoInstallationID)
	if err != nil {
		return false, err
	}
	client := github.NewClient(jwtClient)

	opts := github.CreateCheckSuiteOptions{
		HeadSHA: sha,
	}
	suite, _, err := client.Checks.CreateCheckSuite(ctx, wptRepoOwner, wptRepoName, opts)
	if suite != nil && err != nil {
		getOrCreateCheckSuite(ctx, sha, wptRepoOwner, wptRepoName, wptRepoInstallationID)
	}
	return suite != nil, err
}

func createCheckRun(ctx context.Context, suite shared.CheckSuite, opts github.CreateCheckRunOptions) (bool, error) {
	log := shared.GetLogger(ctx)
	status := ""
	if opts.Status != nil {
		status = *opts.Status
	}
	log.Debugf("Creating %s %s check_run for %s/%s @ %s", status, opts.Name, suite.Owner, suite.Repo, suite.SHA)
	jwtClient, err := getJWTClient(ctx, suite.InstallationID)
	if err != nil {
		return false, err
	}
	client := github.NewClient(jwtClient)
	u := fmt.Sprintf("repos/%v/%v/check-runs", suite.Owner, suite.Repo)
	req, err := client.NewRequest("POST", u, opts)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github.antiope-preview+json")
	checkRun := new(github.CheckRun)
	resp, err := client.Do(ctx, req, checkRun)

	if err != nil {
		log.Warningf("Failed to create check_run: %s", resp.Status)
		body, _ := ioutil.ReadAll(resp.Body)
		log.Warningf(string(body))
		return false, err
	} else if checkRun != nil {
		log.Infof("Created check_run %v", checkRun.GetID())
	}
	return true, nil
}

// NOTE(lukebjerring): oauth2/jwt has incorrect field-names, and doesn't allow
// passing in an http.Client (for GitHub's Authorization header flow), so we
// are forced to copy a little code here :(
func getJWTClient(ctx context.Context, installation int64) (*http.Client, error) {
	ss, err := getSignedJWT(ctx)
	if err != nil {
		return nil, err
	}
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ss},
	)
	oauthClient := oauth2.NewClient(ctx, tokenSource)

	tokenURL := fmt.Sprintf("https://api.github.com/app/installations/%v/access_tokens", installation)
	req, _ := http.NewRequest("POST", tokenURL, nil)
	req.Header.Set("Accept", "application/vnd.github.machine-man-preview+json")
	resp, err := oauthClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch installation token: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("cannot fetch installation token: %v", err)
	}
	if c := resp.StatusCode; c < 200 || c > 299 {
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
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}
	token := &oauth2.Token{
		AccessToken: tokenResponse.AccessToken,
		Expiry:      tokenResponse.ExpiresAt,
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(token)), nil
}

// https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/#authenticating-as-a-github-app
func getSignedJWT(ctx context.Context) (string, error) {
	// Fetch shared.Token entity for GitHub API Token.
	tokenKey := datastore.NewKey(ctx, "Token", "github-app-private-key", 0, nil)
	var token shared.Token
	datastore.Get(ctx, tokenKey, &token)
	block, _ := pem.Decode([]byte(token.Secret))
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	/* Create the jwt token */
	now := time.Now()
	claims := &jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(time.Minute * 10).Unix(),
		Issuer:    strconv.Itoa(wptfyiCheckAppID),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return jwtToken.SignedString(key)
}

func getCheckTitle(product shared.ProductSpec) string {
	return fmt.Sprintf("wpt.fyi - %s results", product.DisplayName())
}
