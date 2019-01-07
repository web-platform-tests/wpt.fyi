// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	mapset "github.com/deckarep/golang-set"
)

type cloudKey struct {
	key *datastore.Key
}

func (k cloudKey) IntID() int64 {
	return k.key.ID
}

// NewCloudDatastore creates a Datastore implementation that is backed by a
// standard cloud datastore client (i.e. not running in AppEngine standard).
func NewCloudDatastore(ctx context.Context, client *datastore.Client) Datastore {
	return cloudDatastore{
		ctx:    ctx,
		client: client,
	}
}

type cloudDatastore struct {
	ctx    context.Context
	client *datastore.Client
}

func (d cloudDatastore) Context() context.Context {
	return d.ctx
}

func (d cloudDatastore) NewQuery(typeName string) Query {
	return cloudQuery{
		query: datastore.NewQuery(typeName),
	}
}

func (d cloudDatastore) NewKey(typeName string, id int64) Key {
	return cloudKey{
		key: datastore.IDKey(typeName, id, nil),
	}
}

func (d cloudDatastore) GetAll(q Query, dst interface{}) ([]Key, error) {
	keys, err := d.client.GetAll(d.ctx, q.(cloudQuery).query, dst)
	cast := make([]Key, len(keys))
	for i := range keys {
		cast[i] = cloudKey{key: keys[i]}
	}
	return cast, err
}

func (d cloudDatastore) GetMulti(keys []Key, dst interface{}) error {
	cast := make([]*datastore.Key, len(keys))
	for i := range keys {
		cast[i] = keys[i].(cloudKey).key
	}
	return d.client.GetMulti(d.ctx, cast, dst)
}

func (d cloudDatastore) LoadTestRun(id int64) (*TestRun, error) {
	return loadTestRun(d, id)
}

func (d cloudDatastore) LoadTestRuns(
	products []ProductSpec,
	labels mapset.Set,
	revisions []string,
	from *time.Time,
	to *time.Time,
	limit,
	offset *int) (result TestRunsByProduct, err error) {
	return loadTestRuns(d, products, labels, revisions, from, to, limit, offset)
}

type cloudQuery struct {
	query *datastore.Query
}

func (q cloudQuery) Filter(filterStr string, value interface{}) Query {
	return cloudQuery{q.query.Filter(filterStr, value)}
}

func (q cloudQuery) Project(project string) Query {
	return cloudQuery{q.query.Project(project)}
}

func (q cloudQuery) Offset(offset int) Query {
	return cloudQuery{q.query.Offset(offset)}
}

func (q cloudQuery) Limit(limit int) Query {
	return cloudQuery{q.query.Limit(limit)}
}

func (q cloudQuery) Order(order string) Query {
	return cloudQuery{q.query.Order(order)}
}

func (q cloudQuery) KeysOnly() Query {
	return cloudQuery{q.query.KeysOnly()}
}

func (q cloudQuery) Distinct() Query {
	return cloudQuery{q.query.Distinct()}
}

func (q cloudQuery) Run(store Datastore) Iterator {
	cStore := store.(cloudDatastore)
	return cloudIterator{
		iter: cStore.client.Run(cStore.ctx, q.query),
	}
}

type cloudIterator struct {
	iter *datastore.Iterator
}

func (i cloudIterator) Next(dst interface{}) (Key, error) {
	key, err := i.iter.Next(dst)
	return cloudKey{key}, err
}
