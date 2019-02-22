// +build medium

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package screenshot

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestNewScreenshot(t *testing.T) {
	s := NewScreenshot([]string{"", "chrome"})
	assert.Equal(t, s.Labels, []string{"chrome"})
}

func TestKeyAndHash(t *testing.T) {
	s := Screenshot{
		HashMethod: "hash",
		HashDigest: "0000abcd",
	}
	assert.Equal(t, "hash:0000abcd", s.Hash())
	assert.Equal(t, "0000abcd:hash", s.Key())
}

func TestSetHashFromFile(t *testing.T) {
	s := Screenshot{}
	err := s.SetHashFromFile(strings.NewReader("Hello, world!"), "sha1")
	assert.Nil(t, err)
	assert.Equal(t, "sha1", s.HashMethod)
	assert.Equal(t, "943a702d06f34599aee1f8da8ef9f7296031d699", s.HashDigest)
}

func TestSetHashFromFile_error(t *testing.T) {
	s := Screenshot{}
	err := s.SetHashFromFile(strings.NewReader(""), "hash")
	assert.Equal(t, ErrUnsupportedHashMethod, err)
}

func TestStore(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	ds := shared.NewAppEngineDatastore(ctx, false)

	t.Run("error", func(t *testing.T) {
		s := Screenshot{}
		err := s.Store(ds)
		assert.Equal(t, ErrInvalidHash, err)
	})
	t.Run("create new screenshot", func(t *testing.T) {
		s := Screenshot{
			HashDigest: "fa52883da345b2525304b54c8bc7bbb1e88b5e3e",
			HashMethod: "sha1",
			Labels:     []string{"chrome"},
		}
		err := s.Store(ds)
		assert.Nil(t, err)

		// Check populated fields.
		assert.Equal(t, 0, s.Counter)
		assert.False(t, s.LastUsed.IsZero())

		var s2 Screenshot
		err = ds.Get(ds.NewNameKey("Screenshot", s.Key()), &s2)
		assert.Nil(t, err)
		assert.Equal(t, s.HashDigest, s2.HashDigest)
		assert.Equal(t, s.HashMethod, s2.HashMethod)
		assert.Equal(t, s.Labels, s2.Labels)
		assert.Equal(t, s.Counter, s2.Counter)
	})
	t.Run("update a screenshot", func(t *testing.T) {
		s := Screenshot{
			HashDigest: "fa52883da345b2525304b54c8bc7bbb1e88b5e3e",
			HashMethod: "sha1",
			Labels:     []string{"firefox"},
		}
		err := s.Store(ds)
		assert.Nil(t, err)

		// Check populated fields.
		assert.Equal(t, 1, s.Counter)
		expectedLabels := shared.NewSetFromStringSlice([]string{"chrome", "firefox"})
		labels := shared.NewSetFromStringSlice(s.Labels)
		assert.True(t, expectedLabels.Equal(labels))
		assert.False(t, s.LastUsed.IsZero())

		var s2 Screenshot
		err = ds.Get(ds.NewNameKey("Screenshot", s.Key()), &s2)
		assert.Nil(t, err)
		assert.Equal(t, s.Labels, s2.Labels)
		assert.Equal(t, s.Counter, s2.Counter)
	})
}

func TestRecentScreenshotHashes(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	ds := shared.NewAppEngineDatastore(ctx, false)

	screenshots := []Screenshot{
		// The order matters: 0001 is the perfect match, and the rest
		// have have fewer and less important matching labels.
		Screenshot{
			HashDigest: "0001",
			HashMethod: "hash",
			Labels:     []string{"chrome", "64", "mac", "10.13"},
		},
		Screenshot{
			HashDigest: "0002",
			HashMethod: "hash",
			Labels:     []string{"chrome", "64", "mac", "10.14"},
		},
		Screenshot{
			HashDigest: "0003",
			HashMethod: "hash",
			Labels:     []string{"chrome", "64", "windows", "10"},
		},
		Screenshot{
			HashDigest: "0004",
			HashMethod: "hash",
			Labels:     []string{"chrome", "65", "windows", "10"},
		},
		Screenshot{
			HashDigest: "0005",
			HashMethod: "hash",
			Labels:     []string{"firefox", "60", "windows", "10"},
		},
	}
	for _, s := range screenshots {
		key := ds.NewNameKey("Screenshot", s.Key())
		_, err := ds.Put(key, &s)
		assert.Nil(t, err)
	}

	for i := 1; i <= 5; i++ {
		t.Run(fmt.Sprintf("%d screenshots", i), func(t *testing.T) {
			hashes, err := RecentScreenshotHashes(ds, "chrome", "64", "mac", "10.13", &i)
			assert.Nil(t, err)
			sort.Strings(hashes)
			for j, hash := range hashes {
				assert.Equal(t, fmt.Sprintf("hash:000%d", j+1), hash)
			}
		})
	}
}
