// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

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
const flagTaskclusterAllBranches = "taskclusterAllBranches"
const flagPendingChecks = "pendingChecks"

var (
	// The pattern is based on task names in https://github.com/web-platform-tests/wpt/blob/master/tools/ci/tc/tasks/test.yml
	taskNameRegex = regexp.MustCompile(`^wpt-([a-z_]+-[a-z]+)-([a-z]+(?:-[a-z]+)*)(?:-\d+)?$`)
	// Taskcluster has used different forms of URLs in their Check & Status
	// updates in history. We accept all of them.
	// See TestExtractTaskGroupID for examples.
	inspectorURLRegex = regexp.MustCompile(`^(https://[^/]*)/task-group-inspector/#/([^/]*)`)
	taskURLRegex      = regexp.MustCompile(`^(https://[^/]*)(?:/tasks)?/groups/([^/]*)(?:/tasks/([^/]*))?`)
)

// Non-fatal error when there is no result (e.g. nothing finishes yet).
var errNoResults = errors.New("no result URLs found in task group")

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
			rootURL, taskGroupID, taskID := parseTaskclusterURL(*status.TargetURL)
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

			return processTaskclusterBuild(aeAPI, rootURL, taskGroupID, taskID, sha, shared.ToStringSlice(labels)...)
		}()
	}

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

// https://developer.github.com/v3/activity/events/types/#statusevent
type statusEventPayload struct {
	github.StatusEvent
}

func (s statusEventPayload) IsCompleted() bool {
	return s.GetState() == "success" || s.GetState() == "failure"
}

func (s statusEventPayload) IsTaskcluster() bool {
	return s.Context != nil && (strings.HasPrefix(*s.Context, "Taskcluster") ||
		strings.HasPrefix(*s.Context, "Community-TC"))
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

func processTaskclusterBuild(aeAPI shared.AppEngineAPI, rootURL, taskGroupID, taskID string, sha string, labels ...string) (bool, error) {
	ctx := aeAPI.Context()
	log := shared.GetLogger(ctx)

	log.Debugf("Taskcluster task group %s", taskGroupID)
	if taskID != "" {
		log.Debugf("Taskcluster task %s", taskID)
	}

	taskGroup, err := getTaskGroupInfo(rootURL, taskGroupID)
	if err != nil {
		return false, err
	}

	urlsByProduct, err := extractArtifactURLs(rootURL, log, taskGroup, taskID)
	if err != nil {
		return false, err
	}

	uploader, err := aeAPI.GetUploader(uploaderName)
	if err != nil {
		log.Errorf("Failed to get uploader creds from Datastore")
		return false, err
	}

	err = createAllRuns(
		log,
		shared.NewAppEngineAPI(ctx),
		sha,
		uploader.Username,
		uploader.Password,
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

func parseTaskclusterURL(targetURL string) (rootURL, taskGroupID, taskID string) {
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

type taskGroupInfo struct {
	TaskGroupID string
	Tasks       []tcqueue.TaskDefinitionAndStatus
}

func getTaskGroupInfo(rootURL string, groupID string) (*taskGroupInfo, error) {
	queue := tcqueue.New(nil, rootURL)

	group := taskGroupInfo{
		TaskGroupID: groupID,
	}
	continuationToken := ""

	for {
		ltgr, err := queue.ListTaskGroup(groupID, continuationToken, "1000")
		if err != nil {
			return nil, err
		}

		group.Tasks = append(group.Tasks, ltgr.Tasks...)

		continuationToken = ltgr.ContinuationToken
		if continuationToken == "" {
			break
		}
	}
	return &group, nil
}

type artifactURLs struct {
	Results     []string
	Screenshots []string
}

func extractArtifactURLs(rootURL string, log shared.Logger, group *taskGroupInfo, taskID string) (
	urlsByProduct map[string]artifactURLs, err error) {
	urlsByProduct = make(map[string]artifactURLs)
	failures := mapset.NewSet()
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
			log.Infof("Ignoring unrecognized task: %s", task.Task.Metadata.Name)
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

		if task.Status.State != "completed" {
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

func createAllRuns(
	log shared.Logger,
	aeAPI shared.AppEngineAPI,
	sha,
	username,
	password string,
	urlsByProduct map[string]artifactURLs,
	labels []string) error {
	errors := make(chan error, len(urlsByProduct))
	var wg sync.WaitGroup
	wg.Add(len(urlsByProduct))
	for product, urls := range urlsByProduct {
		go func(product string, urls artifactURLs) {
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
