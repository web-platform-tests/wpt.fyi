// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"

	"google.golang.org/appengine/datastore"
)

// Key abstracts an int64 based datastore.Key
type Key interface {
	IntID() int64
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
	NewQuery(typeName string) Query
	GetAll(q Query, dst interface{}) ([]Key, error)
}

// LoadTestRun loads the TestRun entity for the given key.
func LoadTestRun(store Datastore, id int64) (*TestRun, error) {
	var testRun TestRun
	ctx := store.Context()
	cs := NewObjectCachedStore(ctx, NewJSONObjectCache(ctx, NewMemcacheReadWritable(ctx, 48*time.Hour)), NewDatastoreObjectStore(ctx, "TestRun"))
	err := cs.Get(getTestRunMemcacheKey(id), id, &testRun)
	if err != nil {
		return nil, err
	}

	testRun.ID = id
	return &testRun, nil
}

// LoadTestRunsBySHAs loads all test runs that belong to any of the given revisions (SHAs).
func LoadTestRunsBySHAs(store Datastore, shas ...string) (runs TestRuns, err error) {
	for _, sha := range shas {
		if len(sha) > 10 {
			sha = sha[:10]
		}
		q := store.NewQuery("TestRun")
		ids, err := loadKeysForRevision(store, q, sha)
		if err != nil {
			return runs, err
		}
		shaRuns := make(TestRuns, len(ids))
		for i := range ids {
			run, err := LoadTestRun(store, ids[i])
			if err != nil {
				return nil, err
			}
			shaRuns[i] = *run
		}
		for i := range ids {
			shaRuns[i].ID = ids[i]
		}
		runs = append(runs, shaRuns...)
	}
	return runs, err
}

// LoadTestRunKeys loads the keys for the TestRun entities for the given parameters.
// It is encapsulated because we cannot run single queries with multiple inequality
// filters, so must load the keys and merge the results.
func LoadTestRunKeys(
	store Datastore,
	products []ProductSpec,
	labels mapset.Set,
	sha string,
	from *time.Time,
	to *time.Time,
	limit *int,
	offset *int) (result KeysByProduct, err error) {
	result = make(KeysByProduct, len(products))
	baseQuery := store.NewQuery("TestRun")
	if offset != nil {
		baseQuery = baseQuery.Offset(*offset)
	}
	if labels != nil {
		labels.Remove("") // Ensure the empty string isn't present.
		for i := range labels.Iter() {
			baseQuery = baseQuery.Filter("Labels =", i.(string))
		}
	}
	var globalKeyFilter mapset.Set
	if !IsLatest(sha) {
		var ids TestRunIDs
		if ids, err = loadKeysForRevision(store, baseQuery, sha); err != nil {
			return nil, err
		}
		globalKeyFilter = mapset.NewSet()
		for _, id := range ids {
			globalKeyFilter.Add(id)
		}
	}
	for i, product := range products {
		var productKeyFilter = merge(globalKeyFilter, nil)
		query := baseQuery.Filter("BrowserName =", product.BrowserName)
		if product.Labels != nil {
			for i := range product.Labels.Iter() {
				query = query.Filter("Labels =", i.(string))
			}
		}
		if !IsLatest(product.Revision) {
			var ids TestRunIDs
			if ids, err = loadKeysForRevision(store, query, product.Revision); err != nil {
				return nil, err
			}
			revKeyFilter := mapset.NewSet()
			for _, id := range ids {
				revKeyFilter.Add(id)
			}
			productKeyFilter = merge(productKeyFilter, revKeyFilter)
		}
		if product.BrowserVersion != "" {
			var versionKeys mapset.Set
			if versionKeys, err = loadKeysForBrowserVersion(store, query, product.BrowserVersion); err != nil {
				return nil, err
			}
			productKeyFilter = merge(productKeyFilter, versionKeys)
		}
		// TODO(lukebjerring): Indexes + filtering for OS + version.
		query = query.Order("-TimeStart")

		if from != nil {
			query = query.Filter("TimeStart >=", *from)
		}
		if to != nil {
			query = query.Filter("TimeStart <", *to)
		}

		var keys []Key
		iter := query.KeysOnly().Run(store)
		for {
			key, err := iter.Next(nil)
			if err == datastore.Done {
				break
			} else if err != nil {
				return result, err
			} else if (limit != nil && len(keys) >= *limit) || len(keys) >= MaxCountMaxValue {
				break
			} else if productKeyFilter != nil && !productKeyFilter.Contains(key.IntID()) {
				continue
			}
			keys = append(keys, key)
		}
		result[i] = ProductTestRunKeys{
			Product: product,
			Keys:    keys,
		}
	}
	return result, nil
}

func merge(s1, s2 mapset.Set) mapset.Set {
	if s1 == nil && s2 == nil {
		return nil
	} else if s1 == nil {
		return merge(s2, nil)
	} else if s2 == nil {
		return mapset.NewSetWith(s1.ToSlice()...)
	}
	return s1.Intersect(s2)
}

// LoadTestRuns loads the test runs for the TestRun entities for the given parameters.
// It is encapsulated because we cannot run single queries with multiple inequality
// filters, so must load the keys and merge the results.
func LoadTestRuns(
	store Datastore,
	products []ProductSpec,
	labels mapset.Set,
	sha string,
	from *time.Time,
	to *time.Time,
	limit,
	offset *int) (result TestRunsByProduct, err error) {
	keys, err := LoadTestRunKeys(store, products, labels, sha, from, to, limit, offset)
	if err != nil {
		return nil, err
	}
	return LoadTestRunsByKeys(store, keys)
}

// LoadTestRunsByKeys loads the given test runs (by key), but also appends the
// ID to the TestRun entity.
func LoadTestRunsByKeys(store Datastore, keysByProduct KeysByProduct) (result TestRunsByProduct, err error) {
	result = TestRunsByProduct{}
	ctx := store.Context()
	cs := NewObjectCachedStore(ctx, NewJSONObjectCache(ctx, NewMemcacheReadWritable(ctx, 48*time.Hour)), NewDatastoreObjectStore(ctx, "TestRun"))
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

func contains(s []string, x string) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}

// Loads any keys for a revision prefix or full string match
func loadKeysForRevision(store Datastore, query Query, sha string) (result TestRunIDs, err error) {
	var revQuery Query
	if len(sha) < 40 {
		revQuery = query.
			Order("FullRevisionHash").
			Limit(MaxCountMaxValue).
			Filter("FullRevisionHash >=", sha).
			Filter("FullRevisionHash <", sha+"g") // g > f
	} else {
		revQuery = query.Filter("FullRevisionHash =", sha[:40])
	}

	var keys []Key
	if keys, err = store.GetAll(revQuery.KeysOnly(), nil); err != nil {
		return nil, err
	}
	return GetTestRunIDs(keys), nil
}

// Loads any keys for a full string match or a version prefix (Between [version].* and [version].9*).
// Entries in the set are the int64 value of the keys.
func loadKeysForBrowserVersion(store Datastore, query Query, version string) (result mapset.Set, err error) {
	versionQuery := VersionPrefix(query, "BrowserVersion", version, true)
	var keys []Key
	keyset := mapset.NewSet()
	if keys, err = store.GetAll(versionQuery.KeysOnly(), nil); err != nil {
		return nil, err
	}
	for _, key := range keys {
		keyset.Add(key.IntID())
	}
	if keys, err = store.GetAll(query.Filter("BrowserVersion =", version).KeysOnly(), nil); err != nil {
		return nil, err
	}
	for _, key := range keys {
		keyset.Add(key.IntID())
	}
	return keyset, nil
}

// VersionPrefix returns the given query with a prefix filter on the given
// field name, using the >= and < filters.
func VersionPrefix(query Query, fieldName, versionPrefix string, desc bool) Query {
	order := fieldName
	if desc {
		order = "-" + order
	}
	return query.
		Limit(MaxCountMaxValue).
		Order(order).
		Filter(fieldName+" >=", fmt.Sprintf("%s.", versionPrefix)).
		Filter(fieldName+" <=", fmt.Sprintf("%s.%c", versionPrefix, '9'+1))
}

// GetAlignedRunSHAs returns an array of the SHA[0:10] for runs that
// exists for all the given products, ordered by most-recent, as well as a map
// of those SHAs to a KeysByProduct map of products to the TestRun keys, for the
// runs in the aligned run.
func GetAlignedRunSHAs(
	store Datastore,
	products ProductSpecs,
	labels mapset.Set,
	from,
	to *time.Time,
	limit *int,
	offset *int) (shas []string, keys map[string]KeysByProduct, err error) {
	if limit == nil {
		maxMax := MaxCountMaxValue
		limit = &maxMax
	}
	query := store.
		NewQuery("TestRun").
		Order("-TimeStart")

	if labels != nil {
		for i := range labels.Iter() {
			query = query.Filter("Labels =", i.(string))
		}
	}
	if from != nil {
		query = query.Filter("TimeStart >=", *from)
	}
	if to != nil {
		query = query.Filter("TimeStart <", *to)
	}

	productsBySHA := make(map[string]mapset.Set)
	keyCollector := make(map[string]KeysByProduct)
	keys = make(map[string]KeysByProduct)
	done := mapset.NewSet()
	it := query.Run(store)
	for {
		var testRun TestRun
		var key Key
		matchingProduct := -1
		key, err := it.Next(&testRun)
		if err == datastore.Done {
			break
		} else if err != nil {
			return nil, nil, err
		} else {
			for i := range products {
				if products[i].Matches(testRun) {
					matchingProduct = i
					break
				}
			}
		}
		if matchingProduct < 0 {
			continue
		}
		if _, ok := productsBySHA[testRun.Revision]; !ok {
			productsBySHA[testRun.Revision] = mapset.NewSet()
			keyCollector[testRun.Revision] = make(KeysByProduct, len(products))
		}
		set := productsBySHA[testRun.Revision]
		if set.Contains(matchingProduct) {
			continue
		}
		set.Add(matchingProduct)
		keyCollector[testRun.Revision][matchingProduct].Keys = []Key{key}
		if set.Cardinality() == len(products) && !done.Contains(testRun.Revision) {
			if offset == nil || done.Cardinality() >= *offset {
				shas = append(shas, testRun.Revision)
			}
			done.Add(testRun.Revision)
			keys[testRun.Revision] = keyCollector[testRun.Revision]
			if len(shas) >= *limit {
				return shas, keys, nil
			}
		}
	}
	return shas, keys, err
}

func getTestRunMemcacheKey(id int64) string {
	return "TEST_RUN-" + strconv.FormatInt(id, 10)
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
