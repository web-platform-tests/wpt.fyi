// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/test_run_query_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared TestRunQuery

package shared

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	mapset "github.com/deckarep/golang-set"
)

var errNoProducts = errors.New("No products specified in request to load test runs")

// TestRunQuery abstracts complex queries of TestRun entities.
type TestRunQuery interface {
	// LoadTestRuns loads the test runs for the TestRun entities for the given parameters.
	// It is encapsulated because we cannot run single queries with multiple inequality
	// filters, so must load the keys and merge the results.
	LoadTestRuns(
		products []ProductSpec,
		labels mapset.Set,
		revisions []string,
		from *time.Time,
		to *time.Time,
		limit,
		offset *int) (result TestRunsByProduct, err error)

	// LoadTestRunKeys loads the keys for the TestRun entities for the given parameters.
	// It is encapsulated because we cannot run single queries with multiple inequality
	// filters, so must load the keys and merge the results.
	LoadTestRunKeys(
		products []ProductSpec,
		labels mapset.Set,
		revisions []string,
		from *time.Time,
		to *time.Time,
		limit *int,
		offset *int) (result KeysByProduct, err error)

	// LoadTestRunsByKeys loads test runs by keys and sets their IDs.
	LoadTestRunsByKeys(KeysByProduct) (result TestRunsByProduct, err error)

	// GetAlignedRunSHAs returns an array of the SHA[0:10] for runs that
	// exists for all the given products, ordered by most-recent, as well as a map
	// of those SHAs to a KeysByProduct map of products to the TestRun keys, for the
	// runs in the aligned run.
	GetAlignedRunSHAs(
		products ProductSpecs,
		labels mapset.Set,
		from,
		to *time.Time,
		limit *int,
		offset *int) (shas []string, keys map[string]KeysByProduct, err error)
}

type testRunQueryImpl struct {
	store Datastore
}

// NewTestRunQuery creates a concrete TestRunQuery backed by a Datastore interface.
func NewTestRunQuery(store Datastore) TestRunQuery {
	return testRunQueryImpl{store}
}

func (t testRunQueryImpl) LoadTestRuns(
	products []ProductSpec,
	labels mapset.Set,
	revisions []string,
	from *time.Time,
	to *time.Time,
	limit,
	offset *int) (result TestRunsByProduct, err error) {
	if len(products) == 0 {
		return nil, errNoProducts
	}

	keys, err := t.LoadTestRunKeys(products, labels, revisions, from, to, limit, offset)
	if err != nil {
		return nil, err
	}
	return t.LoadTestRunsByKeys(keys)
}

func (t testRunQueryImpl) LoadTestRunsByKeys(keysByProduct KeysByProduct) (result TestRunsByProduct, err error) {
	result = TestRunsByProduct{}
	for _, kbp := range keysByProduct {
		runs := make(TestRuns, len(kbp.Keys))
		if err := t.store.GetMulti(kbp.Keys, runs); err != nil {
			return nil, err
		}
		result = append(result, ProductTestRuns{
			Product:  kbp.Product,
			TestRuns: runs,
		})
	}

	// Append the keys as ID
	for i, kbp := range keysByProduct {
		result[i].TestRuns.SetTestRunIDs(GetTestRunIDs(kbp.Keys))
	}
	return result, err
}

func (t testRunQueryImpl) LoadTestRunKeys(
	products []ProductSpec,
	labels mapset.Set,
	revisions []string,
	from *time.Time,
	to *time.Time,
	limit *int,
	offset *int) (result KeysByProduct, err error) {
	log := GetLogger(t.store.Context())
	result = make(KeysByProduct, len(products))
	baseQuery := t.store.NewQuery("TestRun")
	if offset != nil {
		baseQuery = baseQuery.Offset(*offset)
	}
	if labels != nil {
		labels.Remove("") // Ensure the empty string isn't present.
		for i := range labels.Iter() {
			baseQuery = baseQuery.Filter("Labels =", i.(string))
		}
	}
	var globalIDFilter mapset.Set
	if len(revisions) > 1 || len(revisions) == 1 && !IsLatest(revisions[0]) {
		globalIDFilter = mapset.NewSet()
		for _, sha := range revisions {
			var ids TestRunIDs
			if ids, err = loadKeysForRevision(t.store, baseQuery, sha); err != nil {
				return nil, err
			}
			for _, id := range ids {
				globalIDFilter.Add(id)
			}
		}
		log.Debugf("Found %d keys across %d revisions", globalIDFilter.Cardinality(), len(revisions))
	}

	for i, product := range products {
		var productIDFilter = merge(globalIDFilter, nil)
		query := baseQuery.Filter("BrowserName =", product.BrowserName)
		if product.Labels != nil {
			for i := range product.Labels.Iter() {
				query = query.Filter("Labels =", i.(string))
			}
		}
		if !IsLatest(product.Revision) {
			var ids TestRunIDs
			if ids, err = loadKeysForRevision(t.store, query, product.Revision); err != nil {
				return nil, err
			}
			revIDFilter := mapset.NewSet()
			for _, id := range ids {
				revIDFilter.Add(id)
			}
			log.Debugf("Found %v keys for %s@%s", revIDFilter.Cardinality(), product.BrowserName, product.Revision)
			productIDFilter = merge(productIDFilter, revIDFilter)
		}
		if product.BrowserVersion != "" {
			var versionIDs mapset.Set
			if versionIDs, err = loadKeysForBrowserVersion(t.store, query, product.BrowserVersion); err != nil {
				return nil, err
			}
			log.Debugf("Found %v keys for %s", versionIDs.Cardinality(), product.BrowserVersion)
			productIDFilter = merge(productIDFilter, versionIDs)
		}

		// If we have a specific set of possibilities, it's much cheaper to
		// turn the query on its head (filter the entities).
		var keys []Key
		if productIDFilter != nil {
			log.Debugf("Loading %v viable runs to filter them.", productIDFilter.Cardinality())
			keys = make([]Key, 0, productIDFilter.Cardinality())
			for key := range productIDFilter.Iter() {
				keys = append(keys, t.store.NewIDKey("TestRun", key.(int64)))
			}
			runs := make(TestRuns, len(keys))
			err = t.store.GetMulti(keys, runs)
			if err != nil {
				return nil, err
			}
			runs.SetTestRunIDs(GetTestRunIDs(keys))
			// TestRuns sorted by TimeStart asc by default
			sort.Sort(sort.Reverse(runs))
			keys = make([]Key, 0)
			for _, run := range runs {
				if !product.Matches(run) ||
					from != nil && !from.Before(run.TimeStart) ||
					to != nil && !run.TimeStart.Before(*to) {
					continue
				}
				keys = append(keys, t.store.NewIDKey("TestRun", run.ID))
			}
			if limit != nil && len(keys) >= *limit {
				keys = keys[:*limit]
			} else if len(keys) >= MaxCountMaxValue {
				keys = keys[:MaxCountMaxValue]
			}
		} else {
			// Otherwise, just run a "GetAll" filter. Expensive.
			log.Debugf("Falling back to GetAll datastore query.")
			// TODO(lukebjerring): Indexes + filtering for OS + version.
			query = query.Order("-TimeStart")
			if from != nil {
				query = query.Filter("TimeStart >=", *from)
			}
			if to != nil {
				query = query.Filter("TimeStart <", *to)
			}
			max := MaxCountMaxValue
			if limit != nil && *limit < MaxCountMaxValue {
				max = *limit
			}
			keys, err = t.store.GetAll(query.KeysOnly().Limit(max), nil)
			if err != nil {
				return nil, err
			}
			log.Debugf("Loaded %v results for %s", len(keys), product.String())
		}

		log.Debugf("Found %v results for %s", len(keys), product.String())
		result[i] = ProductTestRunKeys{
			Product: product,
			Keys:    keys,
		}
	}
	return result, nil
}

func (t testRunQueryImpl) GetAlignedRunSHAs(
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
	query := t.store.
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
	it := query.Run(t.store)
	for {
		var testRun TestRun
		var key Key
		matchingProduct := -1
		key, err := it.Next(&testRun)
		if err == t.store.Done() {
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
	log := GetLogger(store.Context())
	var revQuery Query
	if len(sha) < 40 {
		log.Debugf("Finding revisions %s <= SHA < %s", sha, sha+"g")
		revQuery = query.
			Order("FullRevisionHash").
			Limit(MaxCountMaxValue).
			Filter("FullRevisionHash >=", sha).
			Filter("FullRevisionHash <", sha+"g") // g > f
	} else {
		log.Debugf("Finding exact revision %s", sha)
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

func getTestRunMemcacheKey(id int64) string {
	return "TEST_RUN-" + strconv.FormatInt(id, 10)
}
