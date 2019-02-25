// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"sync"
	"time"
)

// aeTestRunCacheTTL is the expiration for each test run in Memcache.
var aeTestRunCacheTTL = 48 * time.Hour

type aeCachedDatastore struct {
	aeDatastore
}

func (d aeCachedDatastore) Get(k Key, dst interface{}) error {
	if k.Kind() != "TestRun" {
		return d.aeDatastore.Get(k, dst)
	}

	cs := NewObjectCachedStore(
		d.ctx,
		NewJSONObjectCache(d.ctx, NewMemcacheReadWritable(d.ctx, aeTestRunCacheTTL)),
		aeTestRunObjectStore{d})
	return cs.Get(getTestRunMemcacheKey(k.IntID()), k.IntID(), dst)
}

func (d aeCachedDatastore) GetMulti(keys []Key, dst interface{}) error {
	for _, key := range keys {
		if key.Kind() != "TestRun" {
			return d.aeDatastore.GetMulti(keys, dst)
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

func (d aeCachedDatastore) Delete(k Key) error {
	// TODO: Clear item from cache?
	return d.aeDatastore.Delete(k)
}

// aeTestRunObjectStore is an adapter from Datastore to ObjectStore.
type aeTestRunObjectStore struct {
	aeCachedDatastore
}

func (d aeTestRunObjectStore) Get(id, dst interface{}) error {
	intID, ok := id.(int64)
	if !ok {
		return errDatastoreObjectStoreExpectedInt64
	}
	key := d.NewIDKey("TestRun", intID)
	err := d.aeDatastore.Get(key, dst)
	if err == nil {
		run := dst.(*TestRun)
		run.ID = key.IntID()
	}
	return d.aeDatastore.Get(key, dst)
}
