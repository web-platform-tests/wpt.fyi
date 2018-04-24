package git

import (
	"io"
	"sort"
	"strings"

	"log"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

type refCommit struct {
	ref    *plumbing.Reference
	commit *object.Commit
}
type refCommits []refCommit

func (rcs refCommits) Len() int {
	return len(rcs)
}
func (rcs refCommits) Swap(i, j int) {
	rcs[i], rcs[j] = rcs[j], rcs[i]
}
func (rcs refCommits) Less(i, j int) bool {
	if rcs[i].commit == nil {
		return false
	}
	if rcs[j].commit == nil {
		return true
	}
	return rcs[i].commit.Committer.When.After(rcs[j].commit.Committer.When)
}

func NewTimeOrderedReferenceIter(iter storer.ReferenceIter, repo Repository) (storer.ReferenceIter, error) {
	rcs := make([]refCommit, 0)
	var ref *plumbing.Reference
	var err error
	for ref, err = iter.Next(); ref != nil && err == nil; ref, err = iter.Next() {
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			log.Printf("WARN: Failed to lookup commit for reference %v", ref)
			continue
		}
		rcs = append(rcs, refCommit{
			ref,
			commit,
		})
	}
	if err != io.EOF {
		return nil, err
	}
	iter.Close()
	sort.Sort(refCommits(rcs))
	refs := make([]*plumbing.Reference, 0, len(rcs))
	for _, rcs := range rcs {
		refs = append(refs, rcs.ref)
	}
	return storer.NewReferenceSliceIter(refs), nil
}

func NewMergedPRIter(iter storer.ReferenceIter, repo Repository) (storer.ReferenceIter, error) {
	iter, err := NewTimeOrderedReferenceIter(storer.NewReferenceFilteredIter(func(ref *plumbing.Reference) bool {
		if ref == nil {
			return false
		}
		return strings.HasPrefix(string(ref.Name()), "refs/tags/merge_pr_")
	}, iter), repo)
	if err != nil {
		log.Printf("ERRO: Failed to construct new merged PR iter: %v", err)
		return nil, err
	}
	return iter, err
}

type StartReferenceIter struct {
	startAt func(ref *plumbing.Reference) bool
	iter    storer.ReferenceIter
	started bool
}

func (iter *StartReferenceIter) Next() (ref *plumbing.Reference, err error) {
	if iter.started {
		return iter.iter.Next()
	}
	for ref, err = iter.iter.Next(); err == nil; ref, err = iter.iter.Next() {
		if iter.startAt(ref) {
			iter.started = true
			return ref, nil
		}
	}
	return ref, err
}

func (iter *StartReferenceIter) ForEach(f func(*plumbing.Reference) error) error {
	return iter.iter.ForEach(func(ref *plumbing.Reference) error {
		if iter.started {
			return f(ref)
		}
		if iter.startAt(ref) {
			iter.started = true
			return f(ref)
		}
		return nil
	})
}

func (iter *StartReferenceIter) Close() {
	iter.iter.Close()
}

func NewStartReferenceIter(iter storer.ReferenceIter, startAt func(ref *plumbing.Reference) bool) storer.ReferenceIter {
	return &StartReferenceIter{
		startAt,
		iter,
		false,
	}
}

type StopReferenceIter struct {
	stopAt func(ref *plumbing.Reference) bool
	iter   storer.ReferenceIter
}

func (iter StopReferenceIter) Next() (ref *plumbing.Reference, err error) {
	ref, err = iter.iter.Next()
	if err != nil {
		return ref, err
	}
	if iter.stopAt(ref) {
		iter.Close()
		return nil, io.EOF
	}
	return ref, err
}

func (iter StopReferenceIter) ForEach(f func(*plumbing.Reference) error) error {
	return iter.iter.ForEach(func(ref *plumbing.Reference) error {
		if iter.stopAt(ref) {
			return io.EOF
		}
		return f(ref)
	})
}

func (iter StopReferenceIter) Close() {
	iter.iter.Close()
}

func NewStopReferenceIter(iter storer.ReferenceIter, stopAt func(ref *plumbing.Reference) bool) storer.ReferenceIter {
	return StopReferenceIter{
		stopAt,
		iter,
	}
}
