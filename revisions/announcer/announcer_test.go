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
	assert.Nil(t, iter)
	assert.Equal(t, announcer.GetErrNilRepo(), err)
}

func TestBoundedMergedPRIterFactory_GetIter_Fake(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := agit.NewMockRepository(mockCtrl)
	repo.EXPECT().Tags().Return(nil, errFake)

	iter, err := factory.GetIter(repo, announcer.Limits{})
	assert.Nil(t, iter)
	assert.Equal(t, errFake, err)
}

func TestBoundedMergedPRIterFactory_GetIter(t *testing.T) {
	// Out-of-order tags 1-6; 2, 4, 5, 6 marked as PRs; iter to start just after 5's commit time, going back to (including) 4's commit time.
	iter, err := factory.GetIter(test.NewMockRepository(t, []test.Tag{
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
		At:    time.Date(2018, 4, 5, 0, 0, 0, 1, time.UTC),
		Start: time.Date(2018, 4, 4, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	refNames := []string{"refs/tags/merge_pr_5", "refs/tags/merge_pr_4"}
	i := 0
	var ref *plumbing.Reference
	for ref, err = iter.Next(); ref != nil && err == nil; ref, err = iter.Next() {
		assert.Equal(t, refNames[i], ref.Name().String())
		i++
	}
	assert.Equal(t, io.EOF, err)
}

func TestGitRemoteAnnouncer_Init_FakeGit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGit := agit.NewMockGit(mockCtrl)
	mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errFake)

	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: mockGit,
	})
	assert.Nil(t, a)
	assert.Equal(t, errFake, err)
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
	assert.Nil(t, a)
	assert.Equal(t, errFake, err)
}

func TestGitRemoteAnnouncer_Init_OK(t *testing.T) {
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: test.NewMockRepository(t, []test.Tag{}, test.NilFetchImpl),
	})
	assert.NotNil(t, a)
	assert.Nil(t, err)
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

	repo := test.NewMockRepository(t, []test.Tag{}, test.NilFetchImpl)
	mockFactory := announcer.NewMockEpochReferenceIterFactory(mockCtrl)
	mockFactory.EXPECT().GetIter(repo, gomock.Any()).Return(nil, errFake)

	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: mockFactory,
		Git:                       repo,
	})
	assert.NotNil(t, a)
	assert.Nil(t, err)

	revs, err := a.GetRevisions(make(map[epoch.Epoch]int), announcer.Limits{})
	assert.Nil(t, revs)
	assert.Equal(t, errFake, err)
}

func TestGitRemoteAnnouncer_GetRevisions_ErrEmptyEpochs(t *testing.T) {
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{},
		},
		Git: test.NewMockRepository(t, []test.Tag{}, test.NilFetchImpl),
	})
	assert.NotNil(t, a)
	assert.Nil(t, err)

	revs, err := a.GetRevisions(make(map[epoch.Epoch]int), announcer.Limits{})
	assert.Nil(t, revs)
	assert.Equal(t, announcer.GetErrEmptyEpochs(), err)
}

func TestGitRemoteAnnouncer_GetRevisions_ErrNotAllEpochsConsumed(t *testing.T) {
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{},
		},
		Git: test.NewMockRepository(t, []test.Tag{}, test.NilFetchImpl),
	})
	assert.NotNil(t, a)
	assert.Nil(t, err)

	epochs := make(map[epoch.Epoch]int)
	epochs[epoch.Hourly{}] = 1
	revs, err := a.GetRevisions(epochs, announcer.Limits{})
	assert.NotNil(t, revs)
	assert.Equal(t, announcer.GetErrNotAllEpochsConsumed(), err)
}

func TestGitRemoteAnnouncer_GetRevisions_Single(t *testing.T) {
	tags := []test.Tag{
		test.Tag{
			TagName:    "one",
			Hash:       "01",
			CommitTime: time.Date(2018, 4, 1, 23, 59, 59, 999999999, time.UTC),
		},
	}
	refs := test.Tags(tags).Refs(t)
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{
				refs: refs,
			},
		},
		Git: test.NewMockRepository(t, tags, test.NilFetchImpl),
	})
	assert.NotNil(t, a)
	assert.Nil(t, err)

	epochs := make(map[epoch.Epoch]int)
	epochs[epoch.Daily{}] = 1
	revs, err := a.GetRevisions(epochs, announcer.Limits{
		// Start of next day after tag.
		At: time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC),
		// Time of tag.
		Start: time.Date(2018, 4, 1, 23, 59, 59, 999999999, time.UTC),
	})
	assert.NotNil(t, revs)
	assert.Nil(t, err)
	dailyRevs, ok := revs[epoch.Daily{}]
	assert.True(t, ok)
	assert.Equal(t, 1, len(dailyRevs))
	assert.Equal(t, agit.RevisionData{
		Hash:       tags[0].GetHash(t),
		CommitTime: tags[0].GetCommitTime(),
	}, dailyRevs[0])
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
	refs := test.Tags(tags).Refs(t)
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{
				refs: refs,
			},
		},
		Git: test.NewMockRepository(t, tags, test.NilFetchImpl),
	})
	assert.NotNil(t, a)
	assert.Nil(t, err)

	epochs := make(map[epoch.Epoch]int)
	// All three days included in commit history.
	epochs[epoch.Daily{}] = 3
	revs, err := a.GetRevisions(epochs, announcer.Limits{
		// A day after last tag.
		At: time.Date(2018, 4, 4, 0, 0, 0, 0, time.UTC),
		// Way before first tag.
		Start: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
	})
	assert.NotNil(t, revs)
	assert.Nil(t, err)
	dailyRevs, ok := revs[epoch.Daily{}]
	assert.True(t, ok)
	assert.Equal(t, 3, len(dailyRevs))

	// Last commit from previous day chosen for each:
	// "three" [0], "two-two" [1], "one-one" [3].
	expected := [3]agit.Revision{
		agit.RevisionData{
			Hash:       tags[0].GetHash(t),
			CommitTime: tags[0].GetCommitTime(),
		},
		agit.RevisionData{
			Hash:       tags[1].GetHash(t),
			CommitTime: tags[1].GetCommitTime(),
		},
		agit.RevisionData{
			Hash:       tags[3].GetHash(t),
			CommitTime: tags[3].GetCommitTime(),
		},
	}
	for i := 0; i < 3; i++ {
		assert.Equal(t, expected[i], dailyRevs[i])
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
	refs := test.Tags(tags).Refs(t)
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{
				refs: refs,
			},
		},
		Git: test.NewMockRepository(t, tags, test.NilFetchImpl),
	})
	assert.NotNil(t, a)
	assert.Nil(t, err)

	epochs := make(map[epoch.Epoch]int)
	epochs[epoch.Hourly{}] = 1
	epochs[epoch.TwoHourly{}] = 1
	epochs[epoch.Daily{}] = 1
	revs, err := a.GetRevisions(epochs, announcer.Limits{
		// An hour after latest commit.
		At: time.Date(2018, 4, 2, 3, 0, 0, 0, time.UTC),
		// Way before first tag.
		Start: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
	})
	assert.NotNil(t, revs)
	assert.Nil(t, err)

	hourlyRevs, ok := revs[epoch.Hourly{}]
	assert.True(t, ok)
	assert.Equal(t, 1, len(hourlyRevs))
	assert.Equal(t, agit.RevisionData{
		Hash:       tags[0].GetHash(t),
		CommitTime: tags[0].GetCommitTime(),
	}, hourlyRevs[0])

	twoHourlyRevs, ok := revs[epoch.TwoHourly{}]
	assert.True(t, ok)
	assert.Equal(t, 1, len(twoHourlyRevs))
	assert.Equal(t, agit.RevisionData{
		Hash:       tags[1].GetHash(t),
		CommitTime: tags[1].GetCommitTime(),
	}, twoHourlyRevs[0])

	dailyRevs, ok := revs[epoch.Daily{}]
	assert.True(t, ok)
	assert.Equal(t, 1, len(dailyRevs))
	assert.Equal(t, agit.RevisionData{
		Hash:       tags[2].GetHash(t),
		CommitTime: tags[2].GetCommitTime(),
	}, dailyRevs[0])
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
	refs := test.Tags(tags).Refs(t)
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: SliceReferenceIterFactory{
			&SliceReferenceIter{
				refs: refs,
			},
		},
		Git: test.NewMockRepository(t, tags, test.NilFetchImpl),
	})
	assert.NotNil(t, a)
	assert.Nil(t, err)

	epochs := make(map[epoch.Epoch]int)
	epochs[epoch.Hourly{}] = 2
	epochs[epoch.TwoHourly{}] = 1
	epochs[epoch.Daily{}] = 1
	revs, err := a.GetRevisions(epochs, announcer.Limits{
		// An hour after latest commit.
		At: time.Date(2018, 4, 2, 3, 0, 0, 0, time.UTC),
		// Way before first tag.
		Start: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
	})
	assert.NotNil(t, revs)
	assert.Nil(t, err)

	hourlyRevs, ok := revs[epoch.Hourly{}]
	assert.True(t, ok)
	assert.Equal(t, 2, len(hourlyRevs))
	assert.Equal(t, agit.RevisionData{
		Hash:       tags[0].GetHash(t),
		CommitTime: tags[0].GetCommitTime(),
	}, hourlyRevs[0])
	assert.Equal(t, agit.RevisionData{
		Hash:       tags[1].GetHash(t),
		CommitTime: tags[1].GetCommitTime(),
	}, hourlyRevs[1])

	twoHourlyRevs, ok := revs[epoch.TwoHourly{}]
	assert.True(t, ok)
	assert.Equal(t, 1, len(twoHourlyRevs))
	assert.Equal(t, agit.RevisionData{
		Hash:       tags[1].GetHash(t),
		CommitTime: tags[1].GetCommitTime(),
	}, twoHourlyRevs[0])

	dailyRevs, ok := revs[epoch.Daily{}]
	assert.True(t, ok)
	assert.Equal(t, 1, len(dailyRevs))
	assert.Equal(t, agit.RevisionData{
		Hash:       tags[2].GetHash(t),
		CommitTime: tags[2].GetCommitTime(),
	}, dailyRevs[0])
}

type MockRepositoryProducer struct {
	t      *testing.T
	clones int
}

func (g *MockRepositoryProducer) Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (agit.Repository, error) {
	g.clones++
	return test.NewMockRepository(g.t, []test.Tag{}, test.NilFetchImpl), nil
}

// TODO(markdittmer): Should test that gitRemoteAnnouncer droppped reference to initial repository. Not possible with black box testing.
func TestGitRemoteAnnouncer_Reset(t *testing.T) {
	g := MockRepositoryProducer{t: t}
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		Git: &g,
	})
	assert.NotNil(t, a)
	assert.Nil(t, err)
	prevClones := g.clones
	err = a.Reset()
	assert.Nil(t, err)
	assert.Equal(t, prevClones+1, g.clones)
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
	bRepo := test.NewMockRepository(t, []test.Tag{updatedTag}, test.NilFetchImpl)
	pValue := ProxyRepository{}
	pRepo := &pValue

	// TODO(markdittmer): This is brittle; it depends on the implementation detail
	// that aRepo setup will invoke Fetch() exactly once to ensure that its tags
	// are fetched. Consider switching to mockgen mocks for Repository objects.
	aFetchCount := 0
	aRepo := test.NewMockRepository(t, []test.Tag{}, func(mr *test.MockRepository, o *git.FetchOptions) error {
		aFetchCount++
		if aFetchCount == 1 {
			return nil
		}

		pRepo.Set(bRepo)
		return fetchErr
	})
	pRepo.Set(aRepo)
	a, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
		EpochReferenceIterFactory: pRepo,
		Git:                       pRepo,
	})
	assert.NotNil(t, a)
	assert.Nil(t, err)

	epochs := make(map[epoch.Epoch]int)
	epochs[epoch.Daily{}] = 1
	limits := announcer.Limits{
		// Day after tag in updated pRepo->bRepo.
		At: time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC),
		// Long before any tags.
		Start: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
	}
	getRevisions := func() (map[epoch.Epoch][]agit.Revision, error) {
		return a.GetRevisions(epochs, limits)
	}
	revs, err := getRevisions()
	assert.NotNil(t, revs)
	assert.Equal(t, announcer.GetErrNotAllEpochsConsumed(), err)

	err = a.Fetch()
	assert.Equal(t, fetchErr, err)
	revs, err = getRevisions()
	assert.NotNil(t, revs)
	assert.Nil(t, err)
	dailyRevs, ok := revs[epoch.Daily{}]
	assert.True(t, ok)
	assert.Equal(t, 1, len(dailyRevs))
	assert.Equal(t, agit.RevisionData{
		Hash:       updatedTag.GetHash(t),
		CommitTime: updatedTag.GetCommitTime(),
	}, dailyRevs[0])
}
