package git

import (
	"flag"
	"io"
	"sort"
	"strings"

	"log"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

var gitChunkSize *int

func init() {
	flag.Int("git_size_chunk", 100, "Number of Git objects to fetch in chunks to extend depth-limited fetches")
}

type TimeOrderedReferenceIter struct {
	refs []*plumbing.Reference
	idx  int
	repo Repository
}

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

func (iter *TimeOrderedReferenceIter) Next() (ref *plumbing.Reference, err error) {
	if iter.idx >= len(iter.refs) {
		err = io.EOF
	} else {
		ref = iter.refs[iter.idx]
		iter.idx++
	}
	return ref, err
}

func (iter *TimeOrderedReferenceIter) ForEach(f func(*plumbing.Reference) error) error {
	for _, ref := range iter.refs {
		err := f(ref)
		if err != nil {
			return err
		}
	}
	return nil
}

func (iter *TimeOrderedReferenceIter) Close() {
	iter.refs = make([]*plumbing.Reference, 0)
}

func NewTimeOrderedReferenceIter(iter storer.ReferenceIter, repo Repository) (storer.ReferenceIter, error) {
	tori := TimeOrderedReferenceIter{
		repo: repo,
	}
	rcs := make([]refCommit, 0)
	var ref *plumbing.Reference
	var err error
	for ref, err = iter.Next(); ref != nil && err == nil; ref, err = iter.Next() {
		commit, err := tori.repo.CommitObject(ref.Hash())
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
	for _, rcs := range rcs {
		tori.refs = append(tori.refs, rcs.ref)
	}
	return &tori, nil
}

type ReferencePredicate func(*plumbing.Reference) bool

type FilteredReferenceIter struct {
	filter ReferencePredicate
	iter   storer.ReferenceIter
}

func (iter FilteredReferenceIter) Next() (ref *plumbing.Reference, err error) {
	for ref, err = iter.iter.Next(); err == nil && !iter.filter(ref); ref, err = iter.iter.Next() {
	}
	return ref, err
}

func (iter FilteredReferenceIter) ForEach(f func(*plumbing.Reference) error) error {
	return iter.iter.ForEach(func(ref *plumbing.Reference) error {
		if iter.filter(ref) {
			return f(ref)
		}
		return nil
	})
}

func (iter FilteredReferenceIter) Close() {
	iter.iter.Close()
}

func NewFilteredReferenceIter(iter storer.ReferenceIter, f ReferencePredicate) storer.ReferenceIter {
	return FilteredReferenceIter{
		f,
		iter,
	}
}

func NewMergedPRIter(iter storer.ReferenceIter, repo Repository) (storer.ReferenceIter, error) {
	iter, err := NewTimeOrderedReferenceIter(FilteredReferenceIter{
		filter: func(ref *plumbing.Reference) bool {
			if ref == nil {
				return false
			}
			return strings.HasPrefix(string(ref.Name()), "refs/tags/merge_pr_")
		},
		iter: iter,
	}, repo)
	if err != nil {
		log.Printf("ERRO: Failed to construct new merged PR iter: %v", err)
		return nil, err
	}
	return iter, err
}

type StartReferenceIter struct {
	startAt ReferencePredicate
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

func NewStartReferenceIter(iter storer.ReferenceIter, startAt ReferencePredicate) storer.ReferenceIter {
	return &StartReferenceIter{
		startAt,
		iter,
		false,
	}
}

type StopReferenceIter struct {
	stopAt ReferencePredicate
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

func NewStopReferenceIter(iter storer.ReferenceIter, stopAt ReferencePredicate) storer.ReferenceIter {
	return StopReferenceIter{
		stopAt,
		iter,
	}
}
