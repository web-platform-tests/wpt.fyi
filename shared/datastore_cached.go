// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"sync"
	"time"
)

// testRunCacheTTL is the expiration for each test run in Redis.
var testRunCacheTTL = 48 * time.Hour

// alignedRunSHAsCacheTTL is the expiration for the results of
// GetAlignedRunSHAs in Redis. Finding aligned runs reads every recent TestRun
// entity until enough SHAs have runs for all the products, which makes it the
// most expensive Datastore query that the site runs, and the default results
// page runs it on every load. A newly aligned set of runs shows up at most this
// much later.
var alignedRunSHAsCacheTTL = 2 * time.Minute

type cachedDatastore struct {
	Datastore
	ctx context.Context
}

// TestRunQuery returns a TestRunQuery that caches the results of
// GetAlignedRunSHAs. The other queries go to the underlying Datastore.
func (d cachedDatastore) TestRunQuery() TestRunQuery {
	return cachedTestRunQuery{
		TestRunQuery: d.Datastore.TestRunQuery(),
		store:        d.Datastore,
		ctx:          d.ctx,
	}
}

func (d cachedDatastore) Get(k Key, dst interface{}) error {
	if k.Kind() != "TestRun" {
		return d.Datastore.Get(k, dst)
	}

	cs := NewObjectCachedStore(
		d.ctx,
		NewJSONObjectCache(d.ctx, NewRedisReadWritable(d.ctx, testRunCacheTTL)),
		testRunObjectStore{d})
	return cs.Get(getTestRunRedisKey(k.IntID()), k.IntID(), dst)
}

func (d cachedDatastore) GetMulti(keys []Key, dst interface{}) error {
	for _, key := range keys {
		if key.Kind() != "TestRun" {
			return d.Datastore.GetMulti(keys, dst)
		}
	}

	runs := dst.(TestRuns)
	var wg sync.WaitGroup
	var err error
	for i := range keys {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if localErr := d.Get(keys[i], &runs[i]); localErr != nil {
				err = localErr
			}
		}(i)
	}
	wg.Wait()
	return err
}

// testRunObjectStore is an adapter from Datastore to ObjectStore.
type testRunObjectStore struct {
	cachedDatastore
}

func (d testRunObjectStore) Get(id, dst interface{}) error {
	intID, ok := id.(int64)
	if !ok {
		return errDatastoreObjectStoreExpectedInt64
	}
	key := d.NewIDKey("TestRun", intID)
	err := d.Datastore.Get(key, dst)
	if err == nil {
		run := dst.(*TestRun)
		run.ID = key.IntID()
	}
	return d.Datastore.Get(key, dst)
}
