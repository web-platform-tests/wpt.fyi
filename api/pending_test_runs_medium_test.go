//go:build medium

package api //nolint:revive

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"

	"github.com/stretchr/testify/assert"
)

func createPendingRun(ctx context.Context, run *shared.PendingTestRun) error {
	store := shared.NewAppEngineDatastore(ctx, false)
	key := store.NewIncompleteKey("PendingTestRun")
	key, err := store.Put(key, run)
	run.ID = key.IntID()
	return err
}

func TestAPIPendingTestHandler(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/status", nil)
	assert.Nil(t, err)

	ctx := r.Context()

	now := time.Now().Truncate(time.Minute).In(time.UTC)
	yesterday := now.Add(time.Hour * -24)

	invalid := shared.PendingTestRun{
		Created: yesterday,
		Updated: now,
		Stage:   shared.StageInvalid,
	}
	assert.Nil(t, createPendingRun(ctx, &invalid))

	empty := shared.PendingTestRun{
		Created: yesterday,
		Updated: now.Add(time.Minute * -1),
		Stage:   shared.StageEmpty,
	}
	assert.Nil(t, createPendingRun(ctx, &empty))

	duplicate := shared.PendingTestRun{
		Created: yesterday,
		Updated: now.Add(time.Minute * -2),
		Stage:   shared.StageDuplicate,
	}
	assert.Nil(t, createPendingRun(ctx, &duplicate))

	received := shared.PendingTestRun{
		Created: yesterday.Add(time.Hour),
		Updated: now.Add(time.Minute * -10),
		Stage:   shared.StageWptFyiReceived,
	}
	assert.Nil(t, createPendingRun(ctx, &received))

	running := shared.PendingTestRun{
		Created: yesterday.Add(time.Hour),
		Updated: now.Add(time.Minute * -5),
		Stage:   shared.StageCIRunning,
	}
	assert.Nil(t, createPendingRun(ctx, &running))

	t.Run("/api/status", func(t *testing.T) {
		r, _ = i.NewRequest("GET", "/api/status", nil)
		resp := httptest.NewRecorder()
		apiPendingTestRunsHandler(resp, r)
		body, _ := io.ReadAll(resp.Result().Body)
		assert.Equal(t, http.StatusOK, resp.Code, string(body))
		var results []shared.PendingTestRun
		json.Unmarshal(body, &results)
		assert.Len(t, results, 5)
		// Sorted by Update.
		assert.Equal(t, results[0].ID, invalid.ID)
		assert.Equal(t, results[1].ID, empty.ID)
		assert.Equal(t, results[2].ID, duplicate.ID)
		assert.Equal(t, results[3].ID, running.ID)
		assert.Equal(t, results[4].ID, received.ID)
	})

	t.Run("/api/status/pending", func(t *testing.T) {
		r, _ = i.NewRequest("GET", "/api/status/pending", nil)
		r = mux.SetURLVars(r, map[string]string{"filter": "pending"})
		resp := httptest.NewRecorder()
		apiPendingTestRunsHandler(resp, r)
		body, _ := io.ReadAll(resp.Result().Body)
		assert.Equal(t, http.StatusOK, resp.Code, string(body))
		var results []shared.PendingTestRun
		json.Unmarshal(body, &results)
		assert.Len(t, results, 2)
		assert.Equal(t, results[0].ID, running.ID)
		assert.Equal(t, results[1].ID, received.ID)
	})

	filters := []string{"invalid", "empty", "duplicate"}
	runs := []*shared.PendingTestRun{&invalid, &empty, &duplicate}

	for index, filter := range filters {
		url := "/api/status/" + filter
		t.Run(url, func(t *testing.T) {
			r, _ = i.NewRequest("GET", url, nil)
			r = mux.SetURLVars(r, map[string]string{"filter": filter})
			resp := httptest.NewRecorder()
			apiPendingTestRunsHandler(resp, r)
			body, _ := io.ReadAll(resp.Result().Body)
			assert.Equal(t, http.StatusOK, resp.Code, string(body))
			var results []shared.PendingTestRun
			json.Unmarshal(body, &results)
			assert.Len(t, results, 1)
			assert.Equal(t, results[0].ID, runs[index].ID)
		})
	}
}

func TestAPIPendingTestHandler_invalidFilter(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()

	r, _ := i.NewRequest("GET", "/api/status/foobar", nil)
	r = mux.SetURLVars(r, map[string]string{"filter": "foobar"})
	resp := httptest.NewRecorder()
	apiPendingTestRunsHandler(resp, r)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}
