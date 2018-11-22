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

// LoadTestRun loads the TestRun entity for the given key.
func LoadTestRun(ctx context.Context, id int64) (*TestRun, error) {
	var testRun TestRun
	cs := NewObjectCachedStore(ctx, NewJSONObjectCache(ctx, NewMemcacheReadWritable(ctx, 48*time.Hour)), NewDatastoreObjectStore(ctx, "TestRun"))
	err := cs.Get(getTestRunMemcacheKey(id), id, &testRun)
	if err != nil {
		return nil, err
	}

	testRun.ID = id
	return &testRun, nil
}

// LoadTestRunKeys loads the keys for the TestRun entities for the given parameters.
// It is encapsulated because we cannot run single queries with multiple inequality
// filters, so must load the keys and merge the results.
func LoadTestRunKeys(
	ctx context.Context,
	products []ProductSpec,
	labels mapset.Set,
	sha string,
	from *time.Time,
	to *time.Time,
	limit *int,
	offset *int) (result KeysByProduct, err error) {
	result = make(KeysByProduct, len(products))
	baseQuery := datastore.NewQuery("TestRun")
	if !IsLatest(sha) {
		baseQuery = baseQuery.Filter("Revision =", sha)
	}
	if offset != nil {
		baseQuery = baseQuery.Offset(*offset)
	}
	if labels != nil {
		for i := range labels.Iter() {
			baseQuery = baseQuery.Filter("Labels =", i.(string))
		}
	}
	for _, product := range products {
		var prefiltered *mapset.Set
		query := baseQuery.Filter("BrowserName =", product.BrowserName)
		if product.Labels != nil {
			for i := range product.Labels.Iter() {
				query = query.Filter("Labels =", i.(string))
			}
		}
		if !IsLatest(product.Revision) {
			query = query.Filter("Revision = ", product.Revision)
		}
		if product.BrowserVersion != "" {
			if prefiltered, err = loadKeysForBrowserVersion(ctx, query, product.BrowserVersion); err != nil {
				return nil, err
			}
		}
		// TODO(lukebjerring): Indexes + filtering for OS + version.
		query = query.Order("-TimeStart")

		if from != nil {
			query = query.Filter("TimeStart >=", *from)
		}
		if to != nil {
			query = query.Filter("TimeStart <", *to)
		}

		var keys []*datastore.Key
		iter := query.KeysOnly().Run(ctx)
		for {
			key, err := iter.Next(nil)
			if err == datastore.Done {
				break
			} else if err != nil {
				return result, err
			} else if (limit != nil && len(keys) >= *limit) || len(keys) >= MaxCountMaxValue {
				break
			} else if prefiltered != nil && !(*prefiltered).Contains(key.String()) {
				continue
			}
			keys = append(keys, key)
		}
		result = append(result, ProductTestRunKeys{
			Product: product,
			Keys:    keys,
		})
	}
	return result, nil
}

// LoadTestRuns loads the test runs for the TestRun entities for the given parameters.
// It is encapsulated because we cannot run single queries with multiple inequality
// filters, so must load the keys and merge the results.
func LoadTestRuns(
	ctx context.Context,
	products []ProductSpec,
	labels mapset.Set,
	sha string,
	from *time.Time,
	to *time.Time,
	limit,
	offset *int) (result TestRunsByProduct, err error) {
	keys, err := LoadTestRunKeys(ctx, products, labels, sha, from, to, limit, offset)
	if err != nil {
		return nil, err
	}
	return LoadTestRunsByKeys(ctx, keys)
}

// LoadTestRunsByKeys loads the given test runs (by key), but also appends the
// ID to the TestRun entity.
func LoadTestRunsByKeys(ctx context.Context, keysByProduct KeysByProduct) (result TestRunsByProduct, err error) {
	result = TestRunsByProduct{}
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
		for j, key := range kbp.Keys {
			result[i].TestRuns[j].ID = key.IntID()
		}
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

// Loads any keys for a full string match or a version prefix (Between [version].* and [version].9*)
func loadKeysForBrowserVersion(ctx context.Context, query *datastore.Query, version string) (result *mapset.Set, err error) {
	versionQuery := VersionPrefix(query, "BrowserVersion", version, true)
	var keys []*datastore.Key
	keyset := mapset.NewSet()
	if keys, err = versionQuery.KeysOnly().GetAll(ctx, nil); err != nil {
		return nil, err
	}
	for _, key := range keys {
		keyset.Add(key.String())
	}
	if keys, err = query.Filter("BrowserVersion =", version).KeysOnly().GetAll(ctx, nil); err != nil {
		return nil, err
	}
	for _, key := range keys {
		keyset.Add(key.String())
	}
	return &keyset, nil
}

// VersionPrefix returns the given query with a prefix filter on the given
// field name, using the >= and < filters.
func VersionPrefix(query *datastore.Query, fieldName, versionPrefix string, desc bool) *datastore.Query {
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
	ctx context.Context,
	products ProductSpecs,
	labels mapset.Set,
	from,
	to *time.Time,
	limit *int) (shas []string, keys map[string]KeysByProduct, err error) {
	query := datastore.
		NewQuery("TestRun").
		Limit(MaxCountMaxValue).
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
	it := query.Run(ctx)
	for {
		var testRun TestRun
		var key *datastore.Key
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
		keyCollector[testRun.Revision][matchingProduct].Keys = []*datastore.Key{key}
		if set.Cardinality() == len(products) && !done.Contains(testRun.Revision) {
			done.Add(testRun.Revision)
			shas = append(shas, testRun.Revision)
			keys[testRun.Revision] = keyCollector[testRun.Revision]
			if limit != nil && len(shas) >= *limit {
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
