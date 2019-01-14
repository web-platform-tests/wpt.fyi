// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
	"google.golang.org/appengine/datastore"
)

// NewAppEngineDatastore creates a Datastore implementation that is backed by
// the appengine libraries, used in AppEngine standard.
func NewAppEngineDatastore(ctx context.Context) Datastore {
	return aeDatastore{
		ctx: ctx,
	}
}

// aeCachedStore is an appengine-query backed ObjectCachedStore.
type aeCachedStore struct {
	ctx        context.Context
	entityName string
}

func (cache aeCachedStore) Get(iID, value interface{}) error {
	id, ok := iID.(int64)
	if !ok {
		return errDatastoreObjectStoreExpectedInt64
	}
	err := datastore.Get(cache.ctx, datastore.NewKey(cache.ctx, cache.entityName, "", id, nil), value)
	if err == nil {
		// Set the TestRun.ID field on the way past.
		if run, ok := value.(*TestRun); ok {
			(*run).ID = id
		}
	}
	return err
}

type aeDatastore struct {
	ctx context.Context
}

func (d aeDatastore) Context() context.Context {
	return d.ctx
}

func (d aeDatastore) NewQuery(typeName string) Query {
	return aeQuery{
		query: datastore.NewQuery(typeName),
	}
}

func (d aeDatastore) NewKey(typeName string, id int64) Key {
	return datastore.NewKey(d.ctx, typeName, "", id, nil)
}

func (d aeDatastore) GetAll(q Query, dst interface{}) ([]Key, error) {
	keys, err := q.(aeQuery).query.GetAll(d.ctx, dst)
	cast := make([]Key, len(keys))
	for i := range keys {
		cast[i] = keys[i]
	}
	return cast, err
}

// Get wraps a standard "get by key" functionality of appengine
// datastore with a 48h memcache layer.
func (d aeDatastore) Get(k Key, dst interface{}) error {
	cs := NewObjectCachedStore(
		d.ctx,
		NewJSONObjectCache(d.ctx, NewMemcacheReadWritable(d.ctx, 48*time.Hour)),
		aeCachedStore{
			ctx:        d.ctx,
			entityName: k.Kind(),
		})
	return cs.Get(getTestRunMemcacheKey(k.IntID()), k.IntID(), dst)
}

func (d aeDatastore) GetMulti(keys []Key, dst interface{}) error {
	cast := make([]*datastore.Key, len(keys))
	for i := range keys {
		cast[i] = keys[i].(*datastore.Key)
	}
	return datastore.GetMulti(d.ctx, cast, dst)
}

func (d aeDatastore) LoadTestRuns(
	products []ProductSpec,
	labels mapset.Set,
	revisions []string,
	from *time.Time,
	to *time.Time,
	limit,
	offset *int) (result TestRunsByProduct, err error) {
	return loadTestRuns(d, products, labels, revisions, from, to, limit, offset)
}

// LoadTestRunsByKeys wraps a standard "get by key" functionality of appengine
//  datastore with a 48h memcache layer.
func (d aeDatastore) LoadTestRunsByKeys(keysByProduct KeysByProduct) (result TestRunsByProduct, err error) {
	result = TestRunsByProduct{}
	cs := NewObjectCachedStore(
		d.ctx,
		NewJSONObjectCache(d.ctx, NewMemcacheReadWritable(d.ctx, 48*time.Hour)),
		aeCachedStore{
			ctx:        d.ctx,
			entityName: "TestRun",
		})
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
	}

	if err != nil {
		return nil, err
	}
	// Append the keys as ID
	for i, kbp := range keysByProduct {
		result[i].TestRuns.SetTestRunIDs(GetTestRunIDs(kbp.Keys))
	}
	return result, err
}

type aeQuery struct {
	query *datastore.Query
}

func (q aeQuery) Filter(filterStr string, value interface{}) Query {
	return aeQuery{q.query.Filter(filterStr, value)}
}

func (q aeQuery) Project(project string) Query {
	return aeQuery{q.query.Project(project)}
}

func (q aeQuery) Offset(offset int) Query {
	return aeQuery{q.query.Offset(offset)}
}

func (q aeQuery) Limit(limit int) Query {
	return aeQuery{q.query.Limit(limit)}
}

func (q aeQuery) Order(order string) Query {
	return aeQuery{q.query.Order(order)}
}

func (q aeQuery) KeysOnly() Query {
	return aeQuery{q.query.KeysOnly()}
}

func (q aeQuery) Distinct() Query {
	return aeQuery{q.query.Distinct()}
}

func (q aeQuery) Run(store Datastore) Iterator {
	return aeIterator{
		iter: q.query.Run(store.(aeDatastore).ctx),
	}
}

type aeIterator struct {
	iter *datastore.Iterator
}

func (i aeIterator) Next(dst interface{}) (Key, error) {
	return i.iter.Next(dst)
}
