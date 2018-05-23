// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package announcer

import (
	"errors"
	"fmt"
	"time"

	"log"

	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
	agit "github.com/web-platform-tests/wpt.fyi/revisions/git"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

const mergedPrTagPrefix = "refs/tags/merge_pr_"

var errNotAllEpochsConsumed = errors.New("Not all epochs consumed")
var errNilRepo = errors.New("Repository may not be nil")
var errVacuousEpochs = errors.New("[]epoch.Epoch slice is vacuous: contains no epochs")

// Limits defines the start-time (lower bound) and now-time (upper bound) on a commit time-bounded query for revisions.
type Limits struct {
	Start time.Time
	Now   time.Time
}

// ByCommitTimeDesc implements ordering object.Commit pointers by commit time, descending.
type ByCommitTimeDesc []*object.Commit

func (cs ByCommitTimeDesc) Len() int {
	return len(cs)
}
func (cs ByCommitTimeDesc) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}
func (cs ByCommitTimeDesc) Less(i, j int) bool {
	if cs[i] == nil {
		return false
	}
	if cs[j] == nil {
		return true
	}
	return cs[i].Committer.When.After(cs[j].Committer.When)
}

// GetErrNotAllEpochsConsumed produces the canonical error for failing to consume all input epochs.
func GetErrNotAllEpochsConsumed() error {
	return errNotAllEpochsConsumed
}

// GetErrNilRepo produces the canonical error for a nil repo value that was expected to be non-nil.
func GetErrNilRepo() error {
	return errNilRepo
}

// GetErrVacuousEpochs the canonical error for a vacuous computation over epochs; i.e., passing an empty slice of epochs which would yield an empty output.
func GetErrVacuousEpochs() error {
	return errVacuousEpochs
}

// EpochReferenceIterFactory is an interface for instantiating appropriate storer.ReferenceIter implementations for an announcer configuration.
type EpochReferenceIterFactory interface {
	GetIter(repo agit.Repository, limits Limits) (storer.ReferenceIter, error)
}

type boundedMergedPRIterFactory struct{}

func (f boundedMergedPRIterFactory) GetIter(repo agit.Repository, limits Limits) (storer.ReferenceIter, error) {
	if repo == nil {
		return nil, errNilRepo
	}

	// (1) Start with all tags                                         [tagsIter]
	// (2) Filter tags to included nothing but "merge_pr_*" tags, and  [prIter]
	//     order by descending commit time                             [prIter]
	// (3) Skip tags after basis.Now, and                              [return]
	// (4) Stop iteration when commit time is before basis.Start       [return]

	tagsIter, err := repo.Tags()
	if err != nil {
		log.Printf("ERRO: Failed to create git remote reference iter: %v", err)
		return nil, err
	}

	prIter, err := agit.NewMergedPRIter(tagsIter, repo)
	if err != nil {
		log.Printf("ERRO: Failed to create git remote reference iter: %v", err)
		return nil, err
	}

	return agit.NewStopReferenceIter(agit.NewStartReferenceIter(prIter, func(ref *plumbing.Reference) bool {
		if ref == nil {
			log.Printf("WARN: Announcer iter.StartAt(): Reference is nil; skipping...")
			return false
		}
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			log.Printf("WARN: Announcer iter.StartAt(): Error getting commit; skipping...")
			return false
		}
		return commit.Committer.When.Before(limits.Now)
	}), func(ref *plumbing.Reference) bool {
		if ref == nil {
			log.Printf("WARN: Announcer iter.StopAt(): Reference is nil; not stopping...")
		}
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			log.Printf("WARN: Announcer iter.StopAt(): Error getting commit; not stopping...")
			return false
		}
		return commit.Committer.When.Before(limits.Start)
	}), nil
}

// NewBoundedMergedPRIterFactory produces an EpochReferenceIterFactory for time-bounded commits that constitute merged PRs in the web-platform-tests repository.
func NewBoundedMergedPRIterFactory() EpochReferenceIterFactory {
	return boundedMergedPRIterFactory{}
}

// Announcer constitutes the top-level component for implementing a revisions-of-interest announcer.
type Announcer interface {
	// GetRevisions computes epochal revisions based on current local announcer state.
	GetRevisions(epochs map[epoch.Epoch]int, limits Limits) (map[epoch.Epoch][]agit.Revision, error)

	// Update applies an incremental update to announcer state; e.g., an Announcer bound to a repository may have a local clone and perform an incremental fetch.
	Update() error

	// Reset abandons current announcer state and reloads a valid initial announcer state.
	Reset() error
}

// GitRemoteAnnouncerConfig configures the git operations performed by a GitRemoteAnnouncer.
type GitRemoteAnnouncerConfig struct {
	URL        string
	RemoteName string
	BranchName string
	Depth      int
	Tags       git.TagMode
	EpochReferenceIterFactory
	agit.Git
}

type gitRemoteAnnouncer struct {
	repo agit.Repository
	cfg  *GitRemoteAnnouncerConfig
}

// NewGitRemoteAnnouncer produces an Announcer that is bound to an agit.Repository.
func NewGitRemoteAnnouncer(cfg GitRemoteAnnouncerConfig) (Announcer, error) {
	a := &gitRemoteAnnouncer{
		cfg: &cfg,
	}

	// Initialize freshness and repo according to cfg.
	err := a.Reset()
	if err == nil && a.repo == nil {
		err = errNilRepo
	}
	if err != nil {
		log.Printf("ERRO: Failed to construct git remote announcer: %v", err)
		return nil, err
	}

	return a, err
}

// GetRevisions returns as complete a list of revisions as possible given current state. It will not fetch new revisions, but it will search for newer epochal revisions than a previous invocation (if any) based on current local repository state.
func (a *gitRemoteAnnouncer) GetRevisions(epochs map[epoch.Epoch]int, limits Limits) (map[epoch.Epoch][]agit.Revision, error) {
	// Create copy of epochs; local copy will be mutated.
	es := make(map[epoch.Epoch]int)
	for e, i := range epochs {
		es[e] = i
	}

	// Initialize iterator according to config.
	iter, err := a.cfg.EpochReferenceIterFactory.GetIter(a.repo, limits)
	if err != nil {
		log.Printf("ERRO: Failed to initialize reference iterator: %v", err)
		return nil, err
	}

	revs := make(map[epoch.Epoch][]agit.Revision)
	numChanges := 0
	for e, i := range es {
		revs[e] = make([]agit.Revision, 0, i)
		numChanges += i
	}

	if numChanges == 0 {
		return nil, errVacuousEpochs
	}

	// iter presents potential revisions in reverse chronological order.
	// Scan for first epochal changes between nextTime and prevTime.
	numChangesFound := 0
	prevTime := limits.Now
	for ref, err := iter.Next(); ref != nil && err == nil; ref, err = iter.Next() {
		c, err := a.repo.CommitObject(ref.Hash())
		if err != nil {
			log.Printf("WARN: Failed to locate commit for PR tag: %s; skipping...", ref.Name())
			continue
		}
		nextTime := c.Committer.When

		// Check for epochal change against every epoch.
		for e, i := range es {
			if i == 0 {
				continue
			}
			if e.IsEpochal(nextTime, prevTime) {
				numChangesFound++
				es[e]--

				revs[e] = append(revs[e], agit.RevisionData{
					Hash:       c.Hash,
					CommitTime: nextTime,
				})

				if numChangesFound == numChanges {
					break
				}
			}
		}

		if numChangesFound == numChanges {
			break
		}
		prevTime = nextTime
	}

	// Surface error if not all epochs have a revision.
	if numChangesFound != numChanges {
		return revs, errNotAllEpochsConsumed
	}

	return revs, nil
}

// Update performs a fetch on the underlying repository. Subsequent calls to GetRevisions() will incorporate any newly fetched revisions.
func (a *gitRemoteAnnouncer) Update() (err error) {
	if a.repo == nil {
		err = GetErrNilRepo()
		log.Printf("ERRO: %v", err)
		return err
	}

	name := a.cfg.BranchName
	refSpec := config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", name, name))
	if err = a.repo.Fetch(&git.FetchOptions{
		RemoteName: a.cfg.RemoteName,
		RefSpecs:   []config.RefSpec{refSpec},
		Depth:      a.cfg.Depth,
		Tags:       a.cfg.Tags,
	}); err != nil {
		if err == git.NoErrAlreadyUpToDate {
			log.Printf("INFO: Already up-to-date")
			return nil
		}

		log.Printf("ERRO: %v", err)
		return err
	}

	return nil
}

// Reset drops reference to the current repository (if any) and performs creates a new clone according to a.cfg.
func (a *gitRemoteAnnouncer) Reset() error {
	cfg := a.cfg
	refName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", cfg.BranchName))
	repo, err := cfg.Git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           cfg.URL,
		RemoteName:    cfg.RemoteName,
		ReferenceName: refName,
		Depth:         cfg.Depth,
		Tags:          cfg.Tags,
	})
	if err != nil {
		log.Printf("ERRO: Error creating git clone: %v", err)
		return err
	}
	a.repo = repo
	return nil
}
