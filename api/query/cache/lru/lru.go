// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package lru

import (
	"errors"
	"math"
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
	EvictLRU(float64) []int64
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

func (s byTime) Len() int           { return len(s) }
func (s byTime) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool { return s[i].Time.Before(s[j].Time) }

var errEmpty = errors.New("LRU is empty")

func (l *lru) Access(r int64) {
	l.syncAccess(r)
}

func (l *lru) EvictLRU(percent float64) []int64 {
	if len(l.byRunID) == 0 {
		return nil
	}
	percent = math.Max(0.0, math.Min(1.0, percent))
	return l.syncEvictLRU(int(math.Max(1.0, math.Floor(float64(len(l.byRunID))*percent))))
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

func (l *lru) syncEvictLRU(num int) []int64 {
	l.m.Lock()
	defer l.m.Unlock()

	rs := make(byTime, 0, len(l.byRunID))
	for r, t := range l.byRunID {
		rs = append(rs, lruEntry{r, t})
	}
	sort.Sort(rs)
	toRemove := rs[:num]
	ret := make([]int64, num)
	for i := range toRemove {
		id := toRemove[i].int64
		ret[i] = id
		delete(l.byRunID, id)
	}

	return ret
}
