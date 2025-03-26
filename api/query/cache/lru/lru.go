// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package lru

import (
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
	Access(value int64)
	// EvictLRU removes and returns a fraction of the collection, based on
	// the passed percentage. It will always remove at least one item. When
	// deciding which items to remove, EvictLRU deletes older values from
	// the collection first. If the collection is empty, nil is returned.
	EvictLRU(percent float64) []int64
}

type lru struct {
	values map[int64]time.Time
	m      *sync.Mutex
}

type lruEntry struct {
	int64
	time.Time
}

type lruEntries []lruEntry

// Satisfy the Sort interface requirements (https://golang.org/pkg/sort/#Sort)
func (s lruEntries) Len() int           { return len(s) }
func (s lruEntries) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s lruEntries) Less(i, j int) bool { return (s[i].Time).Before(s[j].Time) }

func (l *lru) Access(v int64) {
	l.syncAccess(v)
}

func (l *lru) EvictLRU(percent float64) []int64 {
	if len(l.values) == 0 {
		return nil
	}
	percent = math.Max(0.0, math.Min(1.0, percent))
	numValuesToDelete := math.Floor(float64(len(l.values)) * percent)
	// Always delete at least one entry.
	return l.syncEvictLRU(int(math.Max(1.0, numValuesToDelete)))
}

// NewLRU constructs a new empty LRU.
// nolint:ireturn // TODO: Fix ireturn lint error
func NewLRU() LRU {
	return &lru{
		values: make(map[int64]time.Time),
		m:      &sync.Mutex{},
	}
}

func (l *lru) syncAccess(v int64) {
	l.m.Lock()
	defer l.m.Unlock()

	l.values[v] = time.Now()
}

func (l *lru) syncEvictLRU(num int) []int64 {
	l.m.Lock()
	defer l.m.Unlock()

	vs := make(lruEntries, 0, len(l.values))
	for v, t := range l.values {
		vs = append(vs, lruEntry{v, t})
	}
	sort.Sort(vs)
	toRemove := vs[:num]
	ret := make([]int64, num)
	for i := range toRemove {
		value := toRemove[i].int64
		ret[i] = value
		delete(l.values, value)
	}

	return ret
}
