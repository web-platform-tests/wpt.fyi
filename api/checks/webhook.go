// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/api/azure"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func isWPTFYIApp(appID int64) bool {
	switch appID {
	case wptfyiCheckAppID, wptfyiStagingCheckAppID, checksStagingAppID:
		return true
	}
	return false
}

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
	case "check_suite", "check_run", "pull_request":
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
	aeAPI := shared.NewAppEngineAPI(ctx)
	checksAPI := NewAPI(ctx)
	if event == "check_suite" {
		processed, err = handleCheckSuiteEvent(aeAPI, checksAPI, payload)
	} else if event == "check_run" {
		azureAPI := azure.NewAPI(ctx)
		processed, err = handleCheckRunEvent(aeAPI, checksAPI, azureAPI, payload)
	} else if event == "pull_request" {
		processed, err = handlePullRequestEvent(aeAPI, checksAPI, payload)
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
func handleCheckSuiteEvent(aeAPI shared.AppEngineAPI, checksAPI API, payload []byte) (bool, error) {
	log := shared.GetLogger(aeAPI.Context())
	var checkSuite github.CheckSuiteEvent
	if err := json.Unmarshal(payload, &checkSuite); err != nil {
		return false, err
	}

	appID := checkSuite.GetCheckSuite().GetApp().GetID()
	if !isWPTFYIApp(appID) {
		log.Infof("Ignoring check_suite App ID %v", appID)
		return false, nil
	}

	login := checkSuite.GetSender().GetLogin()
	if !isUserWhitelisted(aeAPI, login) {
		log.Infof("Sender %s not whitelisted for wpt.fyi checks", login)
		return false, nil
	}

	action := checkSuite.GetAction()
	if action == "requested" || action == "rerequested" {
		owner := checkSuite.GetRepo().GetOwner().GetLogin()
		repo := checkSuite.GetRepo().GetName()
		sha := checkSuite.GetCheckSuite().GetHeadSHA()
		log.Debugf("Check suite %s: %s/%s @ %s", action, owner, repo, sha[:7])

		pullRequests := checkSuite.GetCheckSuite().PullRequests
		prNumbers := []int{}
		for _, pr := range pullRequests {
			if pr.GetBase().GetRepo().GetID() == wptRepoID {
				prNumbers = append(prNumbers, pr.GetNumber())
			}
		}

		installationID := checkSuite.GetInstallation().GetID()
		if action == "requested" {
			// For new suites, check if the pull is across forks; if so, request a suite
			// on the main repo (web-platform-tests/wpt) too.
			for _, p := range pullRequests {
				destRepoID := p.GetBase().GetRepo().GetID()
				if destRepoID == wptRepoID && p.GetHead().GetRepo().GetID() != destRepoID {
					checksAPI.CreateWPTCheckSuite(appID, installationID, sha, prNumbers...)
				}
			}
		}

		suite, err := getOrCreateCheckSuite(aeAPI.Context(), sha, owner, repo, appID, installationID, prNumbers...)
		if err != nil || suite == nil {
			return false, err
		}

		if action == "rerequested" {
			return scheduleProcessingForExistingRuns(aeAPI.Context(), sha)
		}
	}
	return false, nil
}

// handleCheckRunEvent handles a check_run rerequested events by updating
// the status based on whether results for the check_run's product exist.
func handleCheckRunEvent(aeAPI shared.AppEngineAPI, checksAPI API, azureAPI azure.API, payload []byte) (bool, error) {
	log := shared.GetLogger(aeAPI.Context())
	checkRun := new(github.CheckRunEvent)
	if err := json.Unmarshal(payload, checkRun); err != nil {
		return false, err
	}

	appID := checkRun.GetCheckRun().GetApp().GetID()
	if !isWPTFYIApp(appID) && appID != azure.PipelinesAppID {
		log.Infof("Ignoring check_suite App ID %v", appID)
		return false, nil
	}

	login := checkRun.GetSender().GetLogin()
	if !isUserWhitelisted(aeAPI, login) {
		log.Infof("Sender %s not whitelisted for wpt.fyi checks", login)
		return false, nil
	}

	action := checkRun.GetAction()
	status := checkRun.GetCheckRun().GetStatus()

	shouldSchedule := false
	if appID == azure.PipelinesAppID {
		return azureAPI.HandleCheckRunEvent(checkRun)
	} else if (action == "created" && status != "completed") || action == "rerequested" {
		shouldSchedule = true
	} else if action == "requested_action" {
		actionID := checkRun.GetRequestedAction().Identifier
		switch actionID {
		case "recompute":
			shouldSchedule = true
		case "ignore":
			err := checksAPI.IgnoreFailure(
				checkRun.GetSender().GetLogin(),
				checkRun.GetRepo().GetOwner().GetLogin(),
				checkRun.GetRepo().GetName(),
				checkRun.GetCheckRun(),
				checkRun.GetInstallation())
			return err == nil, err
		case "cancel":
			err := checksAPI.CancelRun(
				checkRun.GetSender().GetLogin(),
				checkRun.GetRepo().GetOwner().GetLogin(),
				checkRun.GetRepo().GetName(),
				checkRun.GetCheckRun(),
				checkRun.GetInstallation())
			return err == nil, err
		default:
			log.Debugf("Ignoring %s action with id %s", action, actionID)
			return false, nil
		}
	}
	if shouldSchedule {
		name, sha := checkRun.GetCheckRun().GetName(), checkRun.GetCheckRun().GetHeadSHA()
		log.Debugf("GitHub check run %v (%s @ %s) was %s", checkRun.GetCheckRun().GetID(), name, sha, action)
		spec, err := shared.ParseProductSpec(checkRun.GetCheckRun().GetName())
		if err != nil {
			log.Errorf("Failed to parse \"%s\" as product spec", checkRun.GetCheckRun().GetName())
			return false, err
		}
		checksAPI.ScheduleResultsProcessing(sha, spec)
		return true, nil
	}
	log.Debugf("Ignoring %s action for %s check_run", action, status)
	return false, nil
}

func handlePullRequestEvent(aeAPI shared.AppEngineAPI, checksAPI API, payload []byte) (bool, error) {
	log := shared.GetLogger(aeAPI.Context())
	var pullRequest github.PullRequestEvent
	if err := json.Unmarshal(payload, &pullRequest); err != nil {
		return false, err
	}

	login := pullRequest.GetPullRequest().GetUser().GetLogin()
	if !isUserWhitelisted(aeAPI, login) {
		log.Infof("Sender %s not whitelisted for wpt.fyi checks", login)
		return false, nil
	}

	switch pullRequest.GetAction() {
	case "opened", "synchronize":
		break
	default:
		log.Debugf("Skipping pull request action %s", pullRequest.GetAction())
		return false, nil
	}

	sha := pullRequest.GetPullRequest().GetHead().GetSHA()
	destRepoID := pullRequest.GetPullRequest().GetBase().GetRepo().GetID()
	if destRepoID == wptRepoID && pullRequest.GetPullRequest().GetHead().GetRepo().GetID() != destRepoID {
		// Pull is across forks; request a check suite on the main fork too.
		appID, installationID := checksAPI.GetWPTRepoAppInstallationIDs()
		return checksAPI.CreateWPTCheckSuite(appID, installationID, sha, pullRequest.GetNumber())
	}
	return false, nil
}

func scheduleProcessingForExistingRuns(ctx context.Context, sha string, products ...shared.ProductSpec) (bool, error) {
	// Jump straight to completed check_run for already-present runs for the SHA.
	store := shared.NewAppEngineDatastore(ctx, false)
	products = shared.ProductSpecs(products).OrDefault()
	runsByProduct, err := store.TestRunQuery().LoadTestRuns(products, nil, shared.SHAs{sha}, nil, nil, nil, nil)
	if err != nil {
		return false, fmt.Errorf("Failed to load test runs: %s", err.Error())
	}
	createdSome := false
	api := NewAPI(ctx)
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

// createCheckRun submits an http POST to create the check run on GitHub, handling JWT auth for the app.
func createCheckRun(ctx context.Context, suite shared.CheckSuite, opts github.CreateCheckRunOptions) (bool, error) {
	log := shared.GetLogger(ctx)
	status := ""
	if opts.Status != nil {
		status = *opts.Status
	}
	log.Debugf("Creating %s %s check_run for %s/%s @ %s", status, opts.Name, suite.Owner, suite.Repo, suite.SHA)
	if suite.AppID == 0 {
		suite.AppID = wptfyiStagingCheckAppID
	}
	client, err := getGitHubClient(ctx, suite.AppID, suite.InstallationID)
	if err != nil {
		log.Errorf("Failed to create JWT client: %s", err.Error())
		return false, err
	}

	checkRun, resp, err := client.Checks.CreateCheckRun(ctx, suite.Owner, suite.Repo, opts)
	if err != nil {
		msg := "Failed to create check_run"
		if resp != nil {
			msg = fmt.Sprintf("%s: %s", msg, resp.Status)
		}
		log.Warningf(msg)
		return false, err
	} else if checkRun != nil {
		log.Infof("Created check_run %v", checkRun.GetID())
	}
	return true, nil
}

func isUserWhitelisted(aeAPI shared.AppEngineAPI, login string) bool {
	if aeAPI.IsFeatureEnabled(checksForAllUsersFeature) {
		return true
	}
	whitelist := []string{
		"autofoolip",
		"chromium-wpt-export-bot",
		"foolip",
		"gsnedders",
		"jgraham",
		"jugglinmike",
		"lukebjerring",
		"Ms2ger",
	}
	return shared.StringSliceContains(whitelist, login)
}
