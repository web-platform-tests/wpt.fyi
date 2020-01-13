// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/triage_metadata_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared TriageMetadataInterface

package shared

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v28/github"
	"gopkg.in/yaml.v2"
)

// TriageMetadataInterface encapsulates the Triage() method for testing.
type TriageMetadataInterface interface {
	Triage(metadata MetadataResults) (string, error)
}

// triageMetadata encapsulates all dependencies for the Triage() method.
type triageMetadata struct {
	ctx context.Context
	MetadataGithub
	logger     Logger
	httpClient *http.Client
}

// MetadataGithub encapsulates all Github Information for the createWPTMetadataPR() method.
type MetadataGithub struct {
	githubClient *github.Client
	authorName   string
	authorEmail  string
	wptmetadataGitHubInfo
}

// wptmetadataGitHubInfo encapsulates all static Github Information for the createWPTMetadataPR() method.
type wptmetadataGitHubInfo struct {
	sourceOwner   string
	sourceRepo    string
	commitMessage string
	commitBranch  string
	baseBranch    string
	prRepoOwner   string
	prRepo        string
	prBranch      string
	prSubject     string
	prDescription string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getNewCommitBranchName(ctx context.Context, client *github.Client, sourceOwner string, sourceRepo string) string {
	commitBranch := "auto-triage-branch" + generateRandomInt()
	bound := 0
	// If the commitBranch name already exists, generate a new one. We consider an error to mean that the branch
	// has not been found and needs to be created.
	for ref, _, err := client.Git.GetRef(ctx, sourceOwner, sourceRepo, "refs/heads/"+commitBranch); err == nil && ref != nil; bound++ {
		// This loop will rarely run more than 10 times because only a handful of random PR branches should exist at any time.
		if bound >= 10 {
			break
		}
		commitBranch = "auto-triage-branch" + generateRandomInt()
	}

	return commitBranch
}

func getWptmetadataGitHubInfo(ctx context.Context, client *github.Client) wptmetadataGitHubInfo {
	sourceOwner := "web-platform-tests"
	sourceRepo := "wpt-metadata"
	baseBranch := "master"
	commitBranch := getNewCommitBranchName(ctx, client, sourceOwner, sourceRepo)

	return wptmetadataGitHubInfo{
		sourceOwner:   sourceOwner,
		sourceRepo:    sourceRepo,
		commitMessage: "Commit New Metadata",
		commitBranch:  commitBranch,
		baseBranch:    baseBranch,
		prRepoOwner:   sourceOwner,
		prRepo:        sourceRepo,
		prBranch:      baseBranch,
		prSubject:     "Automatically Triage New Metadata",
		// TODO(kyleju): create a doc describing how to use this service.
		prDescription: "PR for metadata triaged through /api/metadata/triage endpoint. See <insert a doc> for more information about how to use this service."}
}

func (tm triageMetadata) getCommitBranchRef() (ref *github.Reference, err error) {
	client := tm.githubClient
	var baseRef *github.Reference
	if baseRef, _, err = client.Git.GetRef(tm.ctx, tm.sourceOwner, tm.sourceRepo, "refs/heads/"+tm.baseBranch); err != nil {
		return nil, err
	}

	newRef := &github.Reference{Ref: github.String("refs/heads/" + tm.commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = client.Git.CreateRef(tm.ctx, tm.sourceOwner, tm.sourceRepo, newRef)
	return ref, err
}

// getTree generates a github.Tree representing the changes in triagedMetadataMap, pointing at the passed ref.
func (tm triageMetadata) getTree(ref *github.Reference, triagedMetadataMap map[string][]byte) (tree *github.Tree, err error) {
	client := tm.githubClient

	entries := []github.TreeEntry{}
	for folderPath, content := range triagedMetadataMap {
		dest := GetMetadataFilePath(folderPath)
		entries = append(entries, github.TreeEntry{Path: github.String(dest), Type: github.String("blob"), Content: github.String(string(content)), Mode: github.String("100644")})
	}

	tree, _, err = client.Git.CreateTree(tm.ctx, tm.sourceOwner, tm.sourceRepo, *ref.Object.SHA, entries)
	return tree, err
}

// pushCommit creates the commit in the given reference using the given tree.
func (tm triageMetadata) pushCommit(ref *github.Reference, tree *github.Tree) (err error) {
	client := tm.githubClient
	// Get the parent commit to attach the commit to.
	parent, _, err := client.Repositories.GetCommit(tm.ctx, tm.sourceOwner, tm.sourceRepo, *ref.Object.SHA)
	if err != nil {
		return err
	}
	parent.Commit.SHA = parent.SHA

	// Create the commit using the tree.
	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: &tm.authorName, Email: &tm.authorEmail}
	commit := &github.Commit{Author: author, Message: &tm.commitMessage, Tree: tree, Parents: []github.Commit{*parent.Commit}}
	newCommit, _, err := client.Git.CreateCommit(tm.ctx, tm.sourceOwner, tm.sourceRepo, commit)
	if err != nil {
		return err
	}

	ref.Object.SHA = newCommit.SHA
	_, _, err = client.Git.UpdateRef(tm.ctx, tm.sourceOwner, tm.sourceRepo, ref, false)
	return err
}

// createPR creates a pull request from the commit branch (with the new triage changes) to the
// master branch of the repository.
// Based on: https://godoc.org/github.com/google/go-github/github#example-PullRequestsService-Create
func (tm triageMetadata) createPR() (string, error) {
	newPR := &github.NewPullRequest{
		Title:               &tm.prSubject,
		Head:                &tm.commitBranch,
		Base:                &tm.prBranch,
		Body:                &tm.prDescription,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := tm.githubClient.PullRequests.Create(tm.ctx, tm.prRepoOwner, tm.prRepo, newPR)
	if err != nil {
		return "", err
	}

	tm.logger.Infof("PR created: %s", pr.GetHTMLURL())
	return pr.GetHTMLURL(), nil
}

func (tm triageMetadata) createWPTMetadataPR(triagedMetadataMap map[string][]byte) (string, error) {
	log := tm.logger
	ref, err := tm.getCommitBranchRef()
	if err != nil {
		log.Errorf("Unable to get/create the commit reference: %s", err)
		return "", err
	}

	if ref == nil {
		log.Errorf("No error returned but the reference is nil")
		return "", errors.New("No error returned but the reference is nil")
	}

	tree, err := tm.getTree(ref, triagedMetadataMap)
	if err != nil {
		log.Errorf("Unable to create the tree based on the provided files: %s", err)
		return "", err
	}

	if err := tm.pushCommit(ref, tree); err != nil {
		log.Errorf("Unable to create the commit: %s", err)
		return "", err
	}
	pr, err := tm.createPR()
	if err != nil {
		log.Errorf("Error while creating the pull request: %s", err)
		return "", err
	}

	return pr, nil
}

// Add Metadata into the existing Metadata YML files and only return the modified files.
func addToFiles(metadata MetadataResults, filesMap map[string]Metadata, logger Logger) map[string][]byte {
	// Update filesMap with the new information in metadata.
	for test, links := range metadata {
		folderName, _ := SplitWPTTestPath(test)
		appendTestName(test, metadata)
		// If the META.YML does not exist in the repository.
		if _, ok := filesMap[folderName]; !ok {
			filesMap[folderName] = Metadata{Links: links}
			continue
		}

		// Folder already exists.
		for _, link := range links {
			existingMetadata := filesMap[folderName]
			hasMerged := false
			for index, existingLink := range existingMetadata.Links {
				if link.URL == existingLink.URL && link.Product.MatchesProductSpec(existingLink.Product) {
					// Add new MetadataResult to the existing link.
					filesMap[folderName].Links[index].Results = append(existingMetadata.Links[index].Results, link.Results...)
					hasMerged = true
					break
				}
			}

			// Add new MetadataLink to the existing Link if no link was found.
			if !hasMerged {
				filesMap[folderName] = Metadata{Links: append(filesMap[folderName].Links, link)}
			}
		}
	}

	// Grab all newly updated metadata files.
	res := make(map[string][]byte)
	for test := range metadata {
		folderName, _ := SplitWPTTestPath(test)
		metadataBytes, err := yaml.Marshal(filesMap[folderName])
		if err != nil {
			logger.Errorf("Error from marshal %s: %s", folderName, err.Error())
			continue
		}
		res[folderName] = metadataBytes
	}
	return res
}

// The metadata triage API end-point accepts metadata in a flattened JSON structure. To re-create
// the file-sharded structure of the wpt-metadata repository, we have to re-fill in the TestPath field
// for each new test added.
func appendTestName(test string, metadata MetadataResults) {
	links := metadata[test]
	_, testName := SplitWPTTestPath(test)
	for linkIndex, link := range links {
		if len(link.Results) == 0 {
			links[linkIndex].Results = make([]MetadataTestResult, 0)
			links[linkIndex].Results = append(link.Results, MetadataTestResult{TestPath: testName})
			continue
		}

		for resultIndex := range link.Results {
			metadata[test][linkIndex].Results[resultIndex].TestPath = testName
		}
	}
}

func generateRandomInt() string {
	return strconv.Itoa(rand.Intn(10000))
}

func (tm triageMetadata) Triage(metadata MetadataResults) (string, error) {
	filesMap, err := GetMetadataByteMap(tm.httpClient, tm.logger, MetadataArchiveURL)
	if err != nil {
		return "", err
	}

	triagedMetadataMap := addToFiles(metadata, filesMap, tm.logger)
	tm.MetadataGithub.wptmetadataGitHubInfo = getWptmetadataGitHubInfo(tm.ctx, tm.githubClient)
	return tm.createWPTMetadataPR(triagedMetadataMap)
}

// GetTriageMetadata returns an instance of the triageMetadata struct to run Triage() method.
func GetTriageMetadata(ctx context.Context, git MetadataGithub, logger Logger, httpClient *http.Client) TriageMetadataInterface {
	return triageMetadata{
		ctx:            ctx,
		MetadataGithub: git,
		logger:         logger,
		httpClient:     httpClient}
}

// GetMetadataGithub returns an instance of the MetadataGithub struct as a part of a triageMetadata struct.
func GetMetadataGithub(githubClient *github.Client, authorName string, authorEmail string) MetadataGithub {
	return MetadataGithub{
		githubClient: githubClient,
		authorName:   authorName,
		authorEmail:  authorEmail}
}
