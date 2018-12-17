// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package lru

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// LRU is a least recently used collection that supports element acces and
// item eviction. The first access of an unevicted value implicitly adds the
// value to the collection.
type LRU interface {
	// Access stores the current time as the "last accessed" time for the given
	// value in the collection. The first access of an unevicted value implicitly
	// adds the value to the collection.
	Access(int64)
	// EvictLRU deletes the oldest value from the collection and returns it. If
	// the collection is empty, then an error is returned.
	EvictLRU() (int64, error)
}

type lru struct {
	byRunID map[int64]time.Time
	m       *sync.Mutex
}

type lruEntry struct {
	int64
	time.Time
}

type byTime []lruEntry

func (s byTime) Len() int {
	return len(s)
}

func (s byTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byTime) Less(i, j int) bool {
	return s[i].Time.Before(s[j].Time)
}

var errEmpty = errors.New("LRU is empty")

func (l *lru) Access(r int64) {
	l.syncAccess(r)
}

func (l *lru) EvictLRU() (int64, error) {
	if len(l.byRunID) == 0 {
		return int64(0), errEmpty
	}

	return l.syncEvictLRU(), nil
}

// NewLRU constructs a new empty LRU.
func NewLRU() LRU {
	return &lru{
		byRunID: make(map[int64]time.Time),
		m:       &sync.Mutex{},
	}
}

func (l *lru) syncAccess(r int64) {
	l.m.Lock()
	defer l.m.Unlock()

	l.byRunID[r] = time.Now()
}

func (l *lru) syncEvictLRU() int64 {
	l.m.Lock()
	defer l.m.Unlock()

	rs := make(byTime, 0, len(l.byRunID))
	for r, t := range l.byRunID {
		rs = append(rs, lruEntry{r, t})
	}
	sort.Sort(rs)
	toRemove := rs[0].int64
	delete(l.byRunID, toRemove)

	return toRemove
}
