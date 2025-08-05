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

	"github.com/google/go-github/v74/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

const requestedAction = "requested"
const rerequestedAction = "rerequested"

// webhookGithubEvent represents the allowed GitHub webhook event types.
type webhookGithubEvent string

const (
	eventCheckSuite  webhookGithubEvent = "check_suite"
	eventCheckRun    webhookGithubEvent = "check_run"
	eventPullRequest webhookGithubEvent = "pull_request"
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
	ctx := r.Context()
	log := shared.GetLogger(ctx)
	ds := shared.NewAppEngineDatastore(ctx, false)

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Errorf("Invalid content-type: %s", contentType)
		w.WriteHeader(http.StatusBadRequest)

		return
	}
	event := r.Header.Get("X-GitHub-Event")
	inputEvent := webhookGithubEvent(event)
	switch inputEvent {
	case eventCheckSuite, eventCheckRun, eventPullRequest:
		break
	default:
		log.Debugf("Ignoring %s event", event)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	secret, err := shared.GetSecret(ds, "github-check-webhook-secret")
	if err != nil {
		log.Errorf("Missing secret: github-check-webhook-secret")
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
	api := NewAPI(ctx)
	switch inputEvent {
	case eventCheckSuite:
		processed, err = handleCheckSuiteEvent(api, payload)
	case eventCheckRun:
		processed, err = handleCheckRunEvent(api, payload)
	case eventPullRequest:
		processed, err = handlePullRequestEvent(api, payload)
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
}

// handleCheckSuiteEvent handles a check_suite (re)requested event by ensuring
// that a check_run exists for each product that contains results for the head SHA.
// nolint:gocognit // TODO: Fix gocognit lint error
func handleCheckSuiteEvent(api API, payload []byte) (bool, error) {
	log := shared.GetLogger(api.Context())
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

	log.Debugf("Check suite %s: %s/%s @ %s (App %v, ID %v)",
		action,
		owner,
		repo,
		shared.CropString(sha, 7),
		appName,
		appID,
	)

	if !isWPTFYIApp(appID) {
		log.Infof("Ignoring check_suite App ID %v", appID)

		return false, nil
	}

	login := checkSuite.GetSender().GetLogin()
	if !checksEnabledForUser(api, login) {
		log.Infof("Checks not enabled for sender %s", login)

		return false, nil
	}

	// nolint:nestif // TODO: Fix nestif lint error
	if action == requestedAction || action == rerequestedAction {
		pullRequests := checkSuite.GetCheckSuite().PullRequests
		prNumbers := []int{}
		for _, pr := range pullRequests {
			if pr.GetBase().GetRepo().GetID() == wptRepoID {
				prNumbers = append(prNumbers, pr.GetNumber())
			}
		}

		installationID := checkSuite.GetInstallation().GetID()
		if action == requestedAction {
			for _, p := range pullRequests {
				destRepoID := p.GetBase().GetRepo().GetID()
				if destRepoID == wptRepoID && p.GetHead().GetRepo().GetID() != destRepoID {
					// Errors are already logged by CreateWPTCheckSuite
					_, _ = api.CreateWPTCheckSuite(appID, installationID, sha, prNumbers...)
				}
			}
		}

		suite, err := getOrCreateCheckSuite(api.Context(), sha, owner, repo, appID, installationID, prNumbers...)
		if err != nil || suite == nil {
			return false, err
		}

		if action == rerequestedAction {
			return scheduleProcessingForExistingRuns(api.Context(), sha)
		}
	}

	return false, nil
}

// handleCheckRunEvent handles a check_run rerequested events by updating
// the status based on whether results for the check_run's product exist.
func handleCheckRunEvent(
	api API,
	payload []byte) (bool, error) {

	log := shared.GetLogger(api.Context())
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
	if !checksEnabledForUser(api, login) {
		log.Infof("Checks not enabled for sender %s", login)

		return false, nil
	}

	// Determine whether or not we need to schedule processing the results
	// of a CheckRun. The 'requested_action' event occurs when a user
	// clicks on one of the 'action' buttons we setup as part of our
	// CheckRuns[0]; see summaries.Summary.GetActions().
	//
	// [0]: https://developer.github.com/v3/checks/runs/#check-runs-and-requested-actions
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
			err := api.IgnoreFailure(
				login,
				owner,
				repo,
				checkRun.GetCheckRun(),
				checkRun.GetInstallation())

			return err == nil, err
		case "cancel":
			err := api.CancelRun(
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
		// Errors are logged by ScheduleResultsProcessing
		_ = api.ScheduleResultsProcessing(sha, spec)

		return true, nil
	}
	log.Debugf("Ignoring %s action for %s check_run", action, status)

	return false, nil
}

// handlePullRequestEvent reaches to pull requests from forks, ensuring that a
// GitHub check_suite is created in the main WPT repository for those. GitHub
// automatically creates a check_suite for code pushed to the WPT repository,
// so we don't need to do anything for same-repo pull requests.
func handlePullRequestEvent(api API, payload []byte) (bool, error) {
	log := shared.GetLogger(api.Context())
	var pullRequest github.PullRequestEvent
	if err := json.Unmarshal(payload, &pullRequest); err != nil {
		return false, err
	}

	login := pullRequest.GetPullRequest().GetUser().GetLogin()
	if !checksEnabledForUser(api, login) {
		log.Infof("Checks not enabled for sender %s", login)

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
		appID, installationID := api.GetWPTRepoAppInstallationIDs()

		return api.CreateWPTCheckSuite(appID, installationID, sha, pullRequest.GetNumber())
	}

	return false, nil
}

func scheduleProcessingForExistingRuns(ctx context.Context, sha string, products ...shared.ProductSpec) (bool, error) {
	// Jump straight to completed check_run for already-present runs for the SHA.
	store := shared.NewAppEngineDatastore(ctx, false)
	products = shared.ProductSpecs(products).OrDefault()
	runsByProduct, err := store.TestRunQuery().LoadTestRuns(products, nil, shared.SHAs{sha}, nil, nil, nil, nil)
	if err != nil {
		return false, fmt.Errorf("failed to load test runs: %s", err.Error())
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

// checksEnabledForUser returns if a commit from a given GitHub username should
// cause wpt.fyi or staging.wpt.fyi summary results to show up in the GitHub
// UI. Currently this is enabled for all users on prod, but only for some users
// on staging to avoid having a confusing double-set of checks appear.
func checksEnabledForUser(api API, login string) bool {
	if api.IsFeatureEnabled(checksForAllUsersFeature) {
		return true
	}
	enabledLogins := []string{
		"chromium-wpt-export-bot",
		"gsnedders",
		"jgraham",
		"jugglinmike",
		"lukebjerring",
		"Ms2ger",
	}

	return shared.StringSliceContains(enabledLogins, login)
}
