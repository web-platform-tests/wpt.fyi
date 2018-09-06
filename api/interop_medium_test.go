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
	"github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// TestApiInteropHandler_CompleteRunFallback tests that when a ?complete param
// is requested, but the most-recent complete run doesn't have interop computed,
// we fall back to the most-recent complete run that does have interop.
func TestApiInteropHandler_CompleteRunFallback(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()

	r, err := i.NewRequest("GET", "/api/interop", nil)
	assert.Nil(t, err)
	ctx := appengine.NewContext(r)

	firstRun := shared.TestRun{}
	firstRun.Revision = "0000000000"
	firstRun.TimeStart = time.Now().AddDate(0, 0, -1)

	secondRun := shared.TestRun{}
	secondRun.Revision = "1111111111"
	secondRun.TimeStart = time.Now()

	products := shared.GetDefaultProducts()
	firstRunKeys := make([]*datastore.Key, len(products))
	secondRunKeys := make([]*datastore.Key, len(products))
	for i, product := range products {
		run := firstRun
		run.Product = product.Product
		firstRunKeys[i], _ = datastore.Put(ctx, datastore.NewKey(ctx, "TestRun", "", 0, nil), &run)
		run = secondRun
		run.Product = product.Product
		secondRunKeys[i], _ = datastore.Put(ctx, datastore.NewKey(ctx, "TestRun", "", 0, nil), &run)
	}

	// No interop data
	resp := httptest.NewRecorder()
	apiInteropHandler(resp, r)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	// One interop data, for the first run.
	interop := metrics.PassRateMetadata{}
	interop.TestRunIDs = make(shared.TestRunIDs, len(firstRunKeys))
	for i, key := range firstRunKeys {
		interop.TestRunIDs[i] = key.IntID()
	}
	interopKindName := metrics.GetDatastoreKindName(metrics.PassRateMetadata{})
	datastore.Put(ctx, datastore.NewKey(ctx, interopKindName, "", 0, nil), &interop)

	// Needed for equality comparisons below.
	interop.LoadTestRuns(ctx)
	interop.TestRunIDs = nil

	// Latest run
	resp = httptest.NewRecorder()
	apiInteropHandler(resp, r)
	assert.Equal(t, http.StatusOK, resp.Code)
	var bodyInterop metrics.PassRateMetadata
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &bodyInterop)
	assert.Equal(t, interop, bodyInterop)

	// Latest complete run
	r, _ = i.NewRequest("GET", "/api/interop?complete", nil)
	resp = httptest.NewRecorder()
	apiInteropHandler(resp, r)
	assert.Equal(t, http.StatusOK, resp.Code)
	bodyBytes, _ = ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &bodyInterop)
	assert.Equal(t, interop, bodyInterop)
}
