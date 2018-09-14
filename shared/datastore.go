package shared

import (
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// LoadTestRun loads the TestRun entity for the given key.
func LoadTestRun(ctx context.Context, id int64) (*TestRun, error) {
	var testRun TestRun
	key := datastore.NewKey(ctx, "TestRun", "", id, nil)
	if err := datastore.Get(ctx, key, &testRun); err != nil {
		return nil, err
	}
	testRun.ID = key.IntID()
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
	limit *int) (result []*datastore.Key, err error) {
	baseQuery := datastore.NewQuery("TestRun").Limit(1000)
	if !IsLatest(sha) {
		baseQuery = baseQuery.Filter("Revision =", sha)
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

		fetched, err := query.KeysOnly().GetAll(ctx, nil)
		if err != nil {
			return nil, err
		}
		var keys []*datastore.Key
		for _, key := range fetched {
			if limit == nil || *limit > len(keys) {
				if prefiltered == nil || (*prefiltered).Contains(key.String()) {
					keys = append(keys, key)
				}
			}
		}

		if limit != nil && len(keys) > *limit {
			keys = keys[:*limit]
		}
		result = append(result, keys...)
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
	limit *int) (result []TestRun, err error) {
	keys, err := LoadTestRunKeys(ctx, products, labels, sha, from, to, limit)
	if err != nil {
		return nil, err
	}
	return LoadTestRunsByKeys(ctx, keys)
}

// LoadTestRunsByKeys loads the given test runs (by key), but also appends the
// ID to the TestRun entity.
func LoadTestRunsByKeys(ctx context.Context, keys []*datastore.Key) (result TestRuns, err error) {
	result = make(TestRuns, len(keys))
	err = datastore.GetMulti(ctx, keys, result)
	if err != nil {
		return nil, err
	}
	// Append the keys as ID
	for i, key := range keys {
		result[i].ID = key.IntID()
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
		Order(order).
		Filter(fieldName+" >=", fmt.Sprintf("%s.", versionPrefix)).
		Filter(fieldName+" <=", fmt.Sprintf("%s.%c", versionPrefix, '9'+1))
}

// GetAlignedRunSHAs returns an array of the SHA[0:10] for runs that
// exists for all the given products, ordered by most-recent, as well as a map
// of those SHAs to the TestRun keys for the complete run.
func GetAlignedRunSHAs(
	ctx context.Context,
	products ProductSpecs,
	labels mapset.Set,
	from,
	to *time.Time,
	limit *int) (shas []string, keys map[string][]*datastore.Key, err error) {
	query := datastore.
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
	keyCollector := make(map[string][]*datastore.Key)
	keys = make(map[string][]*datastore.Key)
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
			keyCollector[testRun.Revision] = make([]*datastore.Key, len(products))
		}
		set := productsBySHA[testRun.Revision]
		if set.Contains(products[matchingProduct]) {
			continue
		}
		set.Add(products[matchingProduct])
		keyCollector[testRun.Revision][matchingProduct] = key
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
