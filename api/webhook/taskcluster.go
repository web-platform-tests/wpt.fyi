// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

var (
	taskNameRegex          = regexp.MustCompile(`^wpt-(.*)-(testharness|reftest|wdspec)-\d+$`)
	resultsReceiverTimeout = time.Minute
)

func tcWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" ||
		r.Header.Get("X-GitHub-Event") != "status" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineStandardContext(r)
	log := shared.GetLogger(ctx)

	payload, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	secret, err := getSecret(ctx)
	if err != nil {
		log.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !verifySignature(payload, r.Header.Get("X-Hub-Signature"), secret) {
		http.Error(w, "HMAC verification failed", http.StatusUnauthorized)
		return
	}

	log.Debugf("GitHub Delivery: %s", r.Header.Get("X-GitHub-Delivery"))

	processed, err := handleStatusEvent(ctx, payload)
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
	Sha       string       `json:"sha"`
	State     string       `json:"state"`
	Context   string       `json:"context"`
	TargetURL string       `json:"target_url"`
	Branches  []branchInfo `json:"branches"`
}

type branchInfo struct {
	Name string `json:"name"`
}

func handleStatusEvent(ctx context.Context, payload []byte) (bool, error) {
	log := shared.GetLogger(ctx)
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

	log.Infof("Processing task group %s", taskGroupID)
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

	// The default timeout is 5s, not enough for the receiver to download the reports.
	slowCtx, cancel := context.WithTimeout(ctx, resultsReceiverTimeout)
	defer cancel()
	err = createAllRuns(log, urlfetch.Client(slowCtx), api, username, password, urlsByBrowser)
	if err != nil {
		return false, err
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

func getSecret(ctx context.Context) (token string, err error) {
	var t shared.Token
	key := datastore.NewKey(ctx, "Token", "github-tc-webhook-secret", 0, nil)
	err = datastore.Get(ctx, key, &t)
	return t.Secret, err
}

func verifySignature(message []byte, signature string, secret string) bool {
	// https://developer.github.com/webhooks/securing/
	signature = strings.TrimPrefix(signature, "sha1=")
	messageMAC, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func createAllRuns(log shared.Logger, client *http.Client, api, username, password string, urlsByBrowser map[string][]string) error {
	errors := make(chan error, len(urlsByBrowser))
	var wg sync.WaitGroup
	wg.Add(len(urlsByBrowser))
	for browser, urls := range urlsByBrowser {
		go func(browser string, urls []string) {
			defer wg.Done()
			log.Infof("Reports for %s: %v", browser, urls)
			err := createRun(client, api, username, password, urls)
			if err != nil {
				errors <- err
			}
		}(browser, urls)
	}
	wg.Wait()
	close(errors)

	var errStr string
	for err := range errors {
		errStr += err.Error() + "\n"
	}
	if errStr != "" {
		return fmt.Errorf("error(s) occured when talking to %s:\n%s", api, errStr)
	}
	return nil
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
