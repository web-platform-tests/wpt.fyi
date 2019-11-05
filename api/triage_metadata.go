// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v28/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

//TODO: Modify these after learning about the gitclient
var (
	sourceOwner   = "kyleju"
	sourceRepo    = "Bayesian_Inference"
	commitMessage = "random commit testing from"
	commitBranch  = "commit-testing-branch123"
	baseBranch    = "master"
	prRepoOwner   = "kyleju"
	prRepo        = "Bayesian_Inference"
	prBranch      = "master"
	prSubject     = "tesing for PR"
	prDescription = "whatever"
	sourceFiles   = "distribution.py"
	authorName    = "kyle"
	authorEmail   = "kyleju@live.com"
)

type triageMetadata struct {
	ctx          context.Context
	githubClient *github.Client
	logger       shared.Logger
	httpClient   *http.Client
}

// TODO: Create a branch fresh out of master every time.
func (tm triageMetadata) getRef() (ref *github.Reference, err error) {
	client := tm.githubClient
	if ref, _, err = client.Git.GetRef(tm.ctx, sourceOwner, sourceRepo, "refs/heads/"+commitBranch); err == nil {
		return ref, nil
	}

	// We consider that an error means the branch has not been found and needs to
	// be created.
	if commitBranch == baseBranch {
		return nil, errors.New("The commit branch does not exist but `-base-branch` is the same as `-commit-branch`")
	}

	var baseRef *github.Reference
	if baseRef, _, err = client.Git.GetRef(tm.ctx, sourceOwner, sourceRepo, "refs/heads/"+baseBranch); err != nil {
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String("refs/heads/" + commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = client.Git.CreateRef(tm.ctx, sourceOwner, sourceRepo, newRef)
	return ref, err
}

// getTree generates the tree to commit based on the given files and the commit
// of the ref you got in getRef.
func (tm triageMetadata) getTree(ref *github.Reference) (tree *github.Tree, err error) {
	client := tm.githubClient
	// Create a tree with what to commit.
	entries := []github.TreeEntry{}

	// TODO: GET THE OLD FILE FROM THE TRIP OF THE TRIP AND MODIFY IT FOR METADATA -> READ a single file at a time/ or download the repository using the exising metadata.go in shared
	for _, fileArg := range strings.Split(sourceFiles, ",") {
		// TODO: override this to amend the metadata.
		file, content, err := tm.getFileContent(fileArg)
		if err != nil {
			return nil, err
		}
		fmt.Println("content")
		entries = append(entries, github.TreeEntry{Path: github.String(file), Type: github.String("blob"), Content: github.String(string(content)), Mode: github.String("100644")})
	}

	tree, _, err = client.Git.CreateTree(tm.ctx, sourceOwner, sourceRepo, *ref.Object.SHA, entries)
	return tree, err
}

// getFileContent loads the local content of a file and return the target name
// of the file in the target repository and its contents.
func (tm triageMetadata) getFileContent(fileArg string) (targetName string, b []byte, err error) {
	var localFile string
	files := strings.Split(fileArg, ":")
	switch {
	case len(files) < 1:
		return "", nil, errors.New("empty `-files` parameter")
	case len(files) == 1:
		localFile = files[0]
		targetName = files[0]
	default:
		localFile = files[0]
		targetName = files[1]
	}

	b, err = ioutil.ReadFile(localFile)
	return targetName, b, err
}

// createCommit creates the commit in the given reference using the given tree.
func (tm triageMetadata) pushCommit(ref *github.Reference, tree *github.Tree) (err error) {
	client := tm.githubClient
	// Get the parent commit to attach the commit to.
	parent, _, err := client.Repositories.GetCommit(tm.ctx, sourceOwner, sourceRepo, *ref.Object.SHA)
	if err != nil {
		return err
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	// Create the commit using the tree.
	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: &authorName, Email: &authorEmail}
	commit := &github.Commit{Author: author, Message: &commitMessage, Tree: tree, Parents: []github.Commit{*parent.Commit}}
	newCommit, _, err := client.Git.CreateCommit(tm.ctx, sourceOwner, sourceRepo, commit)
	if err != nil {
		return err
	}

	// Attach the commit to the master branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = client.Git.UpdateRef(tm.ctx, sourceOwner, sourceRepo, ref, false)
	return err
}

// createPR creates a pull request. Based on: https://godoc.org/github.com/google/go-github/github#example-PullRequestsService-Create
func (tm triageMetadata) createPR() (err error) {
	client := tm.githubClient
	if prSubject == "" {
		return errors.New("missing `-pr-title` flag; skipping PR creation")
	}

	newPR := &github.NewPullRequest{
		Title:               &prSubject,
		Head:                &commitBranch,
		Base:                &prBranch,
		Body:                &prDescription,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(tm.ctx, prRepoOwner, prRepo, newPR)
	if err != nil {
		return err
	}

	fmt.Printf("PR created: %s\n", pr.GetHTMLURL())
	return nil
}

func (tm triageMetadata) mergeToGithub(triagedMetadataMap map[string][]byte) error {
	ref, err := tm.getRef()
	if err != nil {
		log.Fatalf("Unable to get/create the commit reference: %s\n", err)
		return err
	}
	if ref == nil {
		log.Fatalf("No error where returned but the reference is nil")
	}

	tree, err := tm.getTree(ref)
	if err != nil {
		log.Fatalf("Unable to create the tree based on the provided files: %s\n", err)
		return err
	}

	if err := tm.pushCommit(ref, tree); err != nil {
		log.Fatalf("Unable to create the commit: %s\n", err)
		return err
	}

	if err := tm.createPR(); err != nil {
		log.Fatalf("Error while creating the pull request: %s", err)
		return err
	}

	return nil
}

func (tm triageMetadata) addToFiles(metadata shared.MetadataResults, filesMap map[string]shared.Metadata) (map[string][]byte, error) {

}

func (tm triageMetadata) triage(metadata shared.MetadataResults) error {
	filesMap, err := shared.GetMetadataByteMap(tm.httpClient, tm.logger, shared.MetadataArchiveURL)
	if err != nil {
		return err
	}

	triagedMetadataMap, err := tm.addToFiles(metadata, filesMap)
	if err != nil {
		return err
	}

	return tm.mergeToGithub(triagedMetadataMap)
}
