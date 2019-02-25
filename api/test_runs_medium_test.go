// +build medium

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

func TestGetTestRuns_RunIDs(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/runs?run_id=123", nil)
	assert.Nil(t, err)

	ctx := shared.NewAppEngineContext(r)
	now := time.Now()
	run := shared.TestRun{}
	run.BrowserVersion = "66.0.3359.139"
	run.FullRevisionHash = strings.Repeat("abcdef0123", 4)
	run.Revision = run.FullRevisionHash[:10]
	run.CreatedAt = now.AddDate(0, 0, -1)
	keys := make([]*datastore.Key, 2)

	run.BrowserName = "chrome"
	keys[0], _ = datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	run.BrowserName = "safari"
	keys[1], _ = datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)

	// run_id=123 from above should 404.
	resp := httptest.NewRecorder()
	apiTestRunsHandler(resp, r)
	body, _ := ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	r, _ = i.NewRequest("GET", fmt.Sprintf("/api/runs?run_id=%v", keys[0].IntID()), nil)
	resp = httptest.NewRecorder()
	apiTestRunsHandler(resp, r)
	body, _ = ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	var results shared.TestRuns
	json.Unmarshal(body, &results)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "chrome", results[0].BrowserName)

	r, _ = i.NewRequest("GET", fmt.Sprintf("/api/runs?run_ids=%v,%v", keys[1].IntID(), keys[0].IntID()), nil)
	resp = httptest.NewRecorder()
	apiTestRunsHandler(resp, r)
	body, _ = ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	json.Unmarshal(body, &results)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "safari", results[0].BrowserName)
	assert.Equal(t, "chrome", results[1].BrowserName)
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
	run.FullRevisionHash = strings.Repeat("abcdef0123", 4)
	run.Revision = run.FullRevisionHash[:10]
	run.CreatedAt = now.AddDate(0, 0, -1)

	run.BrowserName = "chrome"
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	run.BrowserName = "safari"
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)

	run.FullRevisionHash = strings.Repeat("9876543210", 4)
	run.Revision = run.FullRevisionHash[:10]
	run.BrowserName = "firefox"
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)

	run.FullRevisionHash = strings.Repeat("9999999999", 4)
	run.Revision = run.FullRevisionHash[:10]
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

	run.FullRevisionHash = strings.Repeat("1111111111", 4)
	run.Revision = run.FullRevisionHash[:10]
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

func TestGetTestRuns_Pagination(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/runs", nil)
	assert.Nil(t, err)

	ctx := shared.NewAppEngineContext(r)
	ds := shared.NewAppEngineDatastore(ctx, false)
	now := time.Now()
	run := shared.TestRun{}
	run.BrowserName = "chrome"
	for _, d := range []int{-3, -2, -1} {
		run.CreatedAt = now.AddDate(0, 0, d)
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}

	r, _ = i.NewRequest("GET", "/api/runs?product=chrome&max-count=2", nil)
	resp := httptest.NewRecorder()

	// Feature disabled
	apiTestRunsHandler(resp, r)
	next := resp.Header().Get(nextPageTokenHeaderName)
	assert.Equal(t, "", next)

	// Feature enabled
	shared.SetFeature(ds, shared.Flag{Name: paginationTokenFeatureFlagName, Enabled: true})
	resp = httptest.NewRecorder()
	apiTestRunsHandler(resp, r)
	next = resp.Header().Get(nextPageTokenHeaderName)
	assert.NotEqual(t, "", next)

	body, _ := ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	var pageOne shared.TestRuns
	json.Unmarshal(body, &pageOne)
	assert.Equal(t, 2, len(pageOne))

	r, _ = i.NewRequest("GET", fmt.Sprintf("/api/runs?page=%s", url.QueryEscape(next)), nil)
	resp = httptest.NewRecorder()
	apiTestRunsHandler(resp, r)
	body, _ = ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	var pageTwo shared.TestRuns
	json.Unmarshal(body, &pageTwo)
	assert.Equal(t, 1, len(pageTwo))
	next = resp.Header().Get(nextPageTokenHeaderName)
	assert.Equal(t, "", next)
}
