//go:build medium

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestGetTestRunByID(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/runs/123", nil)
	r = mux.SetURLVars(r, map[string]string{"id": "123"})
	assert.Nil(t, err)

	ctx := r.Context()
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

	store := shared.NewAppEngineDatastore(ctx, false)
	store.Put(store.NewIDKey("TestRun", 123), &chrome)
	resp = httptest.NewRecorder()
	apiTestRunHandler(resp, r)
	assert.Equal(t, http.StatusOK, resp.Code)
	var bodyTestRun shared.TestRun
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &bodyTestRun)
	assert.Equal(t, int64(123), bodyTestRun.ID)
}
