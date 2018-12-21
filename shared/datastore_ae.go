// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"

	"google.golang.org/appengine/datastore"
)

func NewAppEngineDatastore(ctx context.Context) Datastore {
	return aeDatastore{
		ctx: ctx,
	}
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

func (d aeDatastore) GetAll(q Query, dst interface{}) ([]Key, error) {
	keys, err := q.(aeQuery).query.GetAll(d.ctx, dst)
	cast := make([]Key, len(keys))
	for i := range keys {
		cast[i] = keys[i]
	}
	return cast, err
}

type aeQuery struct {
	query *datastore.Query
}

func (q aeQuery) Filter(filterStr string, value interface{}) Query {
	return aeQuery{query: q.query.Filter(filterStr, value)}
}

func (q aeQuery) Project(project string) Query {
	return aeQuery{query: q.query.Project(project)}
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
	return aeQuery{query: q.query.KeysOnly()}
}

func (q aeQuery) Distinct() Query {
	return aeQuery{query: q.query.Distinct()}
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
