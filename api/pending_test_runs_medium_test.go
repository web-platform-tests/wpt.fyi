// +build medium

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
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

	ctx := shared.NewAppEngineContext(r)

	now := time.Now().Truncate(time.Minute).In(time.UTC)
	yesterday := now.Add(time.Hour * -24)

	invalid := shared.PendingTestRun{}
	invalid.Created = yesterday
	invalid.Updated = now
	invalid.Stage = shared.StageInvalid
	key := datastore.NewIncompleteKey(ctx, "PendingTestRun", nil)
	key, err = datastore.Put(ctx, key, &invalid)
	assert.Nil(t, err)
	invalid.ID = key.IntID()

	running := shared.PendingTestRun{}
	running.Created = yesterday.Add(time.Hour)
	running.Updated = now.Add(time.Minute * -5)
	running.Stage = shared.StageCIRunning
	key = datastore.NewIncompleteKey(ctx, "PendingTestRun", nil)
	key, err = datastore.Put(ctx, key, &running)
	assert.Nil(t, err)
	running.ID = key.IntID()

	t.Run("/api/status", func(t *testing.T) {
		r, _ = i.NewRequest("GET", "/api/status", nil)
		resp := httptest.NewRecorder()
		apiPendingTestRunsHandler(resp, r)
		body, _ := ioutil.ReadAll(resp.Result().Body)
		assert.Equal(t, http.StatusOK, resp.Code, string(body))
		var results []shared.PendingTestRun
		json.Unmarshal(body, &results)
		assert.Len(t, results, 2)
		assert.Equal(t, results[0].ID, invalid.ID)
		assert.Equal(t, results[1].ID, running.ID)
	})

	t.Run("/api/status/pending", func(t *testing.T) {
		r, _ = i.NewRequest("GET", "/api/status/pending", nil)
		r = mux.SetURLVars(r, map[string]string{"filter": "pending"})
		resp := httptest.NewRecorder()
		apiPendingTestRunsHandler(resp, r)
		body, _ := ioutil.ReadAll(resp.Result().Body)
		assert.Equal(t, http.StatusOK, resp.Code, string(body))
		var results []shared.PendingTestRun
		json.Unmarshal(body, &results)
		assert.Len(t, results, 1)
		assert.Equal(t, results[0].ID, running.ID)
	})
}
