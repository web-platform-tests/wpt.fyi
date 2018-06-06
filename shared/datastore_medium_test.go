// +build medium

package shared

import (
	"testing"
	"time"

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
				OSVersion:      "3.16",
			},
			Revision: "1234567890",
		},
		ResultsURL: "/static/chrome-63.0-linux-summary.json.gz",
		CreatedAt:  time.Now(),
	}

	i, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/run?product=chrome-66.0", nil)
	assert.Nil(t, err)

	// 'Yesterday', v66...139 earlier version.
	ctx := appengine.NewContext(r)
	key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
	key, _ = datastore.Put(ctx, key, &testRun)

	chrome, _ := ParseProduct("chrome")
	loaded, err := LoadTestRuns(ctx, []Product{chrome}, nil, "latest", nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(loaded))
	assert.Equalf(t, key.IntID(), loaded[0].ID, "ID field should be populated.")
}
