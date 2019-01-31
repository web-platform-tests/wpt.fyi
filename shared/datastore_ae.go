// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"

	"google.golang.org/appengine/datastore"
)

// NewAppEngineDatastore creates a Datastore implementation, or a Datastore
// implementation with Memcache in front to cache all TestRun reads if cached
// is true.
//
// Both variants are backed by the appengine libraries and to be used in
// AppEngine standard.
func NewAppEngineDatastore(ctx context.Context, cached bool) Datastore {
	if cached {
		return aeCachedDatastore{
			aeDatastore{ctx: ctx},
		}
	}
	return aeDatastore{ctx: ctx}
}

type aeDatastore struct {
	ctx context.Context
}

func (d aeDatastore) TestRunQuery() TestRunQuery {
	return testRunQueryImpl{store: d}
}

func (d aeDatastore) Context() context.Context {
	return d.ctx
}

func (d aeDatastore) Done() interface{} {
	return datastore.Done
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

func (d aeDatastore) Get(k Key, dst interface{}) error {
	return datastore.Get(d.ctx, k.(*datastore.Key), dst)
}

func (d aeDatastore) GetMulti(keys []Key, dst interface{}) error {
	cast := make([]*datastore.Key, len(keys))
	for i := range keys {
		cast[i] = keys[i].(*datastore.Key)
	}
	return datastore.GetMulti(d.ctx, cast, dst)
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
