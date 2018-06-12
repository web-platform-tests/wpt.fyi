// +build medium

package shared

import (
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
	assert.Equal(t, "experimental", loaded[0].Labels[0])
	assert.Equal(t, "chrome-experimental", loaded[1].BrowserName)
}
