// +build medium

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func TestGetTestRunByID(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/runs/123", nil)
	r = mux.SetURLVars(r, map[string]string{"id": "123"})
	assert.Nil(t, err)

	ctx := appengine.NewContext(r)
	resp := httptest.NewRecorder()
	apiTestRunHandler(resp, r)
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
	apiTestRunHandler(resp, r)
	assert.Equal(t, http.StatusOK, resp.Code)
	var bodyTestRun shared.TestRun
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &bodyTestRun)
	assert.Equal(t, int64(123), bodyTestRun.ID)
}
