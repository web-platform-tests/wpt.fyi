// +build small

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
	_, err := l.EvictLRU()
	assert.NotNil(t, err)
}

func TestSimpleOrder(t *testing.T) {
	l := NewLRU()
	l.Access(1)
	l.Access(2)
	r, err := l.EvictLRU()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), r)
}

func TestRepeatAccess(t *testing.T) {
	l := NewLRU()
	l.Access(1)
	l.Access(2)
	l.Access(1)
	r, err := l.EvictLRU()
	assert.Nil(t, err)
	assert.Equal(t, int64(2), r)
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
			_, err := l.EvictLRU()
			assert.Nil(t, err)
		}(int64(i))
	}
	wg.Wait()

	_, err := l.EvictLRU()
	assert.NotNil(t, err)
}
