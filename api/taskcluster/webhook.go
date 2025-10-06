// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -build_flags=--mod=mod -destination mock_taskcluster/webhook_mock.go github.com/web-platform-tests/wpt.fyi/api/taskcluster API

package taskcluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set"
	"github.com/google/go-github/v75/github"
	tcurls "github.com/taskcluster/taskcluster-lib-urls"
	"github.com/taskcluster/taskcluster/v90/clients/client-go/tcqueue"
	uc "github.com/web-platform-tests/wpt.fyi/api/receiver/client"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// AppID is the ID of the Community-TC GitHub app.
const AppID = int64(40788)

const uploaderName = "taskcluster"
const completedState = "completed"

var (
	// TaskNameRegex is based on task names in
	// https://github.com/web-platform-tests/wpt/blob/master/tools/ci/tc/tasks/test.yml.
	TaskNameRegex = regexp.MustCompile(`^wpt-([a-z_]+-[a-z]+)-([a-z]+(?:-[a-z]+)*)(?:-\d+)?$`)
	// Taskcluster has used different forms of URLs in their Check & Status
	// updates in history. We accept all of them.
	// See TestExtractTaskGroupID for examples.
	inspectorURLRegex       = regexp.MustCompile(`^(https://[^/]*)/task-group-inspector/#/([^/]*)`)
	taskURLRegex            = regexp.MustCompile(`^(https://[^/]*)(?:/tasks)?/groups/([^/]*)(?:/tasks/([^/]*))?`)
	checkRunDetailsURLRegex = regexp.MustCompile(`^(https://[^/]*)/tasks/([^/]*)`)
)

// Non-fatal error when there is no result (e.g. nothing finishes yet).
var errNoResults = errors.New("no result URLs found in task group")

// TaskInfo is an abstraction of a Taskcluster task, containing the necessary
// information for us to process the task in wpt.fyi.
type TaskInfo struct {
	Name   string
	TaskID string
	State  string
}

// TaskGroupInfo is an abstraction of a Taskcluster task group, containing the
// necessary information for us to process the group in wpt.fyi.
type TaskGroupInfo struct {
	TaskGroupID string
	Tasks       []TaskInfo
}

// EventInfo is an abstraction of a GitHub Status event, containing the
// necessary information for us to process the event in wpt.fyi.
type EventInfo struct {
	Sha     string
	RootURL string
	TaskID  string
	Master  bool
	Sender  string
	Group   *TaskGroupInfo
}

// API wraps externally provided methods so we can mock them for testing.
type API interface {
	GetTaskGroupInfo(string, string) (*TaskGroupInfo, error)
	ListCheckRuns(owner string, repo string, checkSuiteID int64) ([]*github.CheckRun, error)
}

type apiImpl struct {
	ctx      context.Context // nolint:containedctx // TODO: Fix containedctx lint error
	ghClient *github.Client
}

// GetStatusEventInfo turns a StatusEventPayload into an EventInfo struct.
func GetStatusEventInfo(status StatusEventPayload, log shared.Logger, api API) (EventInfo, error) {
	if status.SHA == nil {
		return EventInfo{}, errors.New("no sha on taskcluster status event")
	}

	if status.TargetURL == nil {
		return EventInfo{}, errors.New("no target_url on taskcluster status event")
	}

	rootURL, taskGroupID, taskID := ParseTaskclusterURL(*status.TargetURL)
	if taskGroupID == "" {
		return EventInfo{}, fmt.Errorf("unrecognized target_url: %s", *status.TargetURL)
	}

	log.Debugf("Taskcluster task group %s", taskGroupID)

	group, err := api.GetTaskGroupInfo(rootURL, taskGroupID)
	if err != nil {
		return EventInfo{}, err
	}

	event := EventInfo{
		Sha:     *status.SHA,
		RootURL: rootURL,
		TaskID:  taskID,
		Master:  status.IsOnMaster(),
		Sender:  status.GetCommit().GetAuthor().GetLogin(),
		Group:   group,
	}

	return event, nil
}

// GetCheckSuiteEventInfo turns a github.CheckSuiteEvent into an EventInfo struct.
func GetCheckSuiteEventInfo(checkSuite github.CheckSuiteEvent, log shared.Logger, api API) (EventInfo, error) {
	if checkSuite.GetCheckSuite().GetHeadSHA() == "" {
		return EventInfo{}, errors.New("no sha on taskcluster check_suite event")
	}

	log.Debugf("Parsing check_suite event for commit %s", checkSuite.GetCheckSuite().GetHeadSHA())

	owner := checkSuite.GetRepo().GetOwner().GetLogin()
	repo := checkSuite.GetRepo().GetName()
	if owner != shared.WPTRepoOwner || repo != shared.WPTRepoName {
		log.Errorf("Received check_suite event from invalid repo %s/%s", owner, repo)

		return EventInfo{}, errors.New("invalid source repository")
	}

	runs, err := api.ListCheckRuns(owner, repo, checkSuite.GetCheckSuite().GetID())
	if err != nil {
		log.Errorf("Failed to fetch check runs for suite %v: %s", checkSuite.GetCheckSuite().GetID(), err.Error())

		return EventInfo{}, err
	}

	if len(runs) == 0 {
		return EventInfo{}, errors.New("no check_runs for check_suite")
	}

	log.Debugf("Found %d check_runs for check_suite", len(runs))

	rootURL := ""
	group := TaskGroupInfo{} // nolint:exhaustruct // TODO: Fix exhaustruct lint error
	for _, run := range runs {
		matches := checkRunDetailsURLRegex.FindStringSubmatch(run.GetDetailsURL())
		if matches == nil {
			log.Errorf(
				"Unable to parse details URL for suite %v, run %v: %s",
				checkSuite.GetCheckSuite().GetID(),
				run.GetID(),
				run.GetDetailsURL(),
			)

			return EventInfo{}, errors.New("unable to parse check_run details URL")
		}
		if rootURL != "" && rootURL != matches[1] {
			log.Errorf(
				"Conflicting root URLs for runs for suite %v (%s vs %s)",
				checkSuite.GetCheckSuite().GetID(),
				rootURL,
				matches[1],
			)

			return EventInfo{}, errors.New("conflicting root URLs for runs in check_suite")
		}
		rootURL = matches[1]
		taskID := matches[2]

		// The task group's ID appear to be equivalent to the ID of the initial task
		// that was created. For WPT, this is the decision task.
		if run.GetName() == "wpt-decision-task" {
			group.TaskGroupID = taskID
		}

		log.Debugf("Adding task: %s, id: %s, conclusion: %s", run.GetName(), taskID, run.GetConclusion())

		// Reconstruct Taskcluster TaskInfo from the check run without calling Taskcluster API.
		state := run.GetConclusion()
		if state == "success" {
			// Checked in ExtractArtifactURLs.
			state = completedState
		}

		group.Tasks = append(group.Tasks, TaskInfo{
			Name:   run.GetName(),
			TaskID: taskID,
			State:  state,
		})
	}

	event := EventInfo{
		Sha:     checkSuite.GetCheckSuite().GetHeadSHA(),
		RootURL: rootURL,
		// The TaskID is a filter for a specific task. For check_suite events we
		// only ever receieve events for an entire suite, so there is no TaskID.
		TaskID: "",
		Master: checkSuite.GetCheckSuite().GetHeadBranch() == "master",
		Sender: checkSuite.GetSender().GetLogin(),
		Group:  &group,
	}

	return event, nil
}

// tcStatusWebhookHandler reacts to GitHub status webhook events.
func tcStatusWebhookHandler(w http.ResponseWriter, r *http.Request) {
	eventName := r.Header.Get("X-GitHub-Event")
	if r.Header.Get("Content-Type") != "application/json" || (eventName != "status" && eventName != "check_suite") {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	ctx := r.Context()
	ds := shared.NewAppEngineDatastore(ctx, false)
	secret, err := shared.GetSecret(ds, "github-tc-webhook-secret")
	if err != nil {
		http.Error(w, "Unable to verify request: secret not found", http.StatusInternalServerError)

		return
	}

	log := shared.GetLogger(ctx)
	log.Debugf("Retrieved GitHub secret from datastore")

	payload, err := github.ValidatePayload(r, []byte(secret))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)

		return
	}
	log.Debugf("Payload validated against secret")

	log.Debugf("GitHub Delivery: %s", r.Header.Get("X-GitHub-Delivery"))

	aeAPI := shared.NewAppEngineAPI(ctx)

	ghClient, err := aeAPI.GetGitHubClient()
	if err != nil {
		log.Errorf("Failed to get GitHub client: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	api := apiImpl{ctx: aeAPI.Context(), ghClient: ghClient}

	var event EventInfo
	// nolint:nestif // TODO: Fix nestif lint error
	if eventName == "status" {
		var status StatusEventPayload
		if err := json.Unmarshal(payload, &status); err != nil {
			log.Errorf("%v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		if !ShouldProcessStatus(log, &status) {
			w.WriteHeader(http.StatusNoContent)
			fmt.Fprintln(w, "Status was ignored")

			return
		}

		event, err = GetStatusEventInfo(status, log, api)
	} else {
		var checkSuite github.CheckSuiteEvent
		if err := json.Unmarshal(payload, &checkSuite); err != nil {
			log.Errorf("%v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		if checkSuite.GetCheckSuite().GetApp().GetID() != AppID {
			log.Debugf("Ignoring non-Taskcluster app: %s (%s)",
				checkSuite.GetCheckSuite().GetApp().GetName(),
				checkSuite.GetCheckSuite().GetApp().GetID())
			w.WriteHeader(http.StatusNoContent)
			fmt.Fprintln(w, "Status was ignored")

			return
		}

		// As a webhook we should only receive completed check_suite events, as per
		// https://developer.github.com/webhooks/event-payloads/#check_suite
		if checkSuite.GetAction() != completedState || checkSuite.GetCheckSuite().GetStatus() != completedState {
			log.Errorf("Received non-completed check_suite event (action: %s, status: %s)",
				checkSuite.GetAction(), checkSuite.GetCheckSuite().GetStatus())
			http.Error(w, "Non-completed check_suite event", http.StatusBadRequest)

			return
		}

		event, err = GetCheckSuiteEventInfo(checkSuite, log, api)
	}
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	labels := mapset.NewSet()
	if event.Master {
		labels.Add(shared.MasterLabel)
	}
	if event.Sender != "" {
		labels.Add(shared.GetUserLabel(event.Sender))
	}

	processed, err := processTaskclusterBuild(aeAPI, event, shared.ToStringSlice(labels)...)

	if errors.Is(err, errNoResults) {
		log.Infof("%v", err)
		http.Error(w, err.Error(), http.StatusNoContent)

		return
	}
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	if processed {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Taskcluster tasks were sent to results receiver")
	} else {
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintln(w, "Status was ignored")
	}
}

// StatusEventPayload wraps a github.StatusEvent so we can declare methods on it
// https://developer.github.com/v3/activity/events/types/#statusevent
type StatusEventPayload struct {
	github.StatusEvent
}

// IsCompleted checks if a github.StatusEvent has completed.
func (s StatusEventPayload) IsCompleted() bool {
	return s.GetState() == "success" || s.GetState() == "failure"
}

// IsTaskcluster checks if a github.StatusEvent is from Taskcluster.
func (s StatusEventPayload) IsTaskcluster() bool {
	return s.Context != nil && (strings.HasPrefix(*s.Context, "Taskcluster") ||
		strings.HasPrefix(*s.Context, "Community-TC"))
}

// IsOnMaster checks if a github.StatusEvent affects the master branch.
func (s StatusEventPayload) IsOnMaster() bool {
	for _, branch := range s.Branches {
		if branch.Name != nil && *branch.Name == "master" {
			return true
		}
	}

	return false
}

func processTaskclusterBuild(aeAPI shared.AppEngineAPI, event EventInfo, labels ...string) (bool, error) {
	ctx := aeAPI.Context()
	log := shared.GetLogger(ctx)

	if event.TaskID != "" {
		log.Debugf("Taskcluster task %s", event.TaskID)
	}

	urlsByProduct, err := ExtractArtifactURLs(event.RootURL, log, event.Group, event.TaskID)
	if err != nil {
		return false, err
	}

	uploader, err := aeAPI.GetUploader(uploaderName)
	if err != nil {
		log.Errorf("Failed to get uploader creds from Datastore")

		return false, err
	}

	err = CreateAllRuns(
		log,
		shared.NewAppEngineAPI(ctx),
		event.Sha,
		uploader.Username,
		uploader.Password,
		urlsByProduct,
		labels)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ShouldProcessStatus determines whether we are interested in processing a
// given StatusEventPayload or not.
func ShouldProcessStatus(log shared.Logger, status *StatusEventPayload) bool {
	if !status.IsCompleted() {
		log.Debugf("Ignoring status: %s", status.GetState())

		return false
	} else if !status.IsTaskcluster() {
		log.Debugf("Ignoring non-Taskcluster context: %s", status.GetContext())

		return false
	}

	return true
}

// ParseTaskclusterURL splits a given URL into its root URL, the Taskcluster
// group id, and an optional specific task ID.
func ParseTaskclusterURL(targetURL string) (rootURL, taskGroupID, taskID string) {
	if matches := inspectorURLRegex.FindStringSubmatch(targetURL); matches != nil {
		rootURL = matches[1]
		taskGroupID = matches[2]
	} else if matches := taskURLRegex.FindStringSubmatch(targetURL); matches != nil {
		rootURL = matches[1]
		taskGroupID = matches[2]
		// matches[3] may be an empty string -- the capturing group is optional.
		taskID = matches[3]
	}
	// Special case for old Taskcluster instance, which uses subdomains for
	// different services and we need to strip the subdomain away.
	if strings.HasSuffix(rootURL, "taskcluster.net") {
		rootURL = "https://taskcluster.net"
	}

	return rootURL, taskGroupID, taskID
}

func (api apiImpl) GetTaskGroupInfo(rootURL string, groupID string) (*TaskGroupInfo, error) {
	queue := tcqueue.New(nil, rootURL)

	group := TaskGroupInfo{ // nolint:exhaustruct // TODO: Fix exhaustruct lint error
		TaskGroupID: groupID,
	}
	continuationToken := ""

	for {
		ltgr, err := queue.ListTaskGroup(groupID, continuationToken, "1000")
		if err != nil {
			return nil, err
		}

		for _, task := range ltgr.Tasks {
			group.Tasks = append(group.Tasks, TaskInfo{
				Name:   task.Task.Metadata.Name,
				TaskID: task.Status.TaskID,
				State:  task.Status.State,
			})
		}

		continuationToken = ltgr.ContinuationToken
		if continuationToken == "" {
			break
		}
	}

	return &group, nil
}

func (api apiImpl) ListCheckRuns(owner string, repo string, checkSuiteID int64) ([]*github.CheckRun, error) {
	var runs []*github.CheckRun
	// nolint:exhaustruct // TODO: Fix exhaustruct lint error.
	options := github.ListCheckRunsOptions{
		// nolint:exhaustruct // TODO: Fix exhaustruct lint error.
		ListOptions: github.ListOptions{
			// 100 is the maximum allowed items per page[0], but due to
			// https://github.com/web-platform-tests/wpt/issues/27243 we
			// request only 25 at a time.
			//
			// [0]: https://developer.github.com/v3/guides/traversing-with-pagination/#changing-the-number-of-items-received
			PerPage: 25,
		},
	}

	// As a safety-check, we will not do more than 20 iterations (at 25
	// check runs per page, this gives us a 500 run upper limit).
	for i := 0; i < 20; i++ {
		result, response, err := api.ghClient.Checks.ListCheckRunsCheckSuite(api.ctx, owner, repo, checkSuiteID, &options)
		if err != nil {
			return runs, err
		}

		runs = append(runs, result.CheckRuns...)

		// GitHub APIs indicate being on the last page by not returning any
		// value for NextPage, which go-github translates into zero.
		// See https://gowalker.org/github.com/google/go-github/github#Response
		if response.NextPage == 0 {
			return runs, nil
		}

		// Setup for the next call.
		options.Page = response.NextPage
	}

	return runs, errors.New("more than 500 CheckRuns returned for CheckSuite")
}

// ArtifactURLs holds the results and screenshot URLs for a Taskcluster run.
type ArtifactURLs struct {
	Results     []string
	Screenshots []string
}

// ExtractArtifactURLs extracts the results and screenshot URLs for a set of
// tasks in a TaskGroupInfo.
func ExtractArtifactURLs(rootURL string, log shared.Logger, group *TaskGroupInfo, taskID string) (
	urlsByProduct map[string]ArtifactURLs, err error) {
	urlsByProduct = make(map[string]ArtifactURLs)
	failures := mapset.NewSet()
	log.Debugf("Extracting artifact URLs for %d tasks", len(group.Tasks))
	for _, task := range group.Tasks {
		id := task.TaskID
		if id == "" {
			return nil, fmt.Errorf("task group %s has a task without taskId", group.TaskGroupID)
		} else if taskID != "" && taskID != id {
			log.Debugf("Skipping task %s", id)

			continue
		}

		matches := TaskNameRegex.FindStringSubmatch(task.Name)
		if len(matches) != 3 { // full match, browser-channel, test type
			log.Infof("Ignoring unrecognized task: %s", task.Name)

			continue
		}
		product := matches[1]
		switch matches[2] {
		case "stability":
			// Skip stability checks.
			continue
		case "results":
			product += "-" + shared.PRHeadLabel
		case "results-without-changes":
			product += "-" + shared.PRBaseLabel
		}

		if task.State != completedState {
			log.Infof("Task group %s has a non-successful task: %s; %s will be ignored in this group.",
				group.TaskGroupID, id, product)
			failures.Add(product)

			continue
		}

		urls := urlsByProduct[product]
		// Generate some URLs that point directly to
		// https://docs.taskcluster.net/docs/reference/platform/queue/api#getLatestArtifact
		urls.Results = append(urls.Results,
			tcurls.API(
				rootURL, "queue", "v1",
				fmt.Sprintf("/task/%s/artifacts/public/results/wpt_report.json.gz", id)))
		// wpt_screenshot.txt.gz might not exist, which is NOT a fatal error in the receiver.
		urls.Screenshots = append(urls.Screenshots,
			tcurls.API(
				rootURL, "queue", "v1",
				fmt.Sprintf("/task/%s/artifacts/public/results/wpt_screenshot.txt.gz", id)))
		// urls is a *copy* of the value so we must store it back to the map.
		urlsByProduct[product] = urls
	}

	for failure := range failures.Iter() {
		delete(urlsByProduct, failure.(string))
	}

	if len(urlsByProduct) == 0 {
		return nil, errNoResults
	}

	return urlsByProduct, nil
}

// CreateAllRuns creates run entries in wpt.fyi for a set of products coming
// from Taskcluster.
func CreateAllRuns(
	log shared.Logger,
	aeAPI shared.AppEngineAPI,
	sha,
	username,
	password string,
	urlsByProduct map[string]ArtifactURLs,
	labels []string) error {
	errors := make(chan error, len(urlsByProduct))
	var wg sync.WaitGroup
	wg.Add(len(urlsByProduct))
	for product, urls := range urlsByProduct {
		go func(product string, urls ArtifactURLs) {
			defer wg.Done()
			log.Infof("Reports for %s: %v", product, urls)

			// chrome-dev-pr_head => [chrome, dev, pr_head]
			bits := strings.Split(product, "-")
			labelsForRun := labels
			switch lastBit := bits[len(bits)-1]; lastBit {
			case shared.PRBaseLabel, shared.PRHeadLabel:
				// We have seen cases where Community-TC triggers a pull request
				// for merged commits. To guard against that, we strip the
				// master label here.
				for i, label := range labelsForRun {
					if label == shared.MasterLabel {
						labelsForRun = append(labelsForRun[:i], labelsForRun[i+1:]...)

						break
					}
				}
				labelsForRun = append(labelsForRun, lastBit)
			}

			uploadClient := uc.NewClient(aeAPI)
			err := uploadClient.CreateRun(sha, username, password, urls.Results, urls.Screenshots, nil, labelsForRun)
			if err != nil {
				errors <- err

				return
			}
		}(product, urls)
	}
	wg.Wait()
	close(errors)

	return shared.NewMultiErrorFromChan(errors, "sending Taskcluster runs to results receiver")
}
