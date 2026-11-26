// Copyright 2026 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
)

// cachedTestRunQuery is a TestRunQuery that caches the results of
// GetAlignedRunSHAs in Redis for alignedRunSHAsCacheTTL.
type cachedTestRunQuery struct {
	TestRunQuery

	store Datastore
	ctx   context.Context // nolint:containedctx // Same as cachedDatastore.
}

// cachedAlignedRunSHAs is the cached form of the results of GetAlignedRunSHAs.
type cachedAlignedRunSHAs struct {
	SHAs []string `json:"shas"`
	// RunIDs maps each SHA to the IDs of its TestRun entities, by product index.
	RunIDs map[string][][]int64 `json:"run_ids"`
}

func (q cachedTestRunQuery) GetAlignedRunSHAs(
	products ProductSpecs,
	labels mapset.Set,
	from,
	to *time.Time,
	limit *int,
	offset *int) (shas []string, keys map[string]KeysByProduct, err error) {
	logger := GetLogger(q.ctx)
	cache := NewJSONObjectCache(q.ctx, NewRedisReadWritable(q.ctx, alignedRunSHAsCacheTTL))
	cacheKey := alignedRunSHAsCacheKey(products, labels, from, to, limit, offset)

	var cached cachedAlignedRunSHAs
	if err := cache.Get(cacheKey, &cached); err == nil {
		return cached.SHAs, cached.keys(q.store), nil
	}

	shas, keys, err = q.TestRunQuery.GetAlignedRunSHAs(products, labels, from, to, limit, offset)
	if err != nil {
		return shas, keys, err
	}
	if err := cache.Put(cacheKey, newCachedAlignedRunSHAs(shas, keys)); err != nil {
		logger.Warningf("Failed to cache aligned run SHAs at key %s: %s", cacheKey, err.Error())
	}

	return shas, keys, nil
}

func newCachedAlignedRunSHAs(shas []string, keys map[string]KeysByProduct) cachedAlignedRunSHAs {
	cached := cachedAlignedRunSHAs{
		SHAs:   shas,
		RunIDs: make(map[string][][]int64, len(keys)),
	}
	for sha, keysByProduct := range keys {
		ids := make([][]int64, len(keysByProduct))
		for i, productKeys := range keysByProduct {
			ids[i] = []int64(GetTestRunIDs(productKeys.Keys))
		}
		cached.RunIDs[sha] = ids
	}

	return cached
}

// idKeyMaker is the part of Datastore that rebuilds keys from IDs.
type idKeyMaker interface {
	NewIDKey(typeName string, id int64) Key
}

// keys rebuilds the KeysByProduct map returned by GetAlignedRunSHAs.
func (c cachedAlignedRunSHAs) keys(store idKeyMaker) map[string]KeysByProduct {
	keys := make(map[string]KeysByProduct, len(c.RunIDs))
	for sha, ids := range c.RunIDs {
		keysByProduct := make(KeysByProduct, len(ids))
		for i, productIDs := range ids {
			for _, id := range productIDs {
				keysByProduct[i].Keys = append(keysByProduct[i].Keys, store.NewIDKey("TestRun", id))
			}
		}
		keys[sha] = keysByProduct
	}

	return keys
}

// alignedRunSHAsCacheKey returns a Redis key for the arguments of
// GetAlignedRunSHAs.
func alignedRunSHAsCacheKey(
	products ProductSpecs,
	labels mapset.Set,
	from,
	to *time.Time,
	limit *int,
	offset *int) string {
	var sortedLabels []string
	if labels != nil {
		sortedLabels = ToStringSlice(labels)
		sort.Strings(sortedLabels)
	}
	formatTime := func(t *time.Time) string {
		if t == nil {
			return ""
		}

		return t.UTC().Format(time.RFC3339Nano)
	}
	formatInt := func(i *int) string {
		if i == nil {
			return ""
		}

		return fmt.Sprint(*i)
	}
	// Product order matters: the returned keys are indexed by product.
	args := strings.Join([]string{
		strings.Join(products.Strings(), ","),
		strings.Join(sortedLabels, ","),
		formatTime(from),
		formatTime(to),
		formatInt(limit),
		formatInt(offset),
	}, "|")
	sum := sha256.Sum256([]byte(args))

	return "ALIGNED_RUN_SHAS-" + hex.EncodeToString(sum[:])
}
