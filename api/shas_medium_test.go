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
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func TestApiSHAsHandler(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/shas", nil)
	assert.Nil(t, err)
	ctx := appengine.NewContext(r)

	// No results - empty JSON array, 404
	var shas []string
	r, err = i.NewRequest("GET", "/api/shas", nil)
	w := httptest.NewRecorder()
	apiSHAsHandler(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
	bytes, _ := ioutil.ReadAll(w.Result().Body)
	json.Unmarshal(bytes, &shas)
	assert.Equal(t, []string{}, shas)

	// Add test runs
	browserNames := shared.GetDefaultBrowserNames()
	run := shared.TestRun{
		ProductAtRevision: shared.ProductAtRevision{
			Revision: "abcdef0123",
		},
	}
	for _, browser := range browserNames {
		run.BrowserName = browser
		datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)
	}
	run.Revision = "abcdef0000"
	datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "TestRun", nil), &run)

	// Complete
	shas = nil
	r, err = i.NewRequest("GET", "/api/shas?complete", nil)
	w = httptest.NewRecorder()
	apiSHAsHandler(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, _ = ioutil.ReadAll(w.Result().Body)
	json.Unmarshal(bytes, &shas)
	assert.Equal(t, []string{"abcdef0123"}, shas)

	// Not complete
	shas = nil
	r, err = i.NewRequest("GET", "/api/shas", nil)
	w = httptest.NewRecorder()
	apiSHAsHandler(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, _ = ioutil.ReadAll(w.Result().Body)
	json.Unmarshal(bytes, &shas)
	assert.Equal(t, []string{"abcdef0123", "abcdef0000"}, shas)

	// Bad param
	r, err = i.NewRequest("GET", "/api/shas?complete=bad-value", nil)
	w = httptest.NewRecorder()
	apiSHAsHandler(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
