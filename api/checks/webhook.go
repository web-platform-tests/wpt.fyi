// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/google/go-github/v31/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

var runNameRegex = regexp.MustCompile(`^(?:(?:staging\.)?wpt\.fyi - )(.*)$`)

func isWPTFYIApp(appID int64) bool {
	return appID == wptfyiCheckAppID || appID == wptfyiStagingCheckAppID
}

// checkWebhookHandler handles GitHub events relating to our wpt.fyi and
// staging.wpt.fyi GitHub Apps[0], sent to the /api/webhook/check endpoint.
//
// [0]: https://github.com/apps/wpt-fyi and https://github.com/apps/staging-wpt-fyi
func checkWebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)
	ds := shared.NewAppEngineDatastore(ctx, false)

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

	secret, err := shared.GetSecret(ds, "github-check-webhook-secret")
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
		processed, err = handleCheckRunEvent(aeAPI, checksAPI, payload)
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

// TODO: Better docstring
// handleCheckSuiteEvent handles a check_suite (re)requested event by ensuring
// that a check_run exists for each product that contains results for the head SHA.
func handleCheckSuiteEvent(aeAPI shared.AppEngineAPI, checksAPI API, payload []byte) (bool, error) {
	log := shared.GetLogger(aeAPI.Context())
	var checkSuite github.CheckSuiteEvent
	if err := json.Unmarshal(payload, &checkSuite); err != nil {
		return false, err
	}

	action := checkSuite.GetAction()
	owner := checkSuite.GetRepo().GetOwner().GetLogin()
	repo := checkSuite.GetRepo().GetName()
	sha := checkSuite.GetCheckSuite().GetHeadSHA()
	appName := checkSuite.GetCheckSuite().GetApp().GetName()
	appID := checkSuite.GetCheckSuite().GetApp().GetID()

	log.Debugf("Check suite %s: %s/%s @ %s (App %v, ID %v)", action, owner, repo, shared.CropString(sha, 7), appName, appID)

	if !isWPTFYIApp(appID) {
		log.Infof("Ignoring check_suite App ID %v", appID)
		return false, nil
	}

	login := checkSuite.GetSender().GetLogin()
	if !isUserWhitelisted(aeAPI, login) {
		log.Infof("Sender %s not whitelisted for wpt.fyi checks", login)
		return false, nil
	}

	// A CheckSuite is 'requested' whenever a commit is made in the
	// repository, and can be re-requested by a user clicking a 'retry'
	// button in the GitHub UI. For either of these events, we... TODO.
	if action == "requested" || action == "rerequested" {
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
			//
			// TODO: Explain why we need to do this when we have the pull_request event too.
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

		// TODO: Explain why a rerequest should lead to calling
		// scheduleProcessingForExistingRuns.
		if action == "rerequested" {
			return scheduleProcessingForExistingRuns(aeAPI.Context(), sha)
		}
	}
	return false, nil
}

// TODO: better docstring
// handleCheckRunEvent handles a check_run rerequested events by updating
// the status based on whether results for the check_run's product exist.
func handleCheckRunEvent(
	aeAPI shared.AppEngineAPI,
	checksAPI API,
	payload []byte) (bool, error) {

	log := shared.GetLogger(aeAPI.Context())
	checkRun := new(github.CheckRunEvent)
	if err := json.Unmarshal(payload, checkRun); err != nil {
		return false, err
	}

	action := checkRun.GetAction()
	owner := checkRun.GetRepo().GetOwner().GetLogin()
	repo := checkRun.GetRepo().GetName()
	sha := checkRun.GetCheckRun().GetHeadSHA()
	appName := checkRun.GetCheckRun().GetApp().GetName()
	appID := checkRun.GetCheckRun().GetApp().GetID()

	log.Debugf("Check run %s: %s/%s @ %s (App %v, ID %v)", action, owner, repo, shared.CropString(sha, 7), appName, appID)

	if !isWPTFYIApp(appID) {
		log.Infof("Ignoring check_run App ID %v", appID)
		return false, nil
	}

	login := checkRun.GetSender().GetLogin()
	if !isUserWhitelisted(aeAPI, login) {
		log.Infof("Sender %s not whitelisted for wpt.fyi checks", login)
		return false, nil
	}

	// Determine whether or not we need to schedule processing the results
	// of a CheckRun. The 'requested_action' event occurs when a user
	// clicks on one of the 'action' buttons we setup as part of our
	// CheckRuns[0]; see summaries.Summary.GetActions().
	//
	// [0]: https://developer.github.com/v3/checks/runs/#check-runs-and-requested-actions
	//
	// TODO: It's not clear to me who created the CheckRun that created
	// this event. Did 'we' (wpt.fyi app) create this CheckRun, or is this
	// an event from a CheckRun from AP/TaskCluster? It seems like there's a loop?
	//
	//	handleCheckRunEvent --> ScheduleResultsProcessing --> /api/checks/{sha}
	//	--> updateChecksHandlers --> updateCheckRunSummary --> CreateCheckRun
	//	--> [GitHub] --> /api/checks/webhook --> checkWebhookHandler
	//	--> handleCheckRunEvent ...
	status := checkRun.GetCheckRun().GetStatus()
	shouldSchedule := false
	if (action == "created" && status != "completed") || action == "rerequested" {
		shouldSchedule = true
	} else if action == "requested_action" {
		actionID := checkRun.GetRequestedAction().Identifier
		switch actionID {
		case "recompute":
			shouldSchedule = true
		case "ignore":
			err := checksAPI.IgnoreFailure(
				login,
				owner,
				repo,
				checkRun.GetCheckRun(),
				checkRun.GetInstallation())
			return err == nil, err
		case "cancel":
			err := checksAPI.CancelRun(
				login,
				owner,
				repo,
				checkRun.GetCheckRun(),
				checkRun.GetInstallation())
			return err == nil, err
		default:
			log.Debugf("Ignoring %s action with id %s", action, actionID)
			return false, nil
		}
	}

	if shouldSchedule {
		name := checkRun.GetCheckRun().GetName()
		log.Debugf("GitHub check run %v (%s @ %s) was %s", checkRun.GetCheckRun().GetID(), name, sha, action)
		// Strip any "wpt.fyi - " prefix.
		if runNameRegex.MatchString(name) {
			name = runNameRegex.FindStringSubmatch(name)[1]
		}
		spec, err := shared.ParseProductSpec(name)
		if err != nil {
			log.Errorf("Failed to parse \"%s\" as product spec", name)
			return false, err
		}
		checksAPI.ScheduleResultsProcessing(sha, spec)
		return true, nil
	}
	log.Debugf("Ignoring %s action for %s check_run", action, status)
	return false, nil
}

// handlePullRequestEvent reaches to pull requests from forks, ensuring that a
// GitHub check_suite is created in the main WPT repository for those. GitHub
// automatically creates a check_suite for code pushed to the WPT repository,
// so we don't need to do anything for same-repo pull requests.
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

// TODO: Document what this function is for.
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
// TODO: Shouldn't this be in /api/check/api.go?
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

// isUserWhitelisted checks if a commit from a given GitHub username should
// result in wpt.fyi or staging.wpt.fyi summary results showing up. Currently
// this is enabled for all users on prod, but only for some users on staging to
// avoid having a confusing double-set of checks appear on the GitHub UI.
//
// TODO: Should we remove checksForAllUsersFeature and just differentiate based
// on whether the AppId is prod/staging?
func isUserWhitelisted(aeAPI shared.AppEngineAPI, login string) bool {
	if aeAPI.IsFeatureEnabled(checksForAllUsersFeature) {
		return true
	}
	whitelist := []string{
		"chromium-wpt-export-bot",
		"gsnedders",
		"jgraham",
		"jugglinmike",
		"lukebjerring",
		"Ms2ger",
	}
	return shared.StringSliceContains(whitelist, login)
}
