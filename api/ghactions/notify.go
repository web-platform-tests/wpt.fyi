// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ghactions

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	mapset "github.com/deckarep/golang-set"
	"github.com/gobwas/glob"
	"github.com/google/go-github/v74/github"
	uc "github.com/web-platform-tests/wpt.fyi/api/receiver/client"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

const uploaderName = "github-actions"

var (
	prHeadRegex        = regexp.MustCompile(`\baffected-tests$`)
	prBaseRegex        = regexp.MustCompile(`\baffected-tests-without-changes$`)
	epochBranchesRegex = regexp.MustCompile("^epochs/.*")
)

func notifyHandler(w http.ResponseWriter, r *http.Request) {
	rawRunID := r.FormValue("run_id")
	var runID int64
	var err error
	if runID, err = strconv.ParseInt(rawRunID, 0, 0); err != nil {
		http.Error(w, fmt.Sprintf("Invalid run id: %s", rawRunID), http.StatusBadRequest)

		return
	}

	owner := r.FormValue("owner")
	repo := r.FormValue("repo")

	if owner != shared.WPTRepoOwner || repo != shared.WPTRepoName {
		http.Error(w, fmt.Sprintf("Invalid repo: %s/%s", owner, repo), http.StatusBadRequest)

		return
	}

	artifactName := r.FormValue("artifact_name")
	artifactNameGlob, err := glob.Compile(artifactName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid artifact name: %s", artifactName), http.StatusBadRequest)

		return
	}

	ctx := r.Context()
	aeAPI := shared.NewAppEngineAPI(ctx)
	log := shared.GetLogger(ctx)

	ghClient, err := aeAPI.GetGitHubClient()
	if err != nil {
		log.Errorf("Failed to get GitHub client: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	processed, err := processBuild(
		ctx,
		aeAPI,
		ghClient,
		owner,
		repo,
		runID,
		artifactNameGlob,
	)

	if err != nil {
		log.Errorf("%v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if processed {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "GitHub Actions workflow run artifacts retrieved successfully")
	} else {
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintln(w, "Notification of workflow run artifacts was ignored")
	}
}

func processBuild(
	ctx context.Context,
	aeAPI shared.AppEngineAPI,
	ghClient *github.Client,
	owner string,
	repo string,
	runID int64,
	artifactNameGlob glob.Glob,
) (bool, error) {
	log := shared.GetLogger(ctx)

	workflowRun, _, err := ghClient.Actions.GetWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return false, err
	}

	// nolint:exhaustruct // TODO: Fix exhaustruct lint error.
	opts := &github.ListOptions{PerPage: 100}

	archiveURLs := []string{}

	var labels mapset.Set

	for {
		artifacts, resp, err := ghClient.Actions.ListWorkflowRunArtifacts(ctx, owner, repo, runID, opts)

		if err != nil {
			return false, err
		}

		for _, artifact := range artifacts.Artifacts {

			if !artifactNameGlob.Match(*artifact.Name) {
				log.Infof("Skipping artifact %s", *artifact.Name)

				continue
			}

			log.Infof("Adding %s for %s/%s run %v to upload...", *artifact.Name, owner, repo, runID)

			// Set the labels based on the first artifact we find.
			if len(archiveURLs) == 0 {
				labels = chooseLabels(workflowRun, *artifact.Name, owner, repo)
			}

			archiveURLs = append(archiveURLs, *artifact.ArchiveDownloadURL)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	var sha string
	if *workflowRun.Event == "pull_request" {
		sha = *workflowRun.HeadSHA
	}

	uploader, err := aeAPI.GetUploader(uploaderName)
	if err != nil {
		return false, fmt.Errorf("failed to get uploader creds from Datastore: %w", err)
	}

	uploadClient := uc.NewClient(aeAPI)
	err = uploadClient.CreateRun(
		sha,
		uploader.Username,
		uploader.Password,
		nil,
		nil,
		archiveURLs,
		shared.ToStringSlice(labels))

	if err != nil {
		return false, fmt.Errorf("failed to create run: %w", err)
	}

	return true, nil
}

func chooseLabels( // nolint:ireturn // TODO: Fix ireturn lint error
	workflowRun *github.WorkflowRun,
	artifactName string,
	owner string,
	repo string,
) mapset.Set {
	labels := mapset.NewSet()

	// We don't actually check the event here, provided it meets
	// the criteria to be a run on master.
	if (*workflowRun.HeadRepository.Owner.Login == owner &&
		*workflowRun.HeadRepository.Name == repo) &&
		(*workflowRun.HeadBranch == "master" ||
			epochBranchesRegex.MatchString(*workflowRun.HeadBranch)) {
		labels.Add(shared.MasterLabel)
	} else if *workflowRun.Event == "pull_request" {
		if prHeadRegex.MatchString(artifactName) {
			labels.Add(shared.PRHeadLabel)
		} else if prBaseRegex.MatchString(artifactName) {
			labels.Add(shared.PRBaseLabel)
		}
	}

	return labels
}
