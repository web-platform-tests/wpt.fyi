// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination mock_taskcluster/webhook_mock.go github.com/web-platform-tests/wpt.fyi/api/taskcluster API

package taskcluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set"
	"github.com/google/go-github/v31/github"
	tcurls "github.com/taskcluster/taskcluster-lib-urls"
	"github.com/taskcluster/taskcluster/v25/clients/client-go/tcqueue"
	uc "github.com/web-platform-tests/wpt.fyi/api/receiver/client"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

const uploaderName = "taskcluster"
const flagPendingChecks = "pendingChecks"

var (
	// TaskNameRegex is based on task names in https://github.com/web-platform-tests/wpt/blob/master/tools/ci/tc/tasks/test.yml
	TaskNameRegex = regexp.MustCompile(`^wpt-([a-z_]+-[a-z]+)-([a-z]+(?:-[a-z]+)*)(?:-\d+)?$`)
	// Taskcluster has used different forms of URLs in their Check & Status
	// updates in history. We accept all of them.
	// See TestExtractTaskGroupID for examples.
	inspectorURLRegex = regexp.MustCompile(`^(https://[^/]*)/task-group-inspector/#/([^/]*)`)
	taskURLRegex      = regexp.MustCompile(`^(https://[^/]*)(?:/tasks)?/groups/([^/]*)(?:/tasks/([^/]*))?`)
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

// API is an interface for Taskcluster related methods.
type API interface {
	GetTaskGroupInfo(string, string) (*TaskGroupInfo, error)
}

type apiImpl struct{}

// GetEventInfo turns a StatusEventPayload into an EventInfo struct.
func GetEventInfo(status StatusEventPayload, log shared.Logger, api API) (EventInfo, error) {
	if status.SHA == nil {
		return EventInfo{}, errors.New("No sha on taskcluster status event")
	}

	if status.TargetURL == nil {
		return EventInfo{}, errors.New("No target_url on taskcluster status event")
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

// tcStatusWebhookHandler reacts to GitHub status webhook events.
func tcStatusWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" ||
		r.Header.Get("X-GitHub-Event") != "status" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	ds := shared.NewAppEngineDatastore(ctx, false)
	secret, err := shared.GetSecret(ds, "github-tc-webhook-secret")
	if err != nil {
		http.Error(w, "Unable to verify request: secret not found", http.StatusInternalServerError)
		return
	}

	payload, err := github.ValidatePayload(r, []byte(secret))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	log := shared.GetLogger(ctx)
	log.Debugf("GitHub Delivery: %s", r.Header.Get("X-GitHub-Delivery"))

	aeAPI := shared.NewAppEngineAPI(ctx)
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

	event, err := GetEventInfo(status, log, &apiImpl{})
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

	if err == errNoResults {
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
	return
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

// taskInfo is an abstraction of a Taskcluster task, containing the necessary
// information for us to process the task in wpt.fyi.
type taskInfo struct {
	name   string
	taskID string
	state  string
}

// taskGroupInfo is an abstraction of a Taskcluster task group, containing the
// necessary information for us to process the group in wpt.fyi.
type taskGroupInfo struct {
	taskGroupID string
	tasks       []taskInfo
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

	group := TaskGroupInfo{
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

		if task.State != "completed" {
			log.Infof("Task group %s has an unfinished task: %s; %s will be ignored in this group.",
				group.TaskGroupID, id, product)
			failures.Add(product)
			continue
		}

		urls := urlsByProduct[product]
		// Generate some URLs that point directly to
		// https://docs.taskcluster.net/docs/reference/platform/queue/api#get-artifact-from-latest-run
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
			err := uploadClient.CreateRun(sha, username, password, urls.Results, urls.Screenshots, labelsForRun)
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
