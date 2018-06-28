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

// LoadTestRuns loads the TestRun entities for the given parameters.
// It is encapsulated because we cannot run single queries with multiple inequality
// filters, so must load the keys and merge the results.
func LoadTestRuns(
	ctx context.Context,
	products []ProductSpec,
	labels mapset.Set,
	shas []string,
	from *time.Time,
	limit *int) (result []TestRun, err error) {
	var testRuns []TestRun
	baseQuery := datastore.NewQuery("TestRun")
	// NOTE(lukebjerring): While we can't filter on multiple SHAs, it's still much more efficient
	// to (pre-)filter for a single SHA during the query.
	if len(shas) == 1 && !IsLatest(shas[0]) {
		baseQuery = baseQuery.Filter("Revision =", shas[0])
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
		if product.BrowserVersion != "" {
			if prefiltered, err = loadKeysForBrowserVersion(ctx, query, product.BrowserVersion); err != nil {
				return nil, err
			}
		}
		// TODO(lukebjerring): Indexes + filtering for OS + version.
		query = query.Order("-CreatedAt")

		if from != nil {
			query = query.Filter("CreatedAt >", *from)
		}

		fetched, err := query.KeysOnly().GetAll(ctx, nil)
		if err != nil {
			return nil, err
		}
		var keys []*datastore.Key
		for _, key := range fetched {
			if len(shas) > 1 || limit == nil || *limit > len(keys) {
				if prefiltered == nil || (*prefiltered).Contains(key.String()) {
					keys = append(keys, key)
				}
			}
		}
		testRunResults := make(TestRuns, len(keys))
		if err = datastore.GetMulti(ctx, keys, testRunResults); err != nil {
			return nil, err
		}
		// Append the keys as ID
		for i, key := range keys {
			testRunResults[i].ID = key.IntID()
		}
		appended := 0
		for _, testRun := range testRunResults {
			if len(shas) > 1 && !contains(shas, testRun.Revision) {
				continue
			}
			testRuns = append(testRuns, testRun)
			appended++
			if limit != nil && appended >= *limit {
				break
			}
		}
	}
	return testRuns, nil
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
