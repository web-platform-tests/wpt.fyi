// +build medium

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestApiSHAsHandler(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/api/shas", nil)
	assert.Nil(t, err)
	ctx := shared.NewAppEngineContext(r)

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
	store := shared.NewAppEngineDatastore(ctx, false)
	for _, browser := range browserNames {
		run.BrowserName = browser
		store.Put(store.NewIncompleteKey("TestRun"), &run)
	}
	run.FullRevisionHash = strings.Repeat("abcdef0000", 4)
	run.Revision = run.FullRevisionHash[:10]
	store.Put(store.NewIncompleteKey("TestRun"), &run)

	// Aligned
	shas = nil
	r, err = i.NewRequest("GET", "/api/shas?aligned", nil)
	w = httptest.NewRecorder()
	apiSHAsHandler(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, _ = ioutil.ReadAll(w.Result().Body)
	json.Unmarshal(bytes, &shas)
	assert.Equal(t, []string{"abcdef0123"}, shas)

	// Not aligned
	shas = nil
	r, err = i.NewRequest("GET", "/api/shas", nil)
	w = httptest.NewRecorder()
	apiSHAsHandler(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	bytes, _ = ioutil.ReadAll(w.Result().Body)
	json.Unmarshal(bytes, &shas)
	assert.Equal(t, []string{"abcdef0123", "abcdef0000"}, shas)

	// Bad param
	r, err = i.NewRequest("GET", "/api/shas?aligned=bad-value", nil)
	w = httptest.NewRecorder()
	apiSHAsHandler(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
