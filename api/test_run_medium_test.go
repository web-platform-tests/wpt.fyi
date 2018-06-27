// +build medium

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestTestRunPostHandler(t *testing.T) {
	i, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	assert.Nil(t, err)
	defer i.Close()
	payload := map[string]interface{}{
		"browser_name":    "firefox",
		"browser_version": "59.0",
		"os_name":         "linux",
		"os_version":      "4.4",
		"revision":        "0123456789",
		"labels":          []string{"foo", "bar"},
		"time_start":      "2018-06-21T18:39:54.218000+00:00",
		"time_end":        "2018-06-21T20:03:49Z",
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

	var testRuns []shared.TestRun
	datastore.NewQuery("TestRun").Limit(1).GetAll(ctx, &testRuns)
	assert.Equal(t, "firefox", testRuns[0].BrowserName)
	assert.Equal(t, []string{"foo", "bar"}, testRuns[0].Labels)
	assert.False(t, testRuns[0].TimeStart.IsZero())
	assert.False(t, testRuns[0].TimeEnd.IsZero())
}

func TestTestRunPostHandler_NoTimestamps(t *testing.T) {
	i, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	assert.Nil(t, err)
	defer i.Close()
	payload := map[string]interface{}{
		"browser_name":    "firefox",
		"browser_version": "59.0",
		"os_name":         "linux",
		"revision":        "0123456789",
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

	var testRuns []shared.TestRun
	datastore.NewQuery("TestRun").Limit(1).GetAll(ctx, &testRuns)
	assert.False(t, testRuns[0].CreatedAt.IsZero())
	assert.False(t, testRuns[0].TimeStart.IsZero())
	assert.Equal(t, testRuns[0].TimeStart, testRuns[0].TimeEnd)
}
