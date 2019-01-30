// +build medium

package shared_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine/datastore"
)

func TestLoadTestRuns(t *testing.T) {
	testRun := shared.TestRun{
		ProductAtRevision: shared.ProductAtRevision{
			Product: shared.Product{
				BrowserName:    "chrome",
				BrowserVersion: "63.0",
				OSName:         "linux",
			},
			Revision: "1234567890",
		},
		ResultsURL: "/static/chrome-63.0-linux-summary.json.gz",
		CreatedAt:  time.Now(),
	}

	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
	key, _ = datastore.Put(ctx, key, &testRun)

	chrome, _ := shared.ParseProductSpec("chrome")
	store := shared.NewAppEngineDatastore(ctx)
	loaded, err := store.TestRunQuery().LoadTestRuns(shared.ProductSpecs{chrome}, nil, nil, nil, nil, nil, nil)
	allRuns := loaded.AllRuns()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allRuns))
	assert.Equalf(t, key.IntID(), allRuns[0].ID, "ID field should be populated.")
}

func TestLoadTestRunsBySHAs(t *testing.T) {
	testRun := shared.TestRun{}
	testRun.BrowserName = "chrome"
	testRun.BrowserVersion = "63.0"
	testRun.OSName = "linux"
	testRun.TimeStart = time.Now()

	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	for i := 0; i < 5; i++ {
		testRun.FullRevisionHash = strings.Repeat(strconv.Itoa(i), 40)
		testRun.Revision = testRun.FullRevisionHash[:10]
		testRun.TimeStart = time.Now().AddDate(0, 0, -i)
		key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
		datastore.Put(ctx, key, &testRun)
	}

	store := shared.NewAppEngineDatastore(ctx)
	q := store.TestRunQuery()
	runsByProduct, err := q.LoadTestRuns(shared.GetDefaultProducts(), nil, shared.SHAs{"1111111111", "333333333"}, nil, nil, nil, nil)
	runs := runsByProduct.AllRuns()
	assert.Nil(t, err)
	assert.Len(t, runs, 2)
	for _, run := range runs {
		assert.True(t, run.ID > 0, "ID field should be populated.")
	}
	assert.Equal(t, "1111111111", runs[0].Revision)
	assert.Equal(t, "3333333333", runs[1].Revision)

	runsByProduct, err = q.LoadTestRuns(shared.GetDefaultProducts(), nil, shared.SHAs{"11111", "33333"}, nil, nil, nil, nil)
	runs = runsByProduct.AllRuns()
	assert.Nil(t, err)
	assert.Len(t, runs, 2)
	for _, run := range runs {
		assert.True(t, run.ID > 0, "ID field should be populated.")
	}
	assert.Equal(t, "1111111111", runs[0].Revision)
	assert.Equal(t, "3333333333", runs[1].Revision)
}

func TestLoadTestRuns_Experimental_Only(t *testing.T) {
	testRuns := shared.TestRuns{
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "chrome",
					BrowserVersion: "63.0",
					OSName:         "linux",
				},
				Revision: "1234567890",
			},
			ResultsURL: "/static/chrome-63.0-linux-summary.json.gz",
			CreatedAt:  time.Now(),
		},
		shared.TestRun{
			// We no longer have runs like this (with the "-experimental" suffix but without
			// the "experimental" label; and it should no longer be considered experimental.
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "chrome-experimental",
					BrowserVersion: "63.0",
					OSName:         "linux",
				},
				Revision: "1234567890",
			},
			ResultsURL: "/static/chrome-experimental-63.0-linux-summary.json.gz",
			CreatedAt:  time.Now(),
		},
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "chrome",
					BrowserVersion: "64.0",
					OSName:         "linux",
				},
				Revision: "1234567890",
			},
			ResultsURL: "/static/chrome-64.0-linux-summary.json.gz",
			CreatedAt:  time.Now(),
			Labels:     []string{"experimental"},
		},
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "chrome-experimental",
					BrowserVersion: "65.0",
					OSName:         "linux",
				},
				Revision: "1234567890",
			},
			ResultsURL: "/static/chrome-experimental-65.0-linux-summary.json.gz",
			CreatedAt:  time.Now(),
			Labels:     []string{"experimental"},
		},
	}

	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	keys := make([]*datastore.Key, len(testRuns))
	for i := range testRuns {
		keys[i] = datastore.NewIncompleteKey(ctx, "TestRun", nil)
	}
	keys, err = datastore.PutMulti(ctx, keys, testRuns)
	assert.Nil(t, err)

	chrome, chromeExperimental := shared.ProductSpec{}, shared.ProductSpec{}
	chrome.BrowserName = "chrome"
	chromeExperimental.BrowserName = "chrome-experimental"
	products := shared.ProductSpecs{chrome, chromeExperimental}
	labels := mapset.NewSet()
	labels.Add("experimental")
	ten := 10
	store := shared.NewAppEngineDatastore(ctx)
	loaded, err := store.TestRunQuery().LoadTestRuns(products, labels, nil, nil, nil, &ten, nil)
	allRuns := loaded.AllRuns()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(allRuns))
	for _, run := range allRuns {
		assert.True(t, run.ID > 0, "ID field should be populated.")
	}
	assert.Equal(t, "64.0", allRuns[0].BrowserVersion)
	assert.Equal(t, "65.0", allRuns[1].BrowserVersion)
}

func TestLoadTestRuns_LabelinProductSpec(t *testing.T) {
	testRuns := []shared.TestRun{
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{BrowserName: "chrome"},
			},
			Labels: []string{"foo"},
		},
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{BrowserName: "chrome"},
			},
			Labels: []string{"bar"},
		},
	}

	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	keys := make([]*datastore.Key, len(testRuns))
	for i := range testRuns {
		keys[i] = datastore.NewIncompleteKey(ctx, "TestRun", nil)
	}
	keys, err = datastore.PutMulti(ctx, keys, testRuns)
	assert.Nil(t, err)

	products := make([]shared.ProductSpec, 1)
	products[0].BrowserName = "chrome"
	products[0].Labels = mapset.NewSetWith("foo")
	store := shared.NewAppEngineDatastore(ctx)
	loaded, err := store.TestRunQuery().LoadTestRuns(products, nil, nil, nil, nil, nil, nil)
	allRuns := loaded.AllRuns()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allRuns))
	assert.Equal(t, "foo", allRuns[0].Labels[0])
}

func TestLoadTestRuns_SHAinProductSpec(t *testing.T) {
	testRuns := []shared.TestRun{
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product:          shared.Product{BrowserName: "chrome"},
				FullRevisionHash: strings.Repeat("0", 40),
				Revision:         strings.Repeat("0", 10),
			},
		},
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "chrome",
					BrowserVersion: "63.1.1.1",
				},
				FullRevisionHash: strings.Repeat("1", 40),
				Revision:         strings.Repeat("1", 10),
			},
		},
	}

	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	keys := make([]*datastore.Key, len(testRuns))
	for i := range testRuns {
		keys[i] = datastore.NewIncompleteKey(ctx, "TestRun", nil)
	}
	keys, err = datastore.PutMulti(ctx, keys, testRuns)
	assert.Nil(t, err)

	products := make([]shared.ProductSpec, 1)
	products[0].BrowserName = "chrome"
	products[0].Revision = strings.Repeat("1", 10)
	store := shared.NewAppEngineDatastore(ctx)
	loaded, err := store.TestRunQuery().LoadTestRuns(products, nil, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	allRuns := loaded.AllRuns()
	assert.Equal(t, 1, len(allRuns))
	assert.Equal(t, "1111111111", allRuns[0].Revision)

	// Partial SHA
	products[0].Revision = "11111"
	loaded, err = store.TestRunQuery().LoadTestRuns(products, nil, nil, nil, nil, nil, nil)
	allRuns = loaded.AllRuns()
	assert.Equal(t, 1, len(allRuns))
	assert.Equal(t, "1111111111", allRuns[0].Revision)

	// Partial SHA, Browser version
	products[0].BrowserVersion = "63"
	loaded, err = store.TestRunQuery().LoadTestRuns(products, nil, nil, nil, nil, nil, nil)
	allRuns = loaded.AllRuns()
	assert.Equal(t, 1, len(allRuns))
	assert.Equal(t, "1111111111", allRuns[0].Revision)
}

func TestLoadTestRuns_Ordering(t *testing.T) {
	testRuns := []shared.TestRun{
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName: "chrome",
				},
				Revision: "1234567890",
			},
			CreatedAt: time.Now(),
			TimeStart: time.Now().AddDate(0, 0, -1),
		},
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName: "chrome",
				},
				Revision: "0987654321",
			},
			CreatedAt: time.Now().AddDate(0, 0, -1),
			TimeStart: time.Now(),
		},
	}

	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	for _, testRun := range testRuns {
		key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
		datastore.Put(ctx, key, &testRun)
	}

	chrome, _ := shared.ParseProductSpec("chrome")
	store := shared.NewAppEngineDatastore(ctx)
	loaded, err := store.TestRunQuery().LoadTestRuns(shared.ProductSpecs{chrome}, nil, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	allRuns := loaded.AllRuns()
	assert.Equal(t, 2, len(allRuns))
	// Runs should be ordered descendingly by TimeStart.
	assert.Equal(t, "0987654321", allRuns[0].Revision)
	assert.Equal(t, "1234567890", allRuns[1].Revision)
}

func TestLoadTestRuns_From(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastWeek := now.AddDate(0, 0, -7)
	testRuns := []shared.TestRun{
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName: "chrome",
				},
				Revision: "1234567890",
			},
			TimeStart: now,
		},
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName: "chrome",
				},
				Revision: "0987654321",
			},
			TimeStart: lastWeek,
		},
	}

	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	for _, testRun := range testRuns {
		key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
		datastore.Put(ctx, key, &testRun)
	}

	chrome, _ := shared.ParseProductSpec("chrome")
	store := shared.NewAppEngineDatastore(ctx)
	loaded, err := store.TestRunQuery().LoadTestRuns(shared.ProductSpecs{chrome}, nil, nil, &yesterday, nil, nil, nil)
	assert.Nil(t, err)
	allRuns := loaded.AllRuns()
	assert.Equal(t, 1, len(allRuns))
	assert.Equal(t, "1234567890", allRuns[0].Revision)
}

func TestLoadTestRuns_To(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	testRuns := []shared.TestRun{
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName: "chrome",
				},
				Revision: "1234567890",
			},
			TimeStart: now,
		},
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName: "chrome",
				},
				Revision: "0987654321",
			},
			TimeStart: yesterday,
		},
	}

	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	for _, testRun := range testRuns {
		key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
		datastore.Put(ctx, key, &testRun)
	}

	chrome, _ := shared.ParseProductSpec("chrome")
	store := shared.NewAppEngineDatastore(ctx)
	loaded, err := store.TestRunQuery().LoadTestRuns(shared.ProductSpecs{chrome}, nil, nil, nil, &now, nil, nil)
	assert.Nil(t, err)
	allRuns := loaded.AllRuns()
	assert.Equal(t, 1, len(allRuns))
	assert.Equal(t, "0987654321", allRuns[0].Revision)
}

func TestGetAlignedRunSHAs(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	browserNames := shared.GetDefaultBrowserNames()

	// Nothing in datastore.
	store := shared.NewAppEngineDatastore(ctx)
	q := store.TestRunQuery()
	shas, _, _ := q.GetAlignedRunSHAs(shared.GetDefaultProducts(), nil, nil, nil, nil, nil)
	assert.Equal(t, 0, len(shas))

	// Only 3 browsers.
	run := shared.TestRun{
		ProductAtRevision: shared.ProductAtRevision{
			Revision: "abcdef0000",
		},
		Labels:    []string{"foo"},
		TimeStart: time.Now().AddDate(0, 0, -1),
	}
	for _, browser := range browserNames[:len(browserNames)-1] {
		run.BrowserName = browser
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _, _ = q.GetAlignedRunSHAs(shared.GetDefaultProducts(), nil, nil, nil, nil, nil)
	assert.Len(t, shas, 0)

	// But, if request by any subset of those 3 browsers, we find the SHA.
	var products shared.ProductSpecs
	for _, browser := range browserNames[:len(browserNames)-1] {
		product := shared.ProductSpec{}
		product.BrowserName = browser
		products = append(products, product)
		shas, _, _ = q.GetAlignedRunSHAs(products, nil, nil, nil, nil, nil)
		assert.Len(t, shas, 1)
	}
	// And labels
	shas, _, _ = q.GetAlignedRunSHAs(products, mapset.NewSetWith("foo"), nil, nil, nil, nil)
	assert.Len(t, shas, 1)
	shas, _, _ = q.GetAlignedRunSHAs(products, mapset.NewSetWith("bar"), nil, nil, nil, nil)
	assert.Len(t, shas, 0)

	// All 4 browsers, but experimental.
	run.Revision = "abcdef0111"
	run.TimeStart = time.Now().AddDate(0, 0, -2)
	for _, browser := range browserNames {
		run.BrowserName = browser + "-" + shared.ExperimentalLabel
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _, _ = q.GetAlignedRunSHAs(shared.GetDefaultProducts(), nil, nil, nil, nil, nil)
	assert.Equal(t, 0, len(shas))

	// 2 browsers, and other 2, but experimental.
	run.Revision = "abcdef0222"
	run.TimeStart = time.Now().AddDate(0, 0, -3)
	for i, browser := range browserNames {
		run.BrowserName = browser
		if i > 1 {
			run.BrowserName += "-" + shared.ExperimentalLabel
		}
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _, _ = q.GetAlignedRunSHAs(shared.GetDefaultProducts(), nil, nil, nil, nil, nil)
	assert.Equal(t, 0, len(shas))

	// 2 browsers which are twice.
	run.Revision = "abcdef0333"
	run.TimeStart = time.Now().AddDate(0, 0, -3)
	for _, browser := range browserNames[:2] {
		run.BrowserName = browser
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _, _ = q.GetAlignedRunSHAs(shared.GetDefaultProducts(), nil, nil, nil, nil, nil)
	assert.Equal(t, 0, len(shas))

	// All 4 browsers.
	run.Revision = "abcdef0123"
	run.TimeStart = time.Now().AddDate(0, 0, -4)
	for _, browser := range browserNames {
		run.BrowserName = browser
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _, _ = q.GetAlignedRunSHAs(shared.GetDefaultProducts(), nil, nil, nil, nil, nil)
	assert.Equal(t, []string{"abcdef0123"}, shas)

	// Another (earlier) run, also all 4 browsers.
	run.Revision = "abcdef9999"
	run.TimeStart = time.Now().AddDate(0, 0, -5)
	for _, browser := range browserNames {
		run.BrowserName = browser
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _, _ = q.GetAlignedRunSHAs(shared.GetDefaultProducts(), nil, nil, nil, nil, nil)
	assert.Equal(t, []string{"abcdef0123", "abcdef9999"}, shas)
	// Limit 1
	one := 1
	shas, _, _ = q.GetAlignedRunSHAs(shared.GetDefaultProducts(), nil, nil, nil, &one, nil)
	assert.Equal(t, []string{"abcdef0123"}, shas)
	// Limit 1, Offset 1
	shas, _, _ = q.GetAlignedRunSHAs(shared.GetDefaultProducts(), nil, nil, nil, &one, &one)
	assert.Equal(t, []string{"abcdef9999"}, shas)
	// From 4 days ago @ midnight.
	from := time.Now().AddDate(0, 0, -4).Truncate(24 * time.Hour)
	shas, _, _ = q.GetAlignedRunSHAs(shared.GetDefaultProducts(), nil, &from, nil, nil, nil)
	assert.Equal(t, []string{"abcdef0123"}, shas)
}

func TestIsFeatureEnabled(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	flagName := "foo"
	ctx := shared.NewAppEngineContext(r)
	key := datastore.NewKey(ctx, "Flag", flagName, 0, nil)

	// No flag value.
	assert.False(t, shared.IsFeatureEnabled(ctx, flagName))
	// Disabled flag.
	datastore.Put(ctx, key, &shared.Flag{Enabled: false})
	assert.False(t, shared.IsFeatureEnabled(ctx, flagName))
	// Enabled flag.
	datastore.Put(ctx, key, &shared.Flag{Enabled: true})
	assert.True(t, shared.IsFeatureEnabled(ctx, flagName))
}

func TestGetSecret(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/", nil)
	assert.Nil(t, err)
	tokenName := "foo"
	ctx := shared.NewAppEngineContext(r)
	key := datastore.NewKey(ctx, "Token", tokenName, 0, nil)
	secret, err := shared.GetSecret(ctx, tokenName)
	assert.NotNil(t, err)
	assert.Equal(t, "", secret)
	// Token.
	datastore.Put(ctx, key, &shared.Token{Secret: "bar"})
	secret, err = shared.GetSecret(ctx, tokenName)
	assert.Nil(t, err)
	assert.Equal(t, "bar", secret)
}
