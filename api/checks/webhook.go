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
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/deckarep/golang-set"
	"github.com/google/go-github/github"
	wptgithub "github.com/web-platform-tests/wpt.fyi/api/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/oauth2"
	"google.golang.org/appengine/datastore"
)

// checkWebhookHandler listens for check_suite and check_run events,
// responding to requested and rerequested events.
func checkWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" ||
		(r.Header.Get("X-GitHub-Event") != "check_suite" &&
			r.Header.Get("X-GitHub-Event") != "check_run") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)

	payload, err := wptgithub.VerifyAndGetPayload(r, "github-check-webhook-secret")
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debugf("GitHub Delivery: %s", r.Header.Get("X-GitHub-Delivery"))

	var processed bool
	if r.Header.Get("X-GitHub-Event") == "check_suite" {
		processed, err = handleCheckSuiteEvent(ctx, payload)
	} else {
		processed, err = handleCheckRunEvent(ctx, payload)
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
	if checkSuite.Action != nil &&
		(*checkSuite.Action == "requested" || *checkSuite.Action == "rerequested") {
		log.Debugf("Check suite %s: %s", *(checkSuite.Action), *(checkSuite.CheckSuite.HeadBranch))

		sha := *checkSuite.CheckSuite.HeadSHA
		owner := *checkSuite.GetRepo().Owner.Login
		repo := *checkSuite.GetRepo().Name
		installation := *checkSuite.Installation.ID
		suite, err := getOrCreateCheckSuite(ctx, sha, owner, repo, installation)
		if err != nil || suite == nil {
			return false, err
		}

		if *checkSuite.Action == "rerequested" {
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

	if checkRun.Action != nil &&
		(*checkRun.Action == "created" || *checkRun.Action == "rerequested") {
		name, sha := *checkRun.CheckRun.Name, *checkRun.CheckRun.HeadSHA
		log.Debugf("Check run %s @ %s %s", name, sha[:7], *checkRun.Action)
		spec, err := shared.ParseProductSpec(*checkRun.CheckRun.Name)
		if err != nil {
			log.Errorf("Failed to parse \"%s\" as product spec")
		}
		return completeChecksForExistingRuns(ctx, sha, spec)
	}
	return false, nil
}

func completeChecksForExistingRuns(ctx context.Context, sha string, products ...shared.ProductSpec) (bool, error) {
	// Jump straight to completed check_run for already-present runs for the SHA.
	products = shared.ProductSpecs(products).OrDefault()
	runs, err := shared.LoadTestRuns(ctx, products, nil, sha[:10], nil, nil, nil)
	if err != nil {
		return false, fmt.Errorf("Failed to load test runs: %s", err.Error())
	}
	createdSome := false
	for _, run := range runs {
		created, err := completeCheckRun(ctx, sha, run.BrowserName)
		createdSome = createdSome || created
		if err != nil {
			return createdSome, err
		}
	}
	return createdSome, nil
}

func createCheckRun(ctx context.Context, suite shared.CheckSuite, opts github.CreateCheckRunOptions) (bool, error) {
	log := shared.GetLogger(ctx)
	status := ""
	if opts.Status != nil {
		status = *opts.Status
	}
	log.Debugf("Creating %s %s check_run for %s/%s", status, opts.Name, suite.Owner, suite.Repo)
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
		log.Infof("Created check_run %v", checkRun.ID)
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
		// NOTE(lukebjerring): This is https://github.com/apps/wpt-fyi-status-check
		Issuer: "19965",
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return jwtToken.SignedString(key)
}

func getDetailsURL(ctx context.Context, sha, browser string) *url.URL {
	hostname := shared.GetHostname(ctx)
	detailsURL, _ := url.Parse(fmt.Sprintf("https://%s/results/", hostname))
	filter := shared.TestRunFilter{}
	filter.Products, _ = shared.ParseProductSpecs(browser, browser)
	filter.Products[0].Labels = mapset.NewSet("master")
	filter.Products[1].Revision = sha
	query := filter.ToQuery()
	query.Set("diff", "")
	detailsURL.RawQuery = query.Encode()
	return detailsURL
}
