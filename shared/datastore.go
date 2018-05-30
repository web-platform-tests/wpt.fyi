package shared

import (
	"time"

	"github.com/deckarep/golang-set"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// LoadTestRuns loads the TestRun entities for the given parameters.
// It is encapsulated because we cannot run single queries with multiple inequality
// filters, so must load the keys and merge the results.
func LoadTestRuns(
	ctx context.Context,
	products []Product,
	sha string,
	from *time.Time,
	limit *int) (result []TestRun, err error) {
	var testRuns []TestRun
	baseQuery := datastore.NewQuery("TestRun")
	if sha != "" && sha != "latest" {
		baseQuery = baseQuery.Filter("Revision =", sha)
	}
	for _, product := range products {
		var prefiltered *mapset.Set
		query := baseQuery.Filter("BrowserName =", product.BrowserName)
		if product.BrowserVersion != "" {
			versionQuery := QueryPrefix(query, "BrowserVersion", product.BrowserVersion, true)
			keys, err := versionQuery.Order("-CreatedAt").KeysOnly().GetAll(ctx, nil)
			if err != nil {
				return nil, err
			}
			keyset := mapset.NewSet()
			for _, key := range keys {
				keyset.Add(key.String())
			}
			prefiltered = &keyset
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
			if (limit == nil || *limit > len(keys)) && (prefiltered == nil || (*prefiltered).Contains(key.String())) {
				keys = append(keys, key)
			}
		}
		testRunResults := make([]TestRun, len(keys))
		if err = datastore.GetMulti(ctx, keys, testRunResults); err != nil {
			return nil, err
		}
		testRuns = append(testRuns, testRunResults...)
	}
	return testRuns, nil
}
