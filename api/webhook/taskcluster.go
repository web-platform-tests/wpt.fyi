// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

var (
	taskNameRegex = regexp.MustCompile(`^wpt-(.*)-(testharness|reftest|wdspec)-\d+$`)
)

func tcWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" ||
		r.Header.Get("X-GitHub-Event") != "status" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)
	processed, err := handleStatusEvent(ctx, r.Body)
	if err != nil {
		log.Errorf(ctx, "%v", err)
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
	Sha       string       `json:"sha"`
	State     string       `json:"state"`
	Context   string       `json:"context"`
	TargetURL string       `json:"target_url"`
	Branches  []branchInfo `json:"branches"`
}

type branchInfo struct {
	Name string `json:"name"`
}

func handleStatusEvent(ctx context.Context, body io.ReadCloser) (bool, error) {
	payload, err := ioutil.ReadAll(body)
	body.Close()
	if err != nil {
		return false, err
	}

	var status statusEventPayload
	if err := json.Unmarshal(payload, &status); err != nil {
		return false, err
	}

	if !shouldProcessStatus(&status) {
		return false, nil
	}

	taskGroupID := extractTaskGroupID(status.TargetURL)
	if taskGroupID == "" {
		return false, fmt.Errorf("unrecognized target_url: %s", status.TargetURL)
	}

	log.Infof(ctx, "Processing task group %s", taskGroupID)
	client := urlfetch.Client(ctx)
	taskGroup, err := getTaskGroupInfo(client, taskGroupID)
	if err != nil {
		return false, err
	}

	urlsByBrowser, err := extractResultURLs(taskGroup)
	if err != nil {
		return false, err
	}

	username, password, err := getAuth(ctx)
	if err != nil {
		return false, err
	}

	// https://github.com/web-platform-tests/wpt.fyi/blob/master/api/README.md#results-creation
	api := fmt.Sprintf("https://%s/api/results/upload", appengine.DefaultVersionHostname(ctx))
	for browser, urls := range urlsByBrowser {
		log.Infof(ctx, "Reports for %s: %v", browser, urls)
		// Set timeout to 1 min (the default is 5s) to give the
		// receiver enough time to download the reports.
		slowCtx, cancel := context.WithTimeout(ctx, time.Minute)
		client := urlfetch.Client(slowCtx)
		err := createRun(client, api, username, password, urls)
		cancel()
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func shouldProcessStatus(status *statusEventPayload) bool {
	if status.State == "success" &&
		strings.HasPrefix(status.Context, "Taskcluster") {
		// Only process the master branch.
		for _, branch := range status.Branches {
			if branch.Name == "master" {
				return true
			}
		}
	}
	return false
}

func extractTaskGroupID(targetURL string) string {
	lastSlash := strings.LastIndex(targetURL, "/")
	if lastSlash == -1 {
		return ""
	}
	return targetURL[lastSlash+1:]
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

func extractResultURLs(group *taskGroupInfo) (map[string][]string, error) {
	resultURLs := make(map[string][]string)
	for _, task := range group.Tasks {
		taskID := task.Status.TaskID
		if taskID == "" {
			return nil, fmt.Errorf("task group %s has a task without taskId", group.TaskGroupID)
		}
		if task.Status.State != "completed" {
			return nil, fmt.Errorf("task group %s has an unfinished task: %s", group.TaskGroupID, taskID)
		}

		matches := taskNameRegex.FindStringSubmatch(task.Task.Metadata.Name)
		if len(matches) != 3 { // full match, browser, test type
			return nil, fmt.Errorf("error parsing the name of task %s: %s", taskID, task.Task.Metadata.Name)
		}
		browser := matches[1]

		resultURLs[browser] = append(resultURLs[browser],
			// https://docs.taskcluster.net/docs/reference/platform/taskcluster-queue/references/api#get-artifact-from-latest-run
			fmt.Sprintf(
				"https://queue.taskcluster.net/v1/task/%s/artifacts/public/results/wpt_report.json.gz", taskID,
			))
	}

	if len(resultURLs) == 0 {
		return nil, fmt.Errorf("no result URLs found in task group")
	}
	return resultURLs, nil
}

func getAuth(ctx context.Context) (username string, password string, err error) {
	var u shared.Uploader
	key := datastore.NewKey(ctx, "Uploader", "taskcluster", 0, nil)
	err = datastore.Get(ctx, key, &u)
	return u.Username, u.Password, err
}

func createRun(client *http.Client, api string, username string, password string, reportURLs []string) error {
	// https://github.com/web-platform-tests/wpt.fyi/blob/master/api/README.md#url-payload
	payload := make(url.Values)
	for _, url := range reportURLs {
		payload.Add("result_url", url)
	}

	req, err := http.NewRequest("POST", api, strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", string(respBody))
	}

	return nil
}
