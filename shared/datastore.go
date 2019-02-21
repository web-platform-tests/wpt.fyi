// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/datastore_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared Datastore

package shared

import (
	"context"

	"google.golang.org/appengine/datastore"
)

// Key abstracts an int64 based datastore.Key
type Key interface {
	IntID() int64
	Kind() string // Type name, e.g. TestRun
}

// Iterator abstracts a datastore.Iterator
type Iterator interface {
	Next(dst interface{}) (Key, error)
}

// Query abstracts a datastore.Query
type Query interface {
	Filter(filterStr string, value interface{}) Query
	Project(project string) Query
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
	Get(key Key, dst interface{}) error
	GetAll(q Query, dst interface{}) ([]Key, error)
	GetMulti(keys []Key, dst interface{}) error
	Put(key Key, src interface{}) (Key, error)

	TestRunQuery() TestRunQuery
}

// GetFeatureFlags returns all feature flag defaults set in the datastore.
func GetFeatureFlags(ctx context.Context) (flags []Flag, err error) {
	var keys []*datastore.Key
	keys, err = datastore.NewQuery("Flag").GetAll(ctx, &flags)
	for i := range keys {
		flags[i].Name = keys[i].StringID()
	}
	return flags, err
}

// IsFeatureEnabled returns true if a feature with the given flag name exists,
// and Enabled is set to true.
func IsFeatureEnabled(ctx context.Context, flagName string) bool {
	key := datastore.NewKey(ctx, "Flag", flagName, 0, nil)
	flag := Flag{}
	if err := datastore.Get(ctx, key, &flag); err != nil {
		return false
	}
	return flag.Enabled
}

// SetFeature puts a feature with the given flag name and enabled state.
func SetFeature(ctx context.Context, flag Flag) error {
	key := datastore.NewKey(ctx, "Flag", flag.Name, 0, nil)
	_, err := datastore.Put(ctx, key, &flag)
	return err
}

// GetSecret is a helper wrapper for loading a token's secret from the datastore
// by name.
func GetSecret(ctx context.Context, tokenName string) (string, error) {
	key := datastore.NewKey(ctx, "Token", tokenName, 0, nil)
	var token Token
	if err := datastore.Get(ctx, key, &token); err != nil {
		return "", err
	}
	return token.Secret, nil
}
