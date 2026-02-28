// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"errors"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

type cloudKey struct {
	key *datastore.Key
}

func (k cloudKey) IntID() int64 {
	return k.key.ID
}

func (k cloudKey) StringID() string {
	return k.key.Name
}

func (k cloudKey) Kind() string {
	return k.key.Kind
}

// NewAppEngineDatastore creates a Datastore implementation, or a Datastore
// implementation with Redis in front to cache all TestRun reads if cached
// is true.
//
// Both variants (cached or not) are backed by Cloud Datastore SDK, using
// Clients initialized at startup in webapp.
func NewAppEngineDatastore(ctx context.Context, cached bool) Datastore {
	ds := cloudDatastore{
		ctx:    ctx,
		client: Clients.datastore,
	}
	if cached {
		return cachedDatastore{ds, ctx}
	}
	return ds
}

// NewCloudDatastore creates a Datastore implementation that is backed by a
// given Cloud Datastore client.
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

func (d cloudDatastore) TestRunQuery() TestRunQuery {
	return testRunQueryImpl{store: d}
}

func (d cloudDatastore) Context() context.Context {
	return d.ctx
}

func (d cloudDatastore) Done() interface{} {
	return iterator.Done
}

func (d cloudDatastore) NewQuery(typeName string) Query {
	return cloudQuery{
		query: datastore.NewQuery(typeName),
	}
}

func (d cloudDatastore) NewIDKey(typeName string, id int64) Key {
	return cloudKey{
		key: datastore.IDKey(typeName, id, nil),
	}
}

func (d cloudDatastore) NewIncompleteKey(typeName string) Key {
	return cloudKey{
		key: datastore.IncompleteKey(typeName, nil),
	}
}

func (d cloudDatastore) ReserveID(typeName string) (Key, error) {
	keys, err := d.client.AllocateIDs(d.ctx, []*datastore.Key{datastore.IncompleteKey(typeName, nil)})
	if err != nil {
		return nil, err
	} else if len(keys) < 1 {
		return nil, errors.New("failed to create a key")
	}
	return cloudKey{
		key: keys[0],
	}, nil
}

func (d cloudDatastore) NewNameKey(typeName string, name string) Key {
	return cloudKey{
		key: datastore.NameKey(typeName, name, nil),
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

func (d cloudDatastore) Get(k Key, dst interface{}) error {
	cast := k.(cloudKey).key
	err := d.client.Get(d.ctx, cast, dst)
	if err == datastore.ErrNoSuchEntity {
		return ErrNoSuchEntity
	}
	return err
}

func (d cloudDatastore) GetMulti(keys []Key, dst interface{}) error {
	cast := make([]*datastore.Key, len(keys))
	for i := range keys {
		cast[i] = keys[i].(cloudKey).key
	}
	err := d.client.GetMulti(d.ctx, cast, dst)
	if multiError, ok := err.(datastore.MultiError); ok {
		errors := make([]error, len(multiError))
		for i, err := range multiError {
			if err == datastore.ErrNoSuchEntity {
				errors[i] = ErrNoSuchEntity
			} else {
				errors[i] = err
			}
		}
		return NewMultiError(errors, "datastore.GetMulti")
	}
	return err
}

func (d cloudDatastore) Put(key Key, src interface{}) (Key, error) {
	newkey, err := d.client.Put(d.ctx, key.(cloudKey).key, src)
	return cloudKey{newkey}, err
}

func (d cloudDatastore) PutMulti(keys []Key, src interface{}) ([]Key, error) {
	cast := make([]*datastore.Key, len(keys))
	for i := range keys {
		cast[i] = keys[i].(cloudKey).key
	}

	srcKeys, err := d.client.PutMulti(d.ctx, cast, src)
	newKeys := make([]Key, len(srcKeys))
	for i := range srcKeys {
		newKeys[i] = cloudKey{srcKeys[i]}
	}
	return newKeys, err
}

func (d cloudDatastore) Insert(key Key, src interface{}) error {
	_, err := d.client.RunInTransaction(d.ctx, func(txn *datastore.Transaction) error {
		var empty map[string]interface{}
		err := txn.Get(key.(cloudKey).key, &empty)
		if err == nil {
			return ErrEntityAlreadyExists
		} else if err != datastore.ErrNoSuchEntity {
			return err
		}
		_, err = txn.Put(key.(cloudKey).key, src)
		return err
	})
	return err
}

func (d cloudDatastore) Update(key Key, dst interface{}, mutator func(obj interface{}) error) error {
	_, err := d.client.RunInTransaction(d.ctx, func(txn *datastore.Transaction) error {
		if err := txn.Get(key.(cloudKey).key, dst); err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}
		if err := mutator(dst); err != nil {
			return err
		}
		_, err := txn.Put(key.(cloudKey).key, dst)
		return err
	})
	return err
}

type cloudQuery struct {
	query *datastore.Query
}

func (q cloudQuery) FilterBuilder() FilterBuilder {
	return cloudFilterBuilder{}
}

func (q cloudQuery) FilterEntity(entityFilter EntityFilter) Query {
	return cloudQuery{q.query.FilterEntity(entityFilter)}
}

func (q cloudQuery) Filter(filterStr string, value interface{}) Query {
	return cloudQuery{q.query.Filter(filterStr, value)}
}

func (q cloudQuery) Project(fields ...string) Query {
	return cloudQuery{q.query.Project(fields...)}
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

// EntityFilter wraps datastore.EntityFilter.
// datastore.EntityFilter does not expose any methods. But using this type
// allows us to be strict on the filters returned by the FilterBuilder.
type EntityFilter interface {
	datastore.EntityFilter
}

type cloudFilterBuilder struct{}

func (b cloudFilterBuilder) PropertyFilter(FieldName string, Operator string, Value interface{}) EntityFilter {
	return datastore.PropertyFilter{FieldName: FieldName, Operator: Operator, Value: Value}
}
