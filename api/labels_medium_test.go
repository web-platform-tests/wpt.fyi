// +build medium

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

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

func TestLabelsHandler(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
	_, err = datastore.Put(ctx, key, &shared.TestRun{Labels: []string{"a"}})
	_, err = datastore.Put(ctx, key, &shared.TestRun{Labels: []string{"b", "c"}})
	assert.Nil(t, err)

	handler := LabelsHandler{ctx: ctx}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/labels", nil)
	handler.ServeHTTP(w, r)
	labels := parseLabelsResponse(t, w)
	assert.Equal(t, labels, []string{"a", "b", "c"})
}

func TestLabelsHandler_Caches(t *testing.T) {
	instance, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer instance.Close()

	r, _ := instance.NewRequest("GET", "/api/labels", nil)
	ctx := shared.NewAppEngineContext(r)

	key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
	_, err = datastore.Put(ctx, key, &shared.TestRun{
		Labels: []string{"a"},
	})
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	apiLabelsHandler(w, r)
	labels := parseLabelsResponse(t, w)
	assert.Equal(t, labels, []string{"a"})

	// Should cache; add a "b" and don't find it.
	_, err = datastore.Put(ctx, key, &shared.TestRun{
		Labels: []string{"b"},
	})
	apiLabelsHandler(w, r)
	labels = parseLabelsResponse(t, w)
	assert.Equal(t, labels, []string{"a"})
}

func parseLabelsResponse(t *testing.T, w *httptest.ResponseRecorder) []string {
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	out, _ := ioutil.ReadAll(w.Body)
	var labels []string
	json.Unmarshal(out, &labels)
	return labels
}
