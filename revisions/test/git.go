// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package test

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	agit "github.com/web-platform-tests/wpt.fyi/revisions/git"
	billy "gopkg.in/src-d/go-billy.v4"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/storage"
)

type MockReferenceIter struct {
	refs     []*plumbing.Reference
	idx      int
	isClosed bool
}

func (iter *MockReferenceIter) Next() (ref *plumbing.Reference, err error) {
	if iter.isClosed {
		return nil, errors.New("iter.Next() after iter.Close()")
	}
	if iter.idx >= len(iter.refs) {
		return nil, io.EOF
	}
	ref = iter.refs[iter.idx]
	iter.idx++
	return ref, err
}

func (iter *MockReferenceIter) ForEach(f func(*plumbing.Reference) error) error {
	if iter.isClosed {
		return errors.New("iter.ForEach() after iter.Close()")
	}
	if iter.idx >= len(iter.refs) {
		return io.EOF
	}

	refs := iter.refs[iter.idx:]
	for _, ref := range refs {
		err := f(ref)
		if err != nil {
			return err
		}
		iter.idx++
	}

	return nil
}

func (iter *MockReferenceIter) Close() {
	iter.isClosed = true
}

func NewMockIter(refs []*plumbing.Reference) MockReferenceIter {
	return MockReferenceIter{refs, 0, false}
}

func NewHash(t *testing.T, hashStr string) plumbing.Hash {
	hashSlice, err := hex.DecodeString(hashStr)
	if err != nil {
		t.Fatalf("NewHash() expects hex string, but got %s", hashStr)
	}
	if len(hashSlice) > 20 {
		t.Fatalf("NewHashRef() expects hex string constituting no more than 20 bytes but got %d bytes from %s", len(hashSlice), hashStr)
	}
	padStop := 20 - len(hashSlice)
	var fixedHash [20]byte
	for i := range fixedHash {
		if i < padStop {
			fixedHash[i] = byte(0)
		} else {
			fixedHash[i] = hashSlice[i-padStop]
		}
	}
	// Uncomment this line when debugging.
	// t.Logf("INFO: %x", fixedHash)
	return plumbing.Hash(fixedHash)
}

func NewTagRefFromHash(hash plumbing.Hash, name string) *plumbing.Reference {
	refName := plumbing.ReferenceName(name)
	return plumbing.NewHashReference("refs/tags/"+refName, hash)
}

func NewTagRef(t *testing.T, name, hashStr string) *plumbing.Reference {
	return NewTagRefFromHash(NewHash(t, hashStr), name)
}

type Tag struct {
	TagName    string
	Hash       string
	CommitTime time.Time

	hash   *plumbing.Hash
	tag    *plumbing.Reference
	commit *object.Commit
}

func (t Tag) GetHash(T *testing.T) plumbing.Hash {
	if t.hash != nil {
		return *t.hash
	}
	hash := NewHash(T, t.Hash)
	t.hash = &hash
	return hash
}

func (t Tag) GetCommitTime() time.Time {
	return t.CommitTime
}

func (t Tag) GetTag(T *testing.T) *plumbing.Reference {
	if t.tag != nil {
		return t.tag
	}
	tag := NewTagRefFromHash(t.GetHash(T), t.TagName)
	t.tag = tag
	return tag
}

func (t Tag) GetCommit(T *testing.T) *object.Commit {
	if t.commit != nil {
		return t.commit
	}
	commit := NewCommitFromHash(t.GetHash(T), t.CommitTime)
	t.commit = commit
	return commit
}

type Tags []Tag

func (ts Tags) Refs(T *testing.T) []*plumbing.Reference {
	refs := make([]*plumbing.Reference, 0, len(ts))
	for _, t := range ts {
		refs = append(refs, t.GetTag(T))
	}
	return refs
}

type FetchImpl func(mr *MockRepository, o *git.FetchOptions) error

var NilFetchImpl = func(mr *MockRepository, o *git.FetchOptions) error {
	return nil
}

type MockRepository struct {
	refs      []*plumbing.Reference
	commits   map[plumbing.Hash]*object.Commit
	fetchImpl FetchImpl
}

func (mr *MockRepository) CommitObject(hash plumbing.Hash) (*object.Commit, error) {
	commit, ok := mr.commits[hash]
	if !ok {
		return nil, fmt.Errorf("Unable to locate commit for hash %x", hash)
	}
	return commit, nil
}

func (mr *MockRepository) Tags() (storer.ReferenceIter, error) {
	iter := NewMockIter(mr.refs)
	return &iter, nil
}

func (mr *MockRepository) Fetch(o *git.FetchOptions) error {
	return mr.fetchImpl(mr, o)
}

func (mr *MockRepository) Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (agit.Repository, error) {
	return mr, nil
}

func NewMockRepository(t *testing.T, tags []Tag, fetchImpl FetchImpl) *MockRepository {
	refs := make([]*plumbing.Reference, 0, len(tags))
	commits := make(map[plumbing.Hash]*object.Commit)
	for _, tag := range tags {
		refs = append(refs, tag.GetTag(t))
		commits[tag.GetHash(t)] = tag.GetCommit(t)
	}
	return &MockRepository{
		refs,
		commits,
		fetchImpl,
	}
}

func NewCommitFromHash(hash plumbing.Hash, commitTime time.Time) *object.Commit {
	return &object.Commit{
		Hash: hash,
		Committer: object.Signature{
			When: commitTime,
		},
	}
}

func NewCommit(t *testing.T, hashStr string, commitTime time.Time) *object.Commit {
	hash := NewHash(t, hashStr)
	return NewCommitFromHash(hash, commitTime)
}
