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
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine/datastore"
)

func TestGetTestRuns_VersionPrefix(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/run?product=chrome-66.0", nil)
	assert.Nil(t, err)

	// 'Yesterday', v66...139 earlier version.
	ctx := shared.NewAppEngineContext(r)
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
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &chrome)

	// 'Today', v66...181 (revision increased)
	chrome.BrowserVersion = "66.0.3359.181 beta"
	chrome.CreatedAt = now
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &chrome)

	// Also 'today', a v68 run.
	chrome.BrowserVersion = "68.0.3432.3"
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &chrome)

	r, _ = i.NewRequest("GET", "/api/run?product=chrome-6", nil)
	resp := httptest.NewRecorder()
	apiTestRunHandler(resp, r)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	r, _ = i.NewRequest("GET", "/api/run?product=chrome-66.0", nil)
	resp = httptest.NewRecorder()
	apiTestRunHandler(resp, r)
	body, _ := ioutil.ReadAll(resp.Result().Body)
	assert.Equalf(t, http.StatusOK, resp.Code, string(body))
	var result66 shared.TestRun
	json.Unmarshal(body, &result66)
	assert.Equal(t, "66.0.3359.181 beta", result66.BrowserVersion)

	r, _ = i.NewRequest("GET", "/api/run?product=chrome-66.0.3359.139", nil)
	resp = httptest.NewRecorder()
	apiTestRunHandler(resp, r)
	body, _ = ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	var result66139 shared.TestRun
	json.Unmarshal(body, &result66139)
	assert.Equal(t, "66.0.3359.139", result66139.BrowserVersion)

	r, _ = i.NewRequest("GET", "/api/run?product=chrome-66.0.3359.181 beta", nil)
	resp = httptest.NewRecorder()
	apiTestRunHandler(resp, r)
	body, _ = ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	json.Unmarshal(body, &result66139)
	assert.Equal(t, "66.0.3359.181 beta", result66139.BrowserVersion)

	r, _ = i.NewRequest("GET", "/api/run?product=chrome-68", nil)
	resp = httptest.NewRecorder()
	apiTestRunHandler(resp, r)
	body, _ = ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	var result68 shared.TestRun
	json.Unmarshal(body, &result68)
	assert.Equal(t, "68.0.3432.3", result68.BrowserVersion)
}

func TestGetTestRuns_SHA(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/runs", nil)
	assert.Nil(t, err)

	ctx := shared.NewAppEngineContext(r)
	now := time.Now()
	run := shared.TestRun{}
	run.BrowserVersion = "66.0.3359.139"
	run.Revision = "abcdef0123"
	run.CreatedAt = now.AddDate(0, 0, -1)

	run.BrowserName = "chrome"
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	run.BrowserName = "safari"
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)

	run.Revision = "9876543210"
	run.BrowserName = "firefox"
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)

	run.Revision = "9999999999"
	run.BrowserName = "edge"
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)

	r, _ = i.NewRequest("GET", "/api/runs?sha=abcdef0123", nil)
	resp := httptest.NewRecorder()
	apiTestRunsHandler(resp, r)
	body, _ := ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	var results shared.TestRuns
	json.Unmarshal(body, &results)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "abcdef0123", results[0].Revision)

	// ?aligned ignored if SHA provided.
	r, _ = i.NewRequest("GET", "/api/runs?sha=abcdef0123&aligned", nil)
	resp = httptest.NewRecorder()
	apiTestRunsHandler(resp, r)
	body, _ = ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	json.Unmarshal(body, &results)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "abcdef0123", results[0].Revision)

	// ?aligned - no aligned runs.
	r, _ = i.NewRequest("GET", "/api/runs?aligned", nil)
	resp = httptest.NewRecorder()
	apiTestRunsHandler(resp, r)
	body, _ = ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	run.Revision = "1111111111"
	for _, name := range shared.GetDefaultBrowserNames() {
		run.BrowserName = name
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	r, _ = i.NewRequest("GET", "/api/runs?aligned", nil)
	resp = httptest.NewRecorder()
	apiTestRunsHandler(resp, r)
	body, _ = ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	json.Unmarshal(body, &results)
	assert.Equal(t, 4, len(results))
	assert.Equal(t, "1111111111", results[0].Revision)
}
