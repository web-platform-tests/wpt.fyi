// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/datastore_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared Datastore

package shared

import (
	"context"
	"errors"
)

// ErrEntityAlreadyExists is returned by Datastore.Insert when the entity already exists.
var ErrEntityAlreadyExists = errors.New("datastore: entity already exists")

// Key abstracts an int64 based datastore.Key
type Key interface {
	IntID() int64
	StringID() string
	Kind() string // Type name, e.g. TestRun
}

// Iterator abstracts a datastore.Iterator
type Iterator interface {
	Next(dst interface{}) (Key, error)
}

// Query abstracts a datastore.Query
type Query interface {
	Filter(filterStr string, value interface{}) Query
	Project(fields ...string) Query
	Limit(limit int) Query
	Offset(offset int) Query
	Order(order string) Query
	KeysOnly() Query
	Distinct() Query
	Run(Datastore) Iterator
}

// Datastore abstracts a datastore, hiding the distinctions between cloud and
// appengine's datastores.
type Datastore interface {
	Context() context.Context
	Done() interface{}
	NewQuery(typeName string) Query
	NewIDKey(typeName string, id int64) Key
	NewNameKey(typeName string, name string) Key
	ReserveID(typeName string) (Key, error)
	Get(key Key, dst interface{}) error
	GetAll(q Query, dst interface{}) ([]Key, error)
	GetMulti(keys []Key, dst interface{}) error
	Put(key Key, src interface{}) (Key, error)

	// Atomically insert a new entity.
	Insert(key Key, src interface{}) error
	// Atomically update or create an entity: the entity is first fetched
	// by key into dst, which must be a struct pointer; if the key cannot
	// be found, no error is returned and dst is not modified. Then
	// mutator(dst) is called; the transaction will be aborted if non-nil
	// error is returned. Finally, write dst back by key.
	Update(key Key, dst interface{}, mutator func(obj interface{}) error) error

	TestRunQuery() TestRunQuery
}

// GetFeatureFlags returns all feature flag defaults set in the datastore.
func GetFeatureFlags(ds Datastore) (flags []Flag, err error) {
	q := ds.NewQuery("Flag")
	keys, err := ds.GetAll(q, &flags)
	for i := range keys {
		flags[i].Name = keys[i].StringID()
	}
	return flags, err
}

// IsFeatureEnabled returns true if a feature with the given flag name exists,
// and Enabled is set to true.
func IsFeatureEnabled(ds Datastore, flagName string) bool {
	key := ds.NewNameKey("Flag", flagName)
	flag := Flag{}
	if err := ds.Get(key, &flag); err != nil {
		return false
	}
	return flag.Enabled
}

// SetFeature puts a feature with the given flag name and enabled state.
func SetFeature(ds Datastore, flag Flag) error {
	key := ds.NewNameKey("Flag", flag.Name)
	_, err := ds.Put(key, &flag)
	return err
}

// GetSecret is a helper wrapper for loading a token's secret from the datastore
// by name.
func GetSecret(ds Datastore, tokenName string) (string, error) {
	key := ds.NewNameKey("Token", tokenName)
	var token Token
	if err := ds.Get(key, &token); err != nil {
		return "", err
	}
	return token.Secret, nil
}

// GetUploader gets the Uploader by the given name.
func GetUploader(ds Datastore, uploader string) (Uploader, error) {
	var result Uploader
	key := ds.NewNameKey("Uploader", uploader)
	err := ds.Get(key, &result)
	return result, err
}
