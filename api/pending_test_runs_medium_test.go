// +build medium

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine/datastore"
)

func TestAPIPendingTestHandler(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/status", nil)
	assert.Nil(t, err)

	testRun := shared.PendingTestRun{
		Stage: shared.StageWptFyiProcessing,
	}

	ctx := shared.NewAppEngineContext(r)
	key := datastore.NewIncompleteKey(ctx, "PendingTestRun", nil)
	key, err = datastore.Put(ctx, key, &testRun)
	assert.Nil(t, err)

	r, _ = i.NewRequest("GET", "/api/status", nil)
	resp := httptest.NewRecorder()
	apiPendingTestRunsHandler(resp, r)
	body, _ := ioutil.ReadAll(resp.Result().Body)
	assert.Equal(t, http.StatusOK, resp.Code, string(body))
	var results []shared.PendingTestRun
	json.Unmarshal(body, &results)
	assert.Len(t, results, 1)
	assert.Equal(t, results[0].ID, key.IntID())
	assert.Equal(t, results[0].Stage, shared.StageWptFyiProcessing)
}
