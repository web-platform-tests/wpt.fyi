// +build medium

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestGetTestRunByID(t *testing.T) {
	i, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/runs/123", nil)
	r = mux.SetURLVars(r, map[string]string{"id": "123"})
	assert.Nil(t, err)

	ctx := appengine.NewContext(r)
	resp := httptest.NewRecorder()
	apiTestRunGetHandler(resp, r)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	chrome := shared.TestRun{
		ProductAtRevision: shared.ProductAtRevision{
			Product: shared.Product{
				BrowserName: "chrome",
			},
			Revision: "abcdef0123",
		},
	}

	datastore.Put(ctx, datastore.NewKey(ctx, "TestRun", "", 123, nil), &chrome)
	resp = httptest.NewRecorder()
	apiTestRunGetHandler(resp, r)
	assert.Equal(t, http.StatusOK, resp.Code)
	var bodyTestRun shared.TestRun
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &bodyTestRun)
	assert.Equal(t, int64(123), bodyTestRun.ID)
}

func TestGetTestRuns_VersionPrefix(t *testing.T) {
	i, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/run?product=chrome-66.0", nil)
	assert.Nil(t, err)

	// 'Yesterday', v66...139 earlier version.
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
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &chrome)

	// 'Today', v66...181 (revision increased)
	chrome.BrowserVersion = "66.0.3359.181"
	chrome.CreatedAt = now
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &chrome)

	// Also 'today', a v68 run.
	chrome.BrowserVersion = "68.0.3432.3"
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &chrome)

	r, _ = i.NewRequest("GET", "/api/run?product=chrome-6", nil)
	resp := httptest.NewRecorder()
	apiTestRunGetHandler(resp, r)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	r, _ = i.NewRequest("GET", "/api/run?product=chrome-66.0", nil)
	resp = httptest.NewRecorder()
	apiTestRunGetHandler(resp, r)
	body, _ := ioutil.ReadAll(resp.Result().Body)
	assert.Equalf(t, http.StatusOK, resp.Code, string(body))
	var result66 shared.TestRun
	json.Unmarshal(body, &result66)
	assert.Equal(t, "66.0.3359.181", result66.BrowserVersion)

	r, _ = i.NewRequest("GET", "/api/run?product=chrome-66.0.3359.139", nil)
	resp = httptest.NewRecorder()
	apiTestRunGetHandler(resp, r)
	body, _ = ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	var result66139 shared.TestRun
	json.Unmarshal(body, &result66139)
	assert.Equal(t, "66.0.3359.139", result66139.BrowserVersion)

	r, _ = i.NewRequest("GET", "/api/run?product=chrome-68", nil)
	resp = httptest.NewRecorder()
	apiTestRunGetHandler(resp, r)
	body, _ = ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	var result68 shared.TestRun
	json.Unmarshal(body, &result68)
	assert.Equal(t, "68.0.3432.3", result68.BrowserVersion)
}

func TestTestRunPostHandler(t *testing.T) {
	i, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	assert.Nil(t, err)
	defer i.Close()
	payload := map[string]string{
		"browser_name":    "firefox",
		"browser_version": "59.0",
		"os_name":         "linux",
		"os_version":      "4.4",
		"revision":        "0123456789",
		// Intentionally missing full_revision_hash; no error should be raised.
		// Unknown parameters should be ignored.
		"_random_extra_key_": "some_value",
	}
	body, err := json.Marshal(payload)
	assert.Nil(t, err)
	r, err := i.NewRequest("POST", "/api/runs?secret=secret-token", strings.NewReader(string(body)))
	assert.Nil(t, err)

	ctx := appengine.NewContext(r)
	token := &shared.Token{Secret: "secret-token"}
	datastore.Put(ctx, datastore.NewKey(ctx, "Token", "upload-token", 0, nil), token)
	resp := httptest.NewRecorder()

	TestRunPostHandler(resp, r)
	assert.Equal(t, http.StatusCreated, resp.Code)
}
