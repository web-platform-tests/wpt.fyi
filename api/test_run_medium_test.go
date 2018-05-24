// +build medium

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestGetTestRuns_VersionPrefix(t *testing.T) {
	i, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	assert.Nil(t, err)
	r, err := i.NewRequest("GET", "/api/run?product=chrome-66.0", nil)
	assert.Nil(t, err)

	// Yesterday, earlier version.
	ctx := appengine.NewContext(r)
	now := time.Now()
	chrome := shared.TestRun{
		ProductAtRevision: shared.ProductAtRevision{
			Product: shared.Product{
				BrowserName:    "chrome",
				BrowserVersion: "66.0.3359.139",
			},
			Revision: "abcdef0123",
		},
		CreatedAt: now.AddDate(0, 0, -1),
	}
	key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
	datastore.Put(ctx, key, &chrome)

	// Today, revision increased and an experimental run
	chrome.BrowserVersion = "66.0.3359.181"
	chrome.CreatedAt = now
	datastore.Put(ctx, key, &chrome)

	// Also today, a v68 run.
	chrome.BrowserVersion = "68.0.3432.3"
	datastore.Put(ctx, key, &chrome)

	r, err = i.NewRequest("GET", "/api/run?product=chrome-66.0", nil)
	resp := httptest.NewRecorder()
	apiTestRunGetHandler(resp, r)
	body, err := ioutil.ReadAll(resp.Result().Body)
	assert.Nil(t, err)
	assert.Equal(t, resp.Code, http.StatusOK)
	var result shared.TestRun
	json.Unmarshal(body, &result)
	assert.Equal(t, "66.0.3359.181", result.BrowserVersion)

	r, err = i.NewRequest("GET", "/api/run?product=chrome-66.0.3359.139", nil)
	assert.Nil(t, err)
	resp = httptest.NewRecorder()
	body, err = ioutil.ReadAll(resp.Result().Body)
	assert.Nil(t, err)
	assert.Equal(t, resp.Code, http.StatusOK)
	json.Unmarshal(body, &result)
	assert.Equal(t, "66.0.3359.139", result.BrowserVersion)

	r, err = i.NewRequest("GET", "/api/run?product=chrome-68", nil)
	assert.Nil(t, err)
	resp = httptest.NewRecorder()
	body, err = ioutil.ReadAll(resp.Result().Body)
	assert.Nil(t, err)
	assert.Equal(t, resp.Code, http.StatusOK)
	json.Unmarshal(body, &result)
	assert.Equal(t, "68.0.3432.3", result.BrowserVersion)
}
