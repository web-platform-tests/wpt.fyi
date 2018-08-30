// +build medium

package shared

import (
	"strconv"
	"strings"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine/datastore"
)

func TestLoadTestRuns(t *testing.T) {
	testRun := TestRun{
		ProductAtRevision: ProductAtRevision{
			Product: Product{
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

	chrome, _ := ParseProductSpec("chrome")
	loaded, err := LoadTestRuns(ctx, []ProductSpec{chrome}, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(loaded))
	assert.Equalf(t, key.IntID(), loaded[0].ID, "ID field should be populated.")
}

func TestLoadTestRuns_Experimental_Only(t *testing.T) {
	testRuns := TestRuns{
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product: Product{
					BrowserName:    "chrome",
					BrowserVersion: "63.0",
					OSName:         "linux",
				},
				Revision: "1234567890",
			},
			ResultsURL: "/static/chrome-63.0-linux-summary.json.gz",
			CreatedAt:  time.Now(),
		},
		TestRun{
			// We no longer have runs like this (with the "-experimental" suffix but without
			// the "experimental" label; and it should no longer be considered experimental.
			ProductAtRevision: ProductAtRevision{
				Product: Product{
					BrowserName:    "chrome-experimental",
					BrowserVersion: "63.0",
					OSName:         "linux",
				},
				Revision: "1234567890",
			},
			ResultsURL: "/static/chrome-experimental-63.0-linux-summary.json.gz",
			CreatedAt:  time.Now(),
		},
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product: Product{
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
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product: Product{
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

	chrome, chromeExperimental := ProductSpec{}, ProductSpec{}
	chrome.BrowserName = "chrome"
	chromeExperimental.BrowserName = "chrome-experimental"
	products := ProductSpecs{chrome, chromeExperimental}
	labels := mapset.NewSet()
	labels.Add("experimental")
	ten := 10
	loaded, err := LoadTestRuns(ctx, products, labels, nil, nil, nil, &ten)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(loaded))
	assert.Equal(t, "64.0", loaded[0].BrowserVersion)
	assert.Equal(t, "65.0", loaded[1].BrowserVersion)
}

func TestLoadTestRuns_LabelinProductSpec(t *testing.T) {
	testRuns := []TestRun{
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product: Product{BrowserName: "chrome"},
			},
			Labels: []string{"foo"},
		},
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product: Product{BrowserName: "chrome"},
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

	products := make([]ProductSpec, 1)
	products[0].BrowserName = "chrome"
	products[0].Labels = mapset.NewSetWith("foo")
	loaded, err := LoadTestRuns(ctx, products, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(loaded))
	assert.Equal(t, "foo", loaded[0].Labels[0])
}

func TestLoadTestRuns_SHAinProductSpec(t *testing.T) {
	testRuns := []TestRun{
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product:  Product{BrowserName: "chrome"},
				Revision: "0000000000",
			},
		},
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product:  Product{BrowserName: "chrome"},
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

	products := make([]ProductSpec, 1)
	products[0].BrowserName = "chrome"
	products[0].Revision = "1111111111"
	loaded, err := LoadTestRuns(ctx, products, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(loaded))
	assert.Equal(t, "1111111111", loaded[0].Revision)
}

func TestLoadTestRuns_MultipleSHAs(t *testing.T) {
	var testRuns TestRuns
	for i := 0; i < 3; i++ {
		testRun := TestRun{}
		testRun.BrowserName = "chrome"
		testRun.Revision = strings.Repeat(strconv.FormatInt(int64(i), 10), 10)
		testRuns = append(testRuns, testRun)
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

	products := make([]ProductSpec, 1)
	products[0].BrowserName = "chrome"
	shas := []string{"1111111111", "2222222222"}
	loaded, err := LoadTestRuns(ctx, products, nil, shas, nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(loaded))
	assert.Equal(t, shas[0], loaded[0].Revision)
	assert.Equal(t, shas[1], loaded[1].Revision)
}

func TestLoadTestRuns_Ordering(t *testing.T) {
	testRuns := []TestRun{
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product: Product{
					BrowserName: "chrome",
				},
				Revision: "1234567890",
			},
			CreatedAt: time.Now(),
			TimeStart: time.Now().AddDate(0, 0, -1),
		},
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product: Product{
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

	chrome, _ := ParseProductSpec("chrome")
	loaded, err := LoadTestRuns(ctx, []ProductSpec{chrome}, nil, nil, nil, nil, nil)
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
	testRuns := []TestRun{
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product: Product{
					BrowserName: "chrome",
				},
				Revision: "1234567890",
			},
			TimeStart: now,
		},
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product: Product{
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

	chrome, _ := ParseProductSpec("chrome")
	loaded, err := LoadTestRuns(ctx, []ProductSpec{chrome}, nil, nil, &yesterday, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(loaded))
	assert.Equal(t, "1234567890", loaded[0].Revision)
}

func TestLoadTestRuns_To(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	testRuns := []TestRun{
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product: Product{
					BrowserName: "chrome",
				},
				Revision: "1234567890",
			},
			TimeStart: now,
		},
		TestRun{
			ProductAtRevision: ProductAtRevision{
				Product: Product{
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

	chrome, _ := ParseProductSpec("chrome")
	loaded, err := LoadTestRuns(ctx, []ProductSpec{chrome}, nil, nil, nil, &now, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(loaded))
	assert.Equal(t, "0987654321", loaded[0].Revision)
}

func TestGetCompleteRunSHAs(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	browserNames := GetDefaultBrowserNames()

	// Nothing in datastore.
	shas, _ := GetCompleteRunSHAs(ctx, nil, nil, nil)
	assert.Equal(t, 0, len(shas))

	// Only 3 browsers.
	run := TestRun{
		ProductAtRevision: ProductAtRevision{
			Revision: "abcdef0000",
		},
		TimeStart: time.Now().AddDate(0, 0, -1),
	}
	for _, browser := range browserNames[:len(browserNames)-1] {
		run.BrowserName = browser
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _ = GetCompleteRunSHAs(ctx, nil, nil, nil)
	assert.Equal(t, 0, len(shas))

	// All 4 browsers, but experimental.
	run.Revision = "abcdef0111"
	run.TimeStart = time.Now().AddDate(0, 0, -2)
	for _, browser := range browserNames {
		run.BrowserName = browser + "-" + ExperimentalLabel
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _ = GetCompleteRunSHAs(ctx, nil, nil, nil)
	assert.Equal(t, 0, len(shas))

	// 2 browsers, and other 2, but experimental.
	run.Revision = "abcdef0222"
	run.TimeStart = time.Now().AddDate(0, 0, -3)
	for i, browser := range browserNames {
		run.BrowserName = browser
		if i > 1 {
			run.BrowserName += "-" + ExperimentalLabel
		}
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _ = GetCompleteRunSHAs(ctx, nil, nil, nil)
	assert.Equal(t, 0, len(shas))

	// All 4 browsers.
	run.Revision = "abcdef0123"
	run.TimeStart = time.Now().AddDate(0, 0, -4)
	for _, browser := range browserNames {
		run.BrowserName = browser
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _ = GetCompleteRunSHAs(ctx, nil, nil, nil)
	assert.Equal(t, []string{"abcdef0123"}, shas)

	// Another (earlier) run, also all 4 browsers.
	run.Revision = "abcdef9999"
	run.TimeStart = time.Now().AddDate(0, 0, -5)
	for _, browser := range browserNames {
		run.BrowserName = browser
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	shas, _ = GetCompleteRunSHAs(ctx, nil, nil, nil)
	assert.Equal(t, []string{"abcdef0123", "abcdef9999"}, shas)
	// Limit 1
	one := 1
	shas, _ = GetCompleteRunSHAs(ctx, nil, nil, &one)
	assert.Equal(t, []string{"abcdef0123"}, shas)
	// From 4 days ago @ midnight.
	from := time.Now().AddDate(0, 0, -4).Truncate(24 * time.Hour)
	shas, _ = GetCompleteRunSHAs(ctx, &from, nil, nil)
	assert.Equal(t, []string{"abcdef0123"}, shas)
}
