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

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

// TestApiInteropHandler_CompleteRunFallback tests that when a ?complete param
// is requested, but the most-recent complete run doesn't have interop computed,
// we fall back to the most-recent complete run that does have interop.
func TestApiInteropHandler_CompleteRunFallback(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()

	r, _ := i.NewRequest("GET", "/api/interop?complete", nil)
	ctx := r.Context()

	firstRun := shared.TestRun{}
	firstRun.Labels = []string{"stable"}
	firstRun.FullRevisionHash = strings.Repeat("0000000000", 4)
	firstRun.Revision = firstRun.FullRevisionHash[:10]
	firstRun.TimeStart = time.Now().AddDate(0, 0, -1)

	secondRun := shared.TestRun{}
	secondRun.FullRevisionHash = strings.Repeat("1111111111", 4)
	secondRun.Revision = secondRun.FullRevisionHash[:10]
	secondRun.TimeStart = time.Now()

	products := shared.GetDefaultProducts()
	store := shared.NewAppEngineDatastore(ctx, false)
	firstRunKeys := make([]shared.Key, len(products))
	secondRunKeys := make([]shared.Key, len(products))
	for i, product := range products {
		run := firstRun
		run.Product = product.Product
		firstRunKeys[i], _ = store.Put(store.NewIncompleteKey("TestRun"), &run)
		run = secondRun
		run.Product = product.Product
		secondRunKeys[i], _ = store.Put(store.NewIncompleteKey("TestRun"), &run)
	}

	// No interop data.
	resp := httptest.NewRecorder()
	apiInteropHandler(resp, r)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	// Interop data spanning across the complete runs.
	interop := metrics.PassRateMetadataLegacy{}
	interop.TestRunIDs = make(shared.TestRunIDs, len(firstRunKeys))
	for i := range firstRunKeys {
		key := firstRunKeys[i]
		if i*2 < len(firstRunKeys) {
			key = secondRunKeys[i]
		}
		interop.TestRunIDs[i] = key.IntID()
	}
	interopKindName := metrics.GetDatastoreKindName(metrics.PassRateMetadata{})
	store.Put(store.NewIncompleteKey(interopKindName), &interop)

	resp = httptest.NewRecorder()
	apiInteropHandler(resp, r)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	// One interop data, for the first run.
	interop = metrics.PassRateMetadataLegacy{}
	interop.TestRunIDs = make(shared.TestRunIDs, len(firstRunKeys))
	for i, key := range firstRunKeys {
		interop.TestRunIDs[i] = key.IntID()
	}
	store.Put(store.NewIncompleteKey(interopKindName), &interop)

	// Load the tests + clear the IDs, to match the output of apiInteropHandler below.
	interop.LoadTestRuns(ctx)
	interop.TestRunIDs = nil

	// "complete" and "complete & stable" have the same outcome.
	reqs := make([]*http.Request, 2)
	reqs[0], _ = i.NewRequest("GET", "/api/interop?complete", nil)
	reqs[1], _ = i.NewRequest("GET", "/api/interop?complete&label=stable", nil)
	reqs[1], _ = i.NewRequest("GET", "/api/interop?label=stable&sha=0000000000", nil)
	for _, req := range reqs {
		resp = httptest.NewRecorder()
		apiInteropHandler(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		var bodyInterop metrics.PassRateMetadataLegacy
		json.Unmarshal(bodyBytes, &bodyInterop)
		assert.Equal(t, interop, bodyInterop)
	}
}
