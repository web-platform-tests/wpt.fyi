// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"sync"
	"time"
)

// aeTestRunCacheTTL is the expiration for each test run in Memcache.
var aeTestRunCacheTTL = 48 * time.Hour

type aeCachedDatastore struct {
	aeDatastore
}

// NewAppEngineCachedDatastore creates a Datastore implementation with Memcache
// in front to cache all TestRun reads. It is backed by the appengine libraries,
// used in AppEngine standard.
func NewAppEngineCachedDatastore(ctx context.Context) Datastore {
	return aeCachedDatastore{
		aeDatastore{ctx: ctx},
	}
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

func (d aeCachedDatastore) loadTestRunsByKeys(keysByProduct KeysByProduct) (result TestRunsByProduct, err error) {
	result = TestRunsByProduct{}
	cs := NewObjectCachedStore(
		d.ctx,
		NewJSONObjectCache(d.ctx, NewMemcacheReadWritable(d.ctx, aeTestRunCacheTTL)),
		aeTestRunObjectStore{d})
	var wg sync.WaitGroup
	for _, kbp := range keysByProduct {
		runs := make(TestRuns, len(kbp.Keys))
		for i := range kbp.Keys {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				localErr := cs.Get(getTestRunMemcacheKey(kbp.Keys[i].IntID()), kbp.Keys[i].IntID(), &runs[i])
				if localErr != nil {
					err = localErr
				}
			}(i)
		}
		result = append(result, ProductTestRuns{
			Product:  kbp.Product,
			TestRuns: runs,
		})
		wg.Wait()

		if err != nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}
	return result, err
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
	key := d.NewKey("TestRun", intID)
	return d.aeCachedDatastore.Get(key, dst)
}
