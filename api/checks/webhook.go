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
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/lukebjerring/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/oauth2"
)

// NOTE(lukebjerring): This is https://github.com/apps/staging-wpt-fyi-status-check
const wptfyiCheckAppID = 19965
const checksStagingAppID = 21580
const wptRepoID = 3618133
const wptRepoInstallationID = 449270
const wptRepoOwner = "web-platform-tests"
const wptRepoName = "wpt"

// checkWebhookHandler listens for check_suite and check_run events,
// responding to requested and rerequested events.
func checkWebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Errorf("Invalid content-type: %s", contentType)
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
		log.Debugf("Ignoring %s event", event)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
		aeAPI := shared.NewAppEngineAPI(ctx)
		suitesAPI := NewSuitesAPI(ctx)
		processed, err = handleCheckRunEvent(aeAPI, suitesAPI, payload)
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
		fmt.Fprintln(w, "wpt.fyi check(s) scheduled successfully")
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
	if appID != wptfyiCheckAppID && appID != checksStagingAppID {
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

		installationID := checkSuite.GetInstallation().GetID()
		if action == "requested" {
			// For new suites, check if the pull is across forks; if so, request a suite
			// on the main repo (web-platform-tests/wpt) too.
			pullRequests := checkSuite.GetCheckSuite().PullRequests
			for _, p := range pullRequests {
				destRepoID := p.GetBase().GetRepo().GetID()
				if destRepoID == wptRepoID && p.GetHead().GetRepo().GetID() != destRepoID {
					_, err := createWPTCheckSuite(ctx, appID, installationID, sha)
					if err != nil {
						log.Errorf("Failed to create wpt check_suite: %s", err.Error())
					}
				}
			}
		}

		suite, err := getOrCreateCheckSuite(ctx, sha, owner, repo, appID, installationID)
		if err != nil || suite == nil {
			return false, err
		}

		if action == "rerequested" {
			return scheduleProcessingForExistingRuns(ctx, sha)
		}
	}
	return false, nil
}

// handleCheckRunEvent handles a check_run rerequested events by updating
// the status based on whether results for the check_run's product exist.
func handleCheckRunEvent(aeAPI shared.AppEngineAPI, suitesAPI SuitesAPI, payload []byte) (bool, error) {
	log := shared.GetLogger(aeAPI.Context())
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

	shouldSchedule := false
	if (action == "created" && status != "completed") || action == "rerequested" {
		shouldSchedule = true
	} else if action == "requested_action" {
		actionID := checkRun.GetRequestedAction().Identifier
		if actionID == "recompute" {
			shouldSchedule = true
		} else if actionID == "ignore" {
			// TODO(lukebjerring): Created IgnoredRegression summary.
		}
	}
	if !shouldSchedule {
		log.Debugf("Ignoring %s action for %s check_run", action, status)
		return false, nil
	}

	name, sha := checkRun.GetCheckRun().GetName(), checkRun.GetCheckRun().GetHeadSHA()
	log.Debugf("GitHub check run %v (%s @ %s) was %s", checkRun.GetCheckRun().GetID(), name, sha, action)
	spec, err := shared.ParseProductSpec(checkRun.GetCheckRun().GetName())
	if err != nil {
		log.Errorf("Failed to parse \"%s\" as product spec", checkRun.GetCheckRun().GetName())
		return false, err
	}
	suitesAPI.ScheduleResultsProcessing(sha, spec)
	return true, nil
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
		return createWPTCheckSuite(ctx, wptfyiCheckAppID, wptRepoInstallationID, sha)
	}
	return false, nil
}

func scheduleProcessingForExistingRuns(ctx context.Context, sha string, products ...shared.ProductSpec) (bool, error) {
	// Jump straight to completed check_run for already-present runs for the SHA.
	products = shared.ProductSpecs(products).OrDefault()
	runsByProduct, err := shared.LoadTestRuns(ctx, products, nil, sha[:10], nil, nil, nil, nil)
	if err != nil {
		return false, fmt.Errorf("Failed to load test runs: %s", err.Error())
	}
	createdSome := false
	api := NewSuitesAPI(ctx)
	for _, rbp := range runsByProduct {
		if len(rbp.TestRuns) > 0 {
			err := api.ScheduleResultsProcessing(sha, rbp.Product)
			createdSome = createdSome || err == nil
			if err != nil {
				return createdSome, err
			}
		}
	}
	return createdSome, nil
}

// createWPTCheckSuite creates a check_suite on the main wpt repo for the given
// SHA. This is needed when a PR comes from a different fork of the repo.
func createWPTCheckSuite(ctx context.Context, appID, installationID int64, sha string) (bool, error) {
	log := shared.GetLogger(ctx)
	log.Debugf("Creating check_suite for web-platform-tests/wpt @ %s", sha)

	jwtClient, err := getJWTClient(ctx, appID, installationID)
	if err != nil {
		return false, err
	}
	client := github.NewClient(jwtClient)

	opts := github.CreateCheckSuiteOptions{
		HeadSHA: sha,
	}
	suite, _, err := client.Checks.CreateCheckSuite(ctx, wptRepoOwner, wptRepoName, opts)
	if err == nil && suite != nil {
		log.Infof("check_suite %v created", suite.GetID())
		getOrCreateCheckSuite(ctx, sha, wptRepoOwner, wptRepoName, appID, installationID)
	}
	return suite != nil, err
}

// createCheckRun submits an http POST to create the check run on GitHub, handling JWT auth for the app.
func createCheckRun(ctx context.Context, suite shared.CheckSuite, opts github.CreateCheckRunOptions) (bool, error) {
	log := shared.GetLogger(ctx)
	status := ""
	if opts.Status != nil {
		status = *opts.Status
	}
	log.Debugf("Creating %s %s check_run for %s/%s @ %s", status, opts.Name, suite.Owner, suite.Repo, suite.SHA)
	if suite.AppID == 0 {
		suite.AppID = wptfyiCheckAppID
	}
	jwtClient, err := getJWTClient(ctx, suite.AppID, suite.InstallationID)
	if err != nil {
		return false, err
	}
	client := github.NewClient(jwtClient)

	checkRun, resp, err := client.Checks.CreateCheckRun(ctx, suite.Owner, suite.Repo, opts)
	if err != nil {
		log.Warningf("Failed to create check_run: %s", resp.Status)
		return false, err
	} else if checkRun != nil {
		log.Infof("Created check_run %v", checkRun.GetID())
	}
	return true, nil
}

// NOTE(lukebjerring): oauth2/jwt has incorrect field-names, and doesn't allow
// passing in an http.Client (for GitHub's Authorization header flow), so we
// are forced to copy a little code here :(
func getJWTClient(ctx context.Context, appID, installation int64) (*http.Client, error) {
	ss, err := getSignedJWT(ctx, appID)
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
func getSignedJWT(ctx context.Context, appID int64) (string, error) {
	secret, err := shared.GetSecret(ctx, fmt.Sprintf("github-app-private-key-%v", appID))
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode([]byte(secret))
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

func getCheckTitle(product shared.ProductSpec) string {
	return fmt.Sprintf("wpt.fyi - %s results", product.DisplayName())
}
