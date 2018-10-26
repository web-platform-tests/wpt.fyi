// +build medium

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine/datastore"
)

func TestApiVersionsHandler(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/versions", nil)
	assert.Nil(t, err)
	ctx := shared.NewAppEngineContext(r)

	// No results - empty JSON array, 404
	var versions []string
	r, err = i.NewRequest("GET", "/api/versions?product=chrome-999", nil)
	w := httptest.NewRecorder()
	apiVersionsHandler(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
	bytes, _ := ioutil.ReadAll(w.Result().Body)
	json.Unmarshal(bytes, &versions)
	assert.Equal(t, []string{}, versions)

	// Add test runs (duplicating 1.1 is deliberate)
	someVersions := []string{"2", "1.1.1", "1.1", "1.1", "1.0", "1"}
	run := shared.TestRun{}
	browserNames := shared.GetDefaultBrowserNames()
	for _, browser := range browserNames {
		run.BrowserName = browser
		for _, version := range someVersions {
			run.BrowserVersion = version
			datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
		}
	}

	// Chrome
	versions = nil
	r, err = i.NewRequest("GET", "/api/versions?product=chrome", nil)
	w = httptest.NewRecorder()
	apiVersionsHandler(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, _ = ioutil.ReadAll(w.Result().Body)
	json.Unmarshal(bytes, &versions)
	// Duplication should be removed.
	assert.Equal(t, []string{"2", "1.1.1", "1.1", "1.0", "1"}, versions)

	// Chrome 1.1
	r, err = i.NewRequest("GET", "/api/versions?product=chrome-1", nil)
	w = httptest.NewRecorder()
	apiVersionsHandler(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, _ = ioutil.ReadAll(w.Result().Body)
	json.Unmarshal(bytes, &versions)
	assert.Equal(t, []string{"1.1.1", "1.1", "1.0", "1"}, versions)

	// No product param
	r, err = i.NewRequest("GET", "/api/versions", nil)
	w = httptest.NewRecorder()
	apiVersionsHandler(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Bad param
	r, err = i.NewRequest("GET", "/api/versions?product=chrome-not.a.version", nil)
	w = httptest.NewRecorder()
	apiVersionsHandler(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
