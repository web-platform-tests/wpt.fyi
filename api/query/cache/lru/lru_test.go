//go:build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package lru

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRUEmpty(t *testing.T) {
	l := NewLRU()
	removed := l.EvictLRU(1.0)
	assert.Nil(t, removed)
}

func TestSimple(t *testing.T) {
	l := NewLRU()
	l.Access(1)
	l.Access(2)
	removed := l.EvictLRU(0.5)
	assert.Equal(t, []int64{int64(1)}, removed)
}

func TestRoundUpToOne(t *testing.T) {
	l := NewLRU()
	l.Access(1)
	l.Access(2)
	removed := l.EvictLRU(-1000.0)
	assert.Equal(t, []int64{int64(1)}, removed)
}

func TestRoundDown(t *testing.T) {
	l := NewLRU()
	l.Access(1)
	l.Access(2)
	removed := l.EvictLRU(0.99999)
	assert.Equal(t, []int64{int64(1)}, removed)
}

func TestRoundWayDown(t *testing.T) {
	l := NewLRU()
	l.Access(1)
	l.Access(2)
	removed := l.EvictLRU(1000.0)
	assert.Equal(t, []int64{int64(1), int64(2)}, removed)
}

func TestRepeatAccess(t *testing.T) {
	l := NewLRU()
	l.Access(1)
	l.Access(2)
	l.Access(1)
	// Remove slightly over half of the items to avoid floating point
	// errors in the case that 0.5 * 2 is < 1.
	removed := l.EvictLRU(0.51)
	assert.Equal(t, []int64{int64(2)}, removed)
}

func TestConcurrency(t *testing.T) {
	l := NewLRU()
	var wg sync.WaitGroup
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			l.Access(id)
			wait := rand.Int() % 1000
			time.Sleep(time.Nanosecond * time.Duration(wait))
			removed := l.EvictLRU(0.0)
			assert.Equal(t, 1, len(removed))
		}(int64(i))
	}
	wg.Wait()

	removed := l.EvictLRU(0.0)
	assert.Nil(t, removed)
}
