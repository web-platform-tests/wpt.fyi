// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package taskcluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"google.golang.org/appengine/urlfetch"

	mapset "github.com/deckarep/golang-set"
	"github.com/google/go-github/github"
	uc "github.com/web-platform-tests/wpt.fyi/api/receiver/client"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

const flagTaskclusterAllBranches = "taskclusterAllBranches"
const flagPendingChecks = "pendingChecks"

var (
	// This should follow https://github.com/web-platform-tests/wpt/blob/master/.taskcluster.yml
	// with a notable exception that "*-stability" runs are not included at the moment.
	taskNameRegex = regexp.MustCompile(`^wpt-(\w+-\w+)-(testharness|reftest|wdspec|results|results-without-changes)(?:-\d+)?$`)
)

// tcStatusWebhookHandler reacts to GitHub status webhook events. This is juxtaposed with
// handleCheckRunEvent below, which is how we react to the (new) CheckRun implementation
// of Taskcluster.
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
	var status statusEventPayload
	if err := json.Unmarshal(payload, &status); err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	processAllBranches := aeAPI.IsFeatureEnabled(flagTaskclusterAllBranches)
	var processed bool
	if !shouldProcessStatus(log, processAllBranches, &status) {
		processed = false
	} else {
		processed, err = func() (bool, error) {
			sha := *status.SHA

			if status.TargetURL == nil {
				return false, errors.New("No target_url on taskcluster status event")
			}
			taskGroupID, taskID := extractTaskGroupID(*status.TargetURL)
			if taskGroupID == "" {
				return false, fmt.Errorf("unrecognized target_url: %s", *status.TargetURL)
			}

			log.Debugf("Taskcluster task group %s", taskGroupID)

			labels := mapset.NewSet()
			if status.IsOnMaster() {
				labels.Add(shared.MasterLabel)
			}
			sender := status.GetCommit().GetAuthor().GetLogin()
			if sender != "" {
				labels.Add(shared.GetUserLabel(sender))
			}

			return processTaskclusterBuild(aeAPI, taskGroupID, taskID, sha, shared.ToStringSlice(labels)...)
		}()
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

// https://developer.github.com/v3/activity/events/types/#statusevent
type statusEventPayload struct {
	github.StatusEvent
}

func (s statusEventPayload) IsCompleted() bool {
	return s.GetState() == "success" || s.GetState() == "failure"
}

func (s statusEventPayload) IsTaskcluster() bool {
	return s.Context != nil && strings.HasPrefix(*s.Context, "Taskcluster")
}

func (s statusEventPayload) IsOnMaster() bool {
	for _, branch := range s.Branches {
		if branch.Name != nil && *branch.Name == "master" {
			return true
		}
	}
	return false
}

func (s statusEventPayload) HeadingBranches() branchInfos {
	var branches branchInfos
	for _, branch := range s.Branches {
		if *branch.Commit.SHA == *s.SHA {
			branches = append(branches, branch)
		}
	}
	return branches
}

type branchInfos []*github.Branch

func (b branchInfos) GetNames() []string {
	names := make([]string, len(b))
	for i := range b {
		names[i] = *b[i].Name
	}
	return names
}

func processTaskclusterBuild(aeAPI shared.AppEngineAPI, taskGroupID, taskID string, sha string, labels ...string) (bool, error) {
	ctx := aeAPI.Context()
	log := shared.GetLogger(ctx)
	log.Debugf("Taskcluster task group %s", taskGroupID)
	if taskID != "" {
		log.Debugf("Taskcluster task %s", taskID)
	}

	client := urlfetch.Client(ctx)
	taskGroup, err := getTaskGroupInfo(client, taskGroupID)
	if err != nil {
		return false, err
	}

	urlsByProduct, err := extractResultURLs(log, taskGroup, taskID)
	if err != nil {
		return false, err
	}

	username, password, err := getAuth(ctx)
	if err != nil {
		return false, err
	}

	err = createAllRuns(
		log,
		shared.NewAppEngineAPI(ctx),
		sha,
		username,
		password,
		urlsByProduct,
		labels)
	if err != nil {
		return false, err
	}

	return true, nil
}

func shouldProcessStatus(log shared.Logger, processAllBranches bool, status *statusEventPayload) bool {
	if !status.IsCompleted() {
		log.Debugf("Ignoring status: %s", status.GetState())
		return false
	} else if !status.IsTaskcluster() {
		log.Debugf("Ignoring non-Taskcluster context: %s", status.GetContext())
		return false
	} else if !processAllBranches && !status.IsOnMaster() {
		log.Debugf("Ignoring non-master status event")
		return false
	}
	return true
}

func extractTaskGroupID(targetURL string) (string, string) {
	inspectorRegex := regexp.MustCompile("/task-group-inspector/#/([^/]*)")
	matches := inspectorRegex.FindStringSubmatch(targetURL)
	if len(matches) > 1 {
		return matches[1], ""
	}
	taskRegex := regexp.MustCompile("/groups/([^/]*)/tasks/([^/]*)")
	matches = taskRegex.FindStringSubmatch(targetURL)
	if len(matches) > 2 {
		return matches[1], matches[2]
	}
	return "", ""
}

// https://docs.taskcluster.net/docs/reference/platform/taskcluster-queue/references/api#response-2
type taskGroupInfo struct {
	TaskGroupID string     `json:"taskGroupId"`
	Tasks       []taskInfo `json:"tasks"`
}

type taskInfo struct {
	Status struct {
		TaskID string `json:"taskId"`
		State  string `json:"state"`
	} `json:"status"`
	Task struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
	} `json:"task"`
}

func getTaskGroupInfo(client *http.Client, groupID string) (*taskGroupInfo, error) {
	// https://docs.taskcluster.net/docs/reference/platform/taskcluster-queue/references/api#list-task-group
	taskgroupURL := fmt.Sprintf("https://queue.taskcluster.net/v1/task-group/%s/list", groupID)
	resp, err := client.Get(taskgroupURL)
	if err != nil {
		return nil, err
	}
	payload, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var group taskGroupInfo
	if err := json.Unmarshal(payload, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

func extractResultURLs(log shared.Logger, group *taskGroupInfo, taskID string) (map[string][]string, error) {
	failures := mapset.NewSet()
	resultURLs := make(map[string][]string)
	for _, task := range group.Tasks {
		id := task.Status.TaskID
		if id == "" {
			return nil, fmt.Errorf("task group %s has a task without taskId", group.TaskGroupID)
		} else if taskID != "" && taskID != id {
			log.Debugf("Skipping task %s", id)
			continue
		}

		matches := taskNameRegex.FindStringSubmatch(task.Task.Metadata.Name)
		if len(matches) != 3 { // full match, browser-channel, test type
			log.Debugf("Ignoring unrecognized task: %s", task.Task.Metadata.Name)
			continue
		}
		product := matches[1]
		switch matches[2] {
		case "results":
			product += "-" + shared.PRHeadLabel
		case "results-without-changes":
			product += "-" + shared.PRBaseLabel
		}

		if task.Status.State != "completed" {
			log.Errorf("Task group %s has an unfinished task: %s; %s will be ignored in this group.",
				group.TaskGroupID, id, product)
			failures.Add(product)
			continue
		}

		resultURLs[product] = append(resultURLs[product],
			// https://docs.taskcluster.net/docs/reference/platform/taskcluster-queue/references/api#get-artifact-from-latest-run
			fmt.Sprintf(
				"https://queue.taskcluster.net/v1/task/%s/artifacts/public/results/wpt_report.json.gz", id,
			))
	}

	for failure := range failures.Iter() {
		delete(resultURLs, failure.(string))
	}

	if len(resultURLs) == 0 {
		return nil, fmt.Errorf("no result URLs found in task group")
	}
	return resultURLs, nil
}

func getAuth(ctx context.Context) (username string, password string, err error) {
	uploader, err := shared.NewAppEngineAPI(ctx).GetUploader("taskcluster")
	return uploader.Username, uploader.Password, err
}

func createAllRuns(
	log shared.Logger,
	aeAPI shared.AppEngineAPI,
	sha,
	username,
	password string,
	urlsByProduct map[string][]string,
	labels []string) error {
	errors := make(chan error, len(urlsByProduct))
	var wg sync.WaitGroup
	wg.Add(len(urlsByProduct))
	for product, urls := range urlsByProduct {
		go func(product string, urls []string) {
			defer wg.Done()
			log.Infof("Reports for %s: %v", product, urls)

			// chrome-dev-pr_head => [chrome, dev, pr_head]
			bits := strings.Split(product, "-")
			labelsForRun := labels
			switch lastBit := bits[len(bits)-1]; lastBit {
			case shared.PRBaseLabel, shared.PRHeadLabel:
				labelsForRun = append(labelsForRun, lastBit)
			}

			uploadClient := uc.NewClient(aeAPI)
			err := uploadClient.CreateRun(sha, username, password, urls, labelsForRun)
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
