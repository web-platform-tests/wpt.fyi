// +build medium

package shared_test

import (
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
	loaded, err := shared.LoadTestRuns(ctx, []shared.ProductSpec{chrome}, nil, shared.LatestSHA, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(loaded))
	assert.Equalf(t, key.IntID(), loaded[0].ID, "ID field should be populated.")
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
	loaded, err := shared.LoadTestRuns(ctx, products, labels, shared.LatestSHA, nil, nil, &ten)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(loaded))
	assert.Equal(t, "64.0", loaded[0].BrowserVersion)
	assert.Equal(t, "65.0", loaded[1].BrowserVersion)
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
	loaded, err := shared.LoadTestRuns(ctx, products, nil, shared.LatestSHA, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(loaded))
	assert.Equal(t, "foo", loaded[0].Labels[0])
}

func TestLoadTestRuns_SHAinProductSpec(t *testing.T) {
	testRuns := []shared.TestRun{
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product:  shared.Product{BrowserName: "chrome"},
				Revision: "0000000000",
			},
		},
		shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product:  shared.Product{BrowserName: "chrome"},
				Revision: "1111111111",
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
	products[0].Revision = "1111111111"
	loaded, err := shared.LoadTestRuns(ctx, products, nil, shared.LatestSHA, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(loaded))
	assert.Equal(t, "1111111111", loaded[0].Revision)
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
	loaded, err := shared.LoadTestRuns(ctx, []shared.ProductSpec{chrome}, nil, shared.LatestSHA, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(loaded))
	// Runs should be ordered descendingly by TimeStart.
	assert.Equal(t, "0987654321", loaded[0].Revision)
	assert.Equal(t, "1234567890", loaded[1].Revision)
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
	loaded, err := shared.LoadTestRuns(ctx, []shared.ProductSpec{chrome}, nil, shared.LatestSHA, &yesterday, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(loaded))
	assert.Equal(t, "1234567890", loaded[0].Revision)
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
	loaded, err := shared.LoadTestRuns(ctx, shared.ProductSpecs{chrome}, nil, shared.LatestSHA, nil, &now, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(loaded))
	assert.Equal(t, "0987654321", loaded[0].Revision)
}

func TestGetAlignedRunSHAs(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	browserNames := shared.GetDefaultBrowserNames()

	// Nothing in datastore.
	shas, _, _ := shared.GetAlignedRunSHAs(ctx, shared.GetDefaultProducts(), nil, nil, nil, nil)
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
	shas, _, _ = shared.GetAlignedRunSHAs(ctx, shared.GetDefaultProducts(), nil, nil, nil, nil)
	assert.Len(t, shas, 0)

	// But, if request by any subset of those 3 browsers, we find the SHA.
	var products shared.ProductSpecs
	for _, browser := range browserNames[:len(browserNames)-1] {
		product := shared.ProductSpec{}
		product.BrowserName = browser
		products = append(products, product)
		shas, _, _ = shared.GetAlignedRunSHAs(ctx, products, nil, nil, nil, nil)
		assert.Len(t, shas, 1)
	}
	// And labels
	shas, _, _ = shared.GetAlignedRunSHAs(ctx, products, mapset.NewSetWith("foo"), nil, nil, nil)
	assert.Len(t, shas, 1)
	shas, _, _ = shared.GetAlignedRunSHAs(ctx, products, mapset.NewSetWith("bar"), nil, nil, nil)
	assert.Len(t, shas, 0)

	// All 4 browsers, but experimental.
	run.Revision = "abcdef0111"
	run.TimeStart = time.Now().AddDate(0, 0, -2)
	for _, browser := range browserNames {
		run.BrowserName = browser + "-" + shared.ExperimentalLabel
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _, _ = shared.GetAlignedRunSHAs(ctx, shared.GetDefaultProducts(), nil, nil, nil, nil)
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
	shas, _, _ = shared.GetAlignedRunSHAs(ctx, shared.GetDefaultProducts(), nil, nil, nil, nil)
	assert.Equal(t, 0, len(shas))

	// 2 browsers which are twice.
	run.Revision = "abcdef0333"
	run.TimeStart = time.Now().AddDate(0, 0, -3)
	for _, browser := range browserNames[:2] {
		run.BrowserName = browser
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _, _ = shared.GetAlignedRunSHAs(ctx, shared.GetDefaultProducts(), nil, nil, nil, nil)
	assert.Equal(t, 0, len(shas))

	// All 4 browsers.
	run.Revision = "abcdef0123"
	run.TimeStart = time.Now().AddDate(0, 0, -4)
	for _, browser := range browserNames {
		run.BrowserName = browser
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _, _ = shared.GetAlignedRunSHAs(ctx, shared.GetDefaultProducts(), nil, nil, nil, nil)
	assert.Equal(t, []string{"abcdef0123"}, shas)

	// Another (earlier) run, also all 4 browsers.
	run.Revision = "abcdef9999"
	run.TimeStart = time.Now().AddDate(0, 0, -5)
	for _, browser := range browserNames {
		run.BrowserName = browser
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _, _ = shared.GetAlignedRunSHAs(ctx, shared.GetDefaultProducts(), nil, nil, nil, nil)
	assert.Equal(t, []string{"abcdef0123", "abcdef9999"}, shas)
	// Limit 1
	one := 1
	shas, _, _ = shared.GetAlignedRunSHAs(ctx, shared.GetDefaultProducts(), nil, nil, nil, &one)
	assert.Equal(t, []string{"abcdef0123"}, shas)
	// From 4 days ago @ midnight.
	from := time.Now().AddDate(0, 0, -4).Truncate(24 * time.Hour)
	shas, _, _ = shared.GetAlignedRunSHAs(ctx, shared.GetDefaultProducts(), nil, &from, nil, nil)
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
