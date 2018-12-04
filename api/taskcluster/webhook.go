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
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"

	mapset "github.com/deckarep/golang-set"
	"github.com/google/go-github/github"
	"github.com/web-platform-tests/wpt.fyi/api/checks"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

const flagTaskclusterAllBranches = "taskclusterAllBranches"

var (
	taskNameRegex          = regexp.MustCompile(`^wpt-(\w+)-(\w+)-(testharness|reftest|wdspec|results)(?:-\d+)?$`)
	resultsReceiverTimeout = time.Minute
)

func tcWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" ||
		r.Header.Get("X-GitHub-Event") != "status" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	secret, err := shared.GetSecret(ctx, "github-tc-webhook-secret")
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

func handleStatusEvent(ctx context.Context, payload []byte) (bool, error) {
	log := shared.GetLogger(ctx)
	var status statusEventPayload
	if err := json.Unmarshal(payload, &status); err != nil {
		return false, err
	}

	if status.TargetURL == nil {
		return false, errors.New("No target_url on taskcluster status event")
	}
	taskGroupID := extractTaskGroupID(*status.TargetURL)
	if taskGroupID == "" {
		return false, fmt.Errorf("unrecognized target_url: %s", *status.TargetURL)
	}

	log.Debugf("Taskcluster task group %s", taskGroupID)
	processAllBranches := shared.IsFeatureEnabled(ctx, flagTaskclusterAllBranches)
	if !shouldProcessStatus(log, processAllBranches, &status) {
		return false, nil
	}

	client := urlfetch.Client(ctx)
	taskGroup, err := getTaskGroupInfo(client, taskGroupID)
	if err != nil {
		return false, err
	}

	urlsByBrowser, err := extractResultURLs(log, taskGroup)
	if err != nil {
		return false, err
	}

	username, password, err := getAuth(ctx)
	if err != nil {
		return false, err
	}

	// https://github.com/web-platform-tests/wpt.fyi/blob/master/api/README.md#results-creation
	uploadURL := fmt.Sprintf("https://%s/api/results/upload", appengine.DefaultVersionHostname(ctx))

	// The default timeout is 5s, not enough for the receiver to download the reports.
	slowCtx, cancel := context.WithTimeout(ctx, resultsReceiverTimeout)
	defer cancel()
	var labels []string
	if status.IsOnMaster() {
		labels = []string{"master"}
	}
	checksAPI := checks.NewAPI(ctx)
	err = createAllRuns(
		log,
		urlfetch.Client(slowCtx),
		shared.NewAppEngineAPI(ctx),
		checksAPI,
		uploadURL,
		*status.SHA,
		username,
		password,
		urlsByBrowser,
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

func extractResultURLs(log shared.Logger, group *taskGroupInfo) (map[string][]string, error) {
	failures := mapset.NewSet()
	resultURLs := make(map[string][]string)
	for _, task := range group.Tasks {
		taskID := task.Status.TaskID
		if taskID == "" {
			return nil, fmt.Errorf("task group %s has a task without taskId", group.TaskGroupID)
		}

		matches := taskNameRegex.FindStringSubmatch(task.Task.Metadata.Name)
		if len(matches) != 4 { // full match, browser, channel, test type
			log.Debugf("Ignoring unrecognized task: %s", task.Task.Metadata.Name)
			continue
		}
		browser := fmt.Sprintf("%s-%s", matches[1], matches[2])

		if task.Status.State != "completed" {
			log.Errorf("Task group %s has an unfinished task: %s; %s will be ignored in this group.",
				group.TaskGroupID, taskID, browser)
			failures.Add(browser)
			continue
		}

		resultURLs[browser] = append(resultURLs[browser],
			// https://docs.taskcluster.net/docs/reference/platform/taskcluster-queue/references/api#get-artifact-from-latest-run
			fmt.Sprintf(
				"https://queue.taskcluster.net/v1/task/%s/artifacts/public/results/wpt_report.json.gz", taskID,
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
	var u shared.Uploader
	key := datastore.NewKey(ctx, "Uploader", "taskcluster", 0, nil)
	err = datastore.Get(ctx, key, &u)
	return u.Username, u.Password, err
}

func createAllRuns(
	log shared.Logger,
	client *http.Client,
	aeAPI shared.AppEngineAPI,
	checksAPI checks.API,
	uploadURL,
	sha,
	username,
	password string,
	urlsByBrowser map[string][]string,
	labels []string) error {
	errors := make(chan error, len(urlsByBrowser))
	var wg sync.WaitGroup
	wg.Add(len(urlsByBrowser))
	suites, _ := checksAPI.GetSuitesForSHA(sha)
	for browser, urls := range urlsByBrowser {
		go func(browser string, urls []string) {
			defer wg.Done()
			log.Infof("Reports for %s: %v", browser, urls)
			err := createRun(log, client, aeAPI, sha, uploadURL, username, password, urls, labels)
			if err != nil {
				errors <- err
			} else {
				spec := shared.ProductSpec{}
				bits := strings.Split(browser, "-") // chrome-dev => [chrome, dev]
				spec.BrowserName = bits[0]
				if len(bits) > 1 {
					if label := shared.ProductChannelToLabel(bits[1]); label != "" {
						spec.Labels = mapset.NewSetWith(label)
					}
				}
				for _, suite := range suites {
					checksAPI.PendingCheckRun(suite, spec)
				}
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
		return fmt.Errorf("error(s) occured when talking to %s:\n%s", uploadURL, errStr)
	}
	return nil
}

func createRun(
	log shared.Logger,
	client *http.Client,
	aeAPI shared.AppEngineAPI,
	sha,
	api string,
	username string,
	password string,
	reportURLs []string,
	labels []string) error {
	// https://github.com/web-platform-tests/wpt.fyi/blob/master/api/README.md#url-payload
	payload := make(url.Values)
	// Not to be confused with `revision` in the wpt.fyi TestRun model, this
	// parameter is the full revision hash.
	payload.Add("revision", sha)
	for _, url := range reportURLs {
		payload.Add("result_url", url)
	}
	if labels != nil {
		payload.Add("labels", strings.Join(labels, ","))
	}
	// Ensure we call back to this appengine version instance.
	host := aeAPI.GetHostname()
	payload.Add("callback_url", fmt.Sprintf("https://%s/api/results/create", host))

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
