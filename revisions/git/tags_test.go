package git_test

import (
	"io"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	agit "github.com/web-platform-tests/wpt.fyi/revisions/git"
	"github.com/web-platform-tests/wpt.fyi/revisions/test"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

func TestTimeOrderedReferenceIter_Simple(t *testing.T) {
	tags := []test.Tag{
		test.Tag{
			TagName:    "not_a_mergedpr_1",
			Hash:       "01",
			CommitTime: time.Date(2018, 4, 3, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "not_a_mergedpr_2",
			Hash:       "02",
			CommitTime: time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "not_a_mergedpr_3",
			Hash:       "03",
			CommitTime: time.Date(2018, 4, 6, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "merge_pr_4",
			Hash:       "04",
			CommitTime: time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "not_a_mergedpr_5",
			Hash:       "05",
			CommitTime: time.Date(2018, 4, 5, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "merge_pr_6",
			Hash:       "06",
			CommitTime: time.Date(2018, 4, 4, 0, 0, 0, 0, time.UTC),
		},
	}
	refs := test.Tags(tags).Refs()
	iterFactory := func(t *testing.T) storer.ReferenceIter {
		baseIter := test.NewMockIter(refs)
		iter, err := agit.NewTimeOrderedReferenceIter(&baseIter, test.NewMockRepository(tags, test.NilFetchImpl))
		assert.True(t, err == nil)
		return iter
	}

	sortedIdxs := []int{2, 4, 5, 0, 1, 3}
	var ref *plumbing.Reference
	var err error

	iter := iterFactory(t)
	i := 0
	for ref, err = iter.Next(); ref != nil && err == nil; ref, err = iter.Next() {
		assert.True(t, refs[sortedIdxs[i]] == ref)
		i++
	}
	assert.True(t, err == io.EOF)

	iter = iterFactory(t)
	i = 0
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, i < len(refs))
		assert.True(t, refs[sortedIdxs[i]] == ref)
		i++
		return nil
	})
	assert.True(t, err == nil)
}

func TestFilteredReferenceIter_Custom(t *testing.T) {
	refs := []*plumbing.Reference{
		test.NewTagRef("tag_1", "01"),
		test.NewTagRef("tag_2", "02"),
		test.NewTagRef("tag_3", "03"),
	}

	// Include every other commit.
	newFilter := func() agit.ReferencePredicate {
		include := false
		return func(ref *plumbing.Reference) bool {
			include = !include
			return include
		}
	}

	iter := test.NewMockIter(refs)
	stopIter := agit.NewFilteredReferenceIter(&iter, newFilter())
	firstRef, err := stopIter.Next()
	assert.True(t, err == nil)
	assert.True(t, firstRef == refs[0])
	secondRef, err := stopIter.Next()
	assert.True(t, err == nil)
	assert.True(t, secondRef == refs[2])
	_, err = stopIter.Next()
	assert.True(t, err == io.EOF)

	iter = test.NewMockIter(refs)
	stopIter = agit.NewFilteredReferenceIter(&iter, newFilter())
	i := 0
	err = stopIter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, ref == refs[i])
		i += 2
		return nil
	})
	assert.True(t, err == nil)
}

// TODO(markdittmer): Rename TestMergedPRIter TestMergedPRIter.
// Test custom filtered reference iter.

func TestMergedPRIter_Simple(t *testing.T) {
	tags := []test.Tag{
		test.Tag{
			TagName:    "not_a_mergedpr_1",
			Hash:       "01",
			CommitTime: time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "not_a_mergedpr_2",
			Hash:       "02",
			CommitTime: time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "not_a_mergedpr_3",
			Hash:       "03",
			CommitTime: time.Date(2018, 4, 3, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "merge_pr_4",
			Hash:       "04",
			CommitTime: time.Date(2018, 4, 4, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "not_a_mergedpr_5",
			Hash:       "05",
			CommitTime: time.Date(2018, 4, 5, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "merge_pr_6",
			Hash:       "06",
			CommitTime: time.Date(2018, 4, 6, 0, 0, 0, 0, time.UTC),
		},
	}
	refs := test.Tags(tags).Refs()
	// merge_pr_* from refs in reverse cronological order.
	prs := [2]*plumbing.Reference{refs[5], refs[3]}
	repo := test.NewMockRepository(tags, test.NilFetchImpl)
	baseIter := test.NewMockIter(refs)
	filteredIter, err := agit.NewMergedPRIter(&baseIter, repo)
	assert.True(t, err == nil)
	i := 0
	for ref, err := filteredIter.Next(); err == nil; ref, err = filteredIter.Next() {
		assert.True(t, ref == prs[i])
		i++
	}
	assert.True(t, i == len(prs))

	baseIter = test.NewMockIter(refs)
	filteredIter, err = agit.NewMergedPRIter(&baseIter, repo)
	assert.True(t, err == nil)
	i = 0
	filteredIter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, ref == prs[i])
		i++
		return nil
	})
	assert.True(t, i == len(prs))
}

func stopAtHash(h plumbing.Hash) agit.ReferencePredicate {
	return func(ref *plumbing.Reference) bool {
		if ref == nil {
			log.Fatal("Unexpected nil reference in test stopAtHash function")
		}
		return ref.Hash() == h
	}
}

func TestStopReferenceIter_Simple(t *testing.T) {
	stopAt := test.NewHash("04")
	includedRefs := []*plumbing.Reference{
		test.NewTagRef("some_tag_1", "01"),
		test.NewTagRef("some_tag_2", "02"),
		test.NewTagRef("some_tag_3", "03"),
	}
	var allRefs []*plumbing.Reference
	allRefs = append(allRefs, includedRefs...)
	allRefs = append(allRefs, []*plumbing.Reference{
		test.NewTagRef("some_tag_4", "04"),
		test.NewTagRef("some_tag_5", "05"),
		test.NewTagRef("some_tag_6", "06"),
	}...)

	baseIter := test.NewMockIter(allRefs)
	filteredIter := agit.NewStopReferenceIter(&baseIter, stopAtHash(stopAt))
	i := 0
	for ref, err := filteredIter.Next(); err == nil; ref, err = filteredIter.Next() {
		assert.True(t, ref == includedRefs[i])
		i++
	}
	assert.True(t, i == len(includedRefs))

	baseIter = test.NewMockIter(allRefs)
	filteredIter = agit.NewStopReferenceIter(&baseIter, stopAtHash(stopAt))
	i = 0
	filteredIter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, ref == includedRefs[i])
		i++
		return nil
	})
	assert.True(t, i == len(includedRefs))
}

func TestMergedPRIter_StopReferenceIter_Compose(t *testing.T) {
	tags := []test.Tag{
		test.Tag{
			TagName:    "not_a_mergedpr_1",
			Hash:       "01",
			CommitTime: time.Date(2018, 4, 1, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "not_a_mergedpr_2",
			Hash:       "02",
			CommitTime: time.Date(2018, 4, 2, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "not_a_mergedpr_3",
			Hash:       "03",
			CommitTime: time.Date(2018, 4, 3, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "merge_pr_4",
			Hash:       "04",
			CommitTime: time.Date(2018, 4, 4, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "not_a_mergedpr_5",
			Hash:       "05",
			CommitTime: time.Date(2018, 4, 5, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "merge_pr_6",
			Hash:       "06",
			CommitTime: time.Date(2018, 4, 6, 0, 0, 0, 0, time.UTC),
		},
		test.Tag{
			TagName:    "merge_pr_7",
			Hash:       "07",
			CommitTime: time.Date(2018, 4, 7, 0, 0, 0, 0, time.UTC),
		},
	}
	refs := test.Tags(tags).Refs()
	// Stop at (reverse cronological) merge_pr_6.
	stopAt := tags[5].GetHash()
	// Included in iteration: merge_pr_7 only.
	includedPrs := [1]*plumbing.Reference{refs[6]}

	repo := test.NewMockRepository(tags, test.NilFetchImpl)
	baseIter := test.NewMockIter(refs)
	filteredIter, err := agit.NewMergedPRIter(&baseIter, repo)
	assert.True(t, err == nil)
	iter := agit.NewStopReferenceIter(filteredIter, stopAtHash(stopAt))
	i := 0
	for ref, err := iter.Next(); err == nil; ref, err = iter.Next() {
		assert.True(t, ref == includedPrs[i])
		i++
	}
	assert.True(t, i == len(includedPrs))

	baseIter = test.NewMockIter(refs)
	filteredIter, err = agit.NewMergedPRIter(&baseIter, repo)
	assert.True(t, err == nil)
	iter = agit.NewStopReferenceIter(filteredIter, stopAtHash(stopAt))
	assert.True(t, err == nil)
	i = 0
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		assert.True(t, ref == includedPrs[i])
		i++
		return nil
	})
	assert.True(t, i == len(includedPrs))
}
