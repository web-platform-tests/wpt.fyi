// +build medium

package shared

import (
	"strconv"
	"strings"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
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

	i, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	assert.Nil(t, err)
	defer i.Close()
	// URL is a placeholder and is not used in this test.
	r, err := i.NewRequest("GET", "/api/run", nil)
	assert.Nil(t, err)

	ctx := appengine.NewContext(r)
	key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
	key, _ = datastore.Put(ctx, key, &testRun)

	chrome, _ := ParseProduct("chrome")
	loaded, err := LoadTestRuns(ctx, []Product{chrome}, nil, nil, nil, nil)
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

	i, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	assert.Nil(t, err)
	defer i.Close()
	// URL is a placeholder and is not used in this test.
	r, err := i.NewRequest("GET", "/api/run", nil)
	assert.Nil(t, err)

	ctx := appengine.NewContext(r)
	keys := make([]*datastore.Key, len(testRuns))
	for i := range testRuns {
		keys[i] = datastore.NewIncompleteKey(ctx, "TestRun", nil)
	}
	keys, err = datastore.PutMulti(ctx, keys, testRuns)
	assert.Nil(t, err)

	products := []Product{Product{BrowserName: "chrome"}, Product{BrowserName: "chrome-experimental"}}
	labels := mapset.NewSet()
	labels.Add("experimental")
	ten := 10
	loaded, err := LoadTestRuns(ctx, products, labels, nil, nil, &ten)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(loaded))
	assert.Equal(t, "64.0", loaded[0].BrowserVersion)
	assert.Equal(t, "65.0", loaded[1].BrowserVersion)
}

func TestLoadTestRuns_MultipleSHAs(t *testing.T) {
	var testRuns TestRuns
	for i := 0; i < 3; i++ {
		testRun := TestRun{}
		testRun.BrowserName = "chrome"
		testRun.Revision = strings.Repeat(strconv.FormatInt(int64(i), 10), 10)
		testRuns = append(testRuns, testRun)
	}

	i, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	assert.Nil(t, err)
	defer i.Close()
	// URL is a placeholder and is not used in this test.
	r, err := i.NewRequest("GET", "/api/runs", nil)
	assert.Nil(t, err)

	ctx := appengine.NewContext(r)
	keys := make([]*datastore.Key, len(testRuns))
	for i := range testRuns {
		keys[i] = datastore.NewIncompleteKey(ctx, "TestRun", nil)
	}
	keys, err = datastore.PutMulti(ctx, keys, testRuns)
	assert.Nil(t, err)

	products := []Product{Product{BrowserName: "chrome"}}
	shas := []string{"1111111111", "2222222222"}
	loaded, err := LoadTestRuns(ctx, products, nil, shas, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(loaded))
	assert.Equal(t, shas[0], loaded[0].Revision)
	assert.Equal(t, shas[1], loaded[1].Revision)
}
