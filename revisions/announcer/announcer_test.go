// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package announcer_test

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/revisions/announcer"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
	agit "github.com/web-platform-tests/wpt.fyi/revisions/git"
	"github.com/web-platform-tests/wpt.fyi/revisions/test"
	billy "gopkg.in/src-d/go-billy.v4"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/storage"
)

var errFake = errors.New("Implementation is fake")

var factory = announcer.NewBoundedMergedPRIterFactory()

func TestBoundedMergedPRIterFactory_GetIter_NilRepo(t *testing.T) {
	iter, err := factory.GetIter(nil, announcer.Limits{})
	assert.True(t, iter == nil)
	assert.True(t, err == announcer.GetErrNilRepo())
}

func TestBoundedMergedPRIterFactory_GetIter_Fake(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := agit.NewMockRepository(mockCtrl)
	repo.EXPECT().Tags().Return(nil, errFake)

	iter, err := factory.GetIter(repo, announcer.Limits{})
	assert.True(t, iter == nil)
	assert.True(t, err == errFake)
}

func TestBoundedMergedPRIterFactory_GetIter(t *testing.T) {
	// Out-of-order tags 1-6; 2, 4, 5, 6 marked as PRs; iter to start just after 5's commit time, going back to (including) 4's commit time.
	iter, err := factory.GetIter(test.NewMockRepository([]test.Tag{
		test.Tag{
			TagName:    "not_a_pr_1",
			Hash:       "01",
			CommitTime: time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "merge_pr_6",
			Hash:       "06",
			CommitTime: time.Date(2018, 4, 6, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "merge_pr_5",
			Hash:       "05",
			CommitTime: time.Date(2018, 4, 5, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "not_a_pr_3",
			Hash:       "03",
			CommitTime: time.Date(2018, 4, 3, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "merge_pr_4",
			Hash:       "04",
			CommitTime: time.Date(2018, 4, 4, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "merge_pr_2",
			Hash:       "02",
			CommitTime: time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC),
		},
	}, test.NilFetchImpl), announcer.Limits{
		Now:   time.Date(2018, 4, 5, 0, 0, 0, 1, time.UTC),
		Start: time.Date(2018, 4, 4, 0, 0, 0, 0, time.UTC),
	})
	assert.True(t, err == nil)
	refNames := []string{"refs/tags/merge_pr_5", "refs/tags/merge_pr_4"}
	i := 0
	var ref *plumbing.Reference
	for ref, err = iter.Next(); ref != nil && err == nil; ref, err = iter.Next() {
		assert.True(t, ref.Name().String() == refNames[i])
		i++
	}
	assert.True(t, err == io.EOF)
}

func TestGitRemoteAnnouncer_Init_FakeGit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGit := agit.NewMockGit(mockCtrl)
	mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errFake)

	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: mockGit,
	})
	assert.True(t, a == nil)
	assert.True(t, err == errFake)
}

type NilRepoProducer struct{}

func (NilRepoProducer) CommitObject(h plumbing.Hash) (*object.Commit, error) {
	return nil, errFake
}

func (NilRepoProducer) Tags() (storer.ReferenceIter, error) {
	return nil, errFake
}

func (NilRepoProducer) Fetch(o *git.FetchOptions) error {
	return errFake
}

func (NilRepoProducer) Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (agit.Repository, error) {
	return nil, nil
}

func TestGitRemoteAnnouncer_Init_NilRepo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGit := agit.NewMockGit(mockCtrl)
	mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errFake)

	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: mockGit,
	})
	assert.True(t, a == nil)
	assert.True(t, err == errFake)
}

func TestGitRemoteAnnouncer_Init_OK(t *testing.T) {
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: test.NewMockRepository([]test.Tag{}, test.NilFetchImpl),
	})
	assert.True(t, a != nil)
	assert.True(t, err == nil)
}

type SliceReferenceIter struct {
	refs []*plumbing.Reference
	idx  int
}

func (iter *SliceReferenceIter) Next() (ref *plumbing.Reference, err error) {
	if iter.idx >= len(iter.refs) {
		err = io.EOF
	} else {
		ref = iter.refs[iter.idx]
		iter.idx++
	}
	return ref, err
}

func (iter *SliceReferenceIter) ForEach(f func(*plumbing.Reference) error) error {
	for _, ref := range iter.refs {
		err := f(ref)
		if err != nil {
			return err
		}
	}
	return nil
}

func (iter *SliceReferenceIter) Close() {
	iter.refs = make([]*plumbing.Reference, 0)
}

type SliceReferenceIterFactory struct {
	*SliceReferenceIter
}

func (f SliceReferenceIterFactory) GetIter(repo agit.Repository, limits announcer.Limits) (storer.ReferenceIter, error) {
	return f.SliceReferenceIter, nil
}

func TestGitRemoteAnnouncer_GetRevisions_ErrFake(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := test.NewMockRepository([]test.Tag{}, test.NilFetchImpl)
	mockFactory := announcer.NewMockEpochReferenceIterFactory(mockCtrl)
	mockFactory.EXPECT().GetIter(repo, gomock.Any()).Return(nil, errFake)

	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: mockFactory,
		Git: repo,
	})
	assert.True(t, a != nil)
	assert.True(t, err == nil)

	revs, err := a.GetRevisions(make(map[epoch.Epoch]int), announcer.Limits{})
	assert.True(t, revs == nil)
	assert.True(t, err == errFake)
}

func TestGitRemoteAnnouncer_GetRevisions_ErrEmptyEpochs(t *testing.T) {
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{},
		},
		Git: test.NewMockRepository([]test.Tag{}, test.NilFetchImpl),
	})
	assert.True(t, a != nil)
	assert.True(t, err == nil)

	revs, err := a.GetRevisions(make(map[epoch.Epoch]int), announcer.Limits{})
	assert.True(t, revs == nil)
	assert.True(t, err == announcer.GetErrEmptyEpochs())
}

func TestGitRemoteAnnouncer_GetRevisions_ErrNotAllEpochsConsumed(t *testing.T) {
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{},
		},
		Git: test.NewMockRepository([]test.Tag{}, test.NilFetchImpl),
	})
	assert.True(t, a != nil)
	assert.True(t, err == nil)

	epochs := make(map[epoch.Epoch]int)
	epochs[epoch.Hourly{}] = 1
	revs, err := a.GetRevisions(epochs, announcer.Limits{})
	assert.True(t, revs != nil)
	assert.True(t, err == announcer.GetErrNotAllEpochsConsumed())
}

func TestGitRemoteAnnouncer_GetRevisions_Single(t *testing.T) {
	tags := []test.Tag{
		test.Tag{
			TagName:    "one",
			Hash:       "01",
			CommitTime: time.Date(2018, 4, 1, 23, 59, 59, 999999999, time.UTC),
		},
	}
	refs := test.Tags(tags).Refs()
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{
				refs: refs,
			},
		},
		Git: test.NewMockRepository(tags, test.NilFetchImpl),
	})
	assert.True(t, a != nil)
	assert.True(t, err == nil)

	epochs := make(map[epoch.Epoch]int)
	epochs[epoch.Daily{}] = 1
	revs, err := a.GetRevisions(epochs, announcer.Limits{
		// Start of next day after tag.
		Now: time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC),
		// Time of tag.
		Start: time.Date(2018, 4, 1, 23, 59, 59, 999999999, time.UTC),
	})
	assert.True(t, revs != nil)
	assert.True(t, err == nil)
	dailyRevs, ok := revs[epoch.Daily{}]
	assert.True(t, ok)
	assert.True(t, len(dailyRevs) == 1)
	assert.True(t, dailyRevs[0] == agit.RevisionData{
		Hash:       tags[0].GetHash(),
		CommitTime: tags[0].GetCommitTime(),
	})
}

func TestGitRemoteAnnouncer_GetRevisions_MultiSameEpoch(t *testing.T) {
	tags := []test.Tag{
		test.Tag{
			TagName:    "three",
			Hash:       "03",
			CommitTime: time.Date(2018, 4, 3, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "two-two",
			Hash:       "22",
			CommitTime: time.Date(2018, 4, 2, 12, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "two",
			Hash:       "02",
			CommitTime: time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "one-one",
			Hash:       "11",
			CommitTime: time.Date(2018, 4, 1, 12, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "one",
			Hash:       "01",
			CommitTime: time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	refs := test.Tags(tags).Refs()
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{
				refs: refs,
			},
		},
		Git: test.NewMockRepository(tags, test.NilFetchImpl),
	})
	assert.True(t, a != nil)
	assert.True(t, err == nil)

	epochs := make(map[epoch.Epoch]int)
	// All three days included in commit history.
	epochs[epoch.Daily{}] = 3
	revs, err := a.GetRevisions(epochs, announcer.Limits{
		// A day after last tag.
		Now: time.Date(2018, 4, 4, 0, 0, 0, 0, time.UTC),
		// Way before first tag.
		Start: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
	})
	assert.True(t, revs != nil)
	assert.True(t, err == nil)
	dailyRevs, ok := revs[epoch.Daily{}]
	assert.True(t, ok)
	assert.True(t, len(dailyRevs) == 3)

	// Last commit from previous day chosen for each:
	// "three" [0], "two-two" [1], "one-one" [3].
	expected := [3]agit.Revision{
		agit.RevisionData{
			Hash:       tags[0].GetHash(),
			CommitTime: tags[0].GetCommitTime(),
		},
		agit.RevisionData{
			Hash:       tags[1].GetHash(),
			CommitTime: tags[1].GetCommitTime(),
		},
		agit.RevisionData{
			Hash:       tags[3].GetHash(),
			CommitTime: tags[3].GetCommitTime(),
		},
	}
	for i := 0; i < 3; i++ {
		assert.True(t, dailyRevs[i] == expected[i])
	}
}

func TestGitRemoteAnnouncer_GetRevisions_MultiEpochs(t *testing.T) {
	tags := []test.Tag{
		test.Tag{
			TagName:    "hourly",
			Hash:       "03",
			CommitTime: time.Date(2018, 4, 2, 2, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "two-hourly",
			Hash:       "02",
			CommitTime: time.Date(2018, 4, 2, 1, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "daily",
			Hash:       "01",
			CommitTime: time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	refs := test.Tags(tags).Refs()
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{
				refs: refs,
			},
		},
		Git: test.NewMockRepository(tags, test.NilFetchImpl),
	})
	assert.True(t, a != nil)
	assert.True(t, err == nil)

	epochs := make(map[epoch.Epoch]int)
	epochs[epoch.Hourly{}] = 1
	epochs[epoch.TwoHourly{}] = 1
	epochs[epoch.Daily{}] = 1
	revs, err := a.GetRevisions(epochs, announcer.Limits{
		// An hour after latest commit.
		Now: time.Date(2018, 4, 2, 3, 0, 0, 0, time.UTC),
		// Way before first tag.
		Start: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
	})
	assert.True(t, revs != nil)
	assert.True(t, err == nil)

	hourlyRevs, ok := revs[epoch.Hourly{}]
	assert.True(t, ok)
	assert.True(t, len(hourlyRevs) == 1)
	assert.True(t, hourlyRevs[0] == agit.RevisionData{
		Hash:       tags[0].GetHash(),
		CommitTime: tags[0].GetCommitTime(),
	})

	twoHourlyRevs, ok := revs[epoch.TwoHourly{}]
	assert.True(t, ok)
	assert.True(t, len(twoHourlyRevs) == 1)
	assert.True(t, twoHourlyRevs[0] == agit.RevisionData{
		Hash:       tags[1].GetHash(),
		CommitTime: tags[1].GetCommitTime(),
	})

	dailyRevs, ok := revs[epoch.Daily{}]
	assert.True(t, ok)
	assert.True(t, len(dailyRevs) == 1)
	assert.True(t, dailyRevs[0] == agit.RevisionData{
		Hash:       tags[2].GetHash(),
		CommitTime: tags[2].GetCommitTime(),
	})
}

func TestGitRemoteAnnouncer_GetRevisions_MultiMultiEpochs(t *testing.T) {
	tags := []test.Tag{
		test.Tag{
			TagName:    "hourly",
			Hash:       "03",
			CommitTime: time.Date(2018, 4, 2, 2, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "two-hourly",
			Hash:       "02",
			CommitTime: time.Date(2018, 4, 2, 1, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "daily",
			Hash:       "01",
			CommitTime: time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	refs := test.Tags(tags).Refs()
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{
				refs: refs,
			},
		},
		Git: test.NewMockRepository(tags, test.NilFetchImpl),
	})
	assert.True(t, a != nil)
	assert.True(t, err == nil)

	epochs := make(map[epoch.Epoch]int)
	epochs[epoch.Hourly{}] = 2
	epochs[epoch.TwoHourly{}] = 1
	epochs[epoch.Daily{}] = 1
	revs, err := a.GetRevisions(epochs, announcer.Limits{
		// An hour after latest commit.
		Now: time.Date(2018, 4, 2, 3, 0, 0, 0, time.UTC),
		// Way before first tag.
		Start: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
	})
	assert.True(t, revs != nil)
	assert.True(t, err == nil)

	hourlyRevs, ok := revs[epoch.Hourly{}]
	assert.True(t, ok)
	assert.True(t, len(hourlyRevs) == 2)
	assert.True(t, hourlyRevs[0] == agit.RevisionData{
		Hash:       tags[0].GetHash(),
		CommitTime: tags[0].GetCommitTime(),
	})
	assert.True(t, hourlyRevs[1] == agit.RevisionData{
		Hash:       tags[1].GetHash(),
		CommitTime: tags[1].GetCommitTime(),
	})

	twoHourlyRevs, ok := revs[epoch.TwoHourly{}]
	assert.True(t, ok)
	assert.True(t, len(twoHourlyRevs) == 1)
	assert.True(t, twoHourlyRevs[0] == agit.RevisionData{
		Hash:       tags[1].GetHash(),
		CommitTime: tags[1].GetCommitTime(),
	})

	dailyRevs, ok := revs[epoch.Daily{}]
	assert.True(t, ok)
	assert.True(t, len(dailyRevs) == 1)
	assert.True(t, dailyRevs[0] == agit.RevisionData{
		Hash:       tags[2].GetHash(),
		CommitTime: tags[2].GetCommitTime(),
	})
}

type MockRepositoryProducer struct {
	clones int
}

func (g *MockRepositoryProducer) Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (agit.Repository, error) {
	g.clones++
	return test.NewMockRepository([]test.Tag{}, test.NilFetchImpl), nil
}

// TODO(markdittmer): Should test that gitRemoteAnnouncer droppped reference to initial repository. Not possible with black box testing.
func TestGitRemoteAnnouncer_Reset(t *testing.T) {
	g := MockRepositoryProducer{}
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: &g,
	})
	assert.True(t, a != nil)
	assert.True(t, err == nil)
	prevClones := g.clones
	err = a.Reset()
	assert.True(t, err == nil)
	assert.True(t, g.clones == prevClones+1)
}

type ProxyRepository struct {
	agit.Repository
}

func (p *ProxyRepository) Set(r agit.Repository) {
	p.Repository = r
}

func (p *ProxyRepository) CommitObject(h plumbing.Hash) (*object.Commit, error) {
	return p.Repository.CommitObject(h)
}
func (p *ProxyRepository) Tags() (storer.ReferenceIter, error) {
	return p.Repository.Tags()
}

func (p *ProxyRepository) Fetch(o *git.FetchOptions) error {
	return p.Repository.Fetch(o)
}

func (p *ProxyRepository) Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (agit.Repository, error) {
	return p, nil
}

func (p *ProxyRepository) GetIter(repo agit.Repository, limits announcer.Limits) (storer.ReferenceIter, error) {
	return p.Repository.Tags()
}

func TestGitRemoteAnnouncer_Update(t *testing.T) {
	// Use a proxy to swap out mock repos on update:
	// - Start with empty aRepo,\;
	// - Swap in bRepo with one commit,\;
	// - aRepo test.FetchImpl returns error to test against.
	fetchErr := errors.New("Error returned by Fetch()")
	updatedTag := test.Tag{
		TagName:    "daily",
		Hash:       "01",
		CommitTime: time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC),
	}
	bRepo := test.NewMockRepository([]test.Tag{updatedTag}, test.NilFetchImpl)
	pValue := ProxyRepository{}
	pRepo := &pValue
	aRepo := test.NewMockRepository([]test.Tag{}, func(mr *test.MockRepository, o *git.FetchOptions) error {
		pRepo.Set(bRepo)
		return fetchErr
	})
	pRepo.Set(aRepo)
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: pRepo,
		Git: pRepo,
	})
	assert.True(t, a != nil)
	assert.True(t, err == nil)

	epochs := make(map[epoch.Epoch]int)
	epochs[epoch.Daily{}] = 1
	limits := announcer.Limits{
		// Day after tag in updated pRepo->bRepo.
		Now: time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC),
		// Long before any tags.
		Start: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
	}
	getRevisions := func() (map[epoch.Epoch][]agit.Revision, error) {
		return a.GetRevisions(epochs, limits)
	}
	revs, err := getRevisions()
	assert.True(t, revs != nil)
	assert.True(t, err == announcer.GetErrNotAllEpochsConsumed())

	err = a.Fetch()
	assert.True(t, err == fetchErr)
	revs, err = getRevisions()
	assert.True(t, revs != nil)
	assert.True(t, err == nil)
	dailyRevs, ok := revs[epoch.Daily{}]
	assert.True(t, ok)
	assert.True(t, len(dailyRevs) == 1)
	assert.True(t, dailyRevs[0] == agit.RevisionData{
		Hash:       updatedTag.GetHash(),
		CommitTime: updatedTag.GetCommitTime(),
	})
}
