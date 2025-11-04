//go:build medium

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestLabelsHandler(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	store := shared.NewAppEngineDatastore(ctx, false)
	key := store.NewIncompleteKey("TestRun")
	_, err = store.Put(key, &shared.TestRun{Labels: []string{"b", "c"}})
	_, err = store.Put(key, &shared.TestRun{Labels: []string{"b", "a"}})
	assert.Nil(t, err)

	handler := LabelsHandler{ctx: ctx}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/labels", nil)
	handler.ServeHTTP(w, r)
	labels := parseLabelsResponse(t, w)
	assert.Equal(t, labels, []string{"a", "b", "c"}) // Ordered and deduped
}

func parseLabelsResponse(t *testing.T, w *httptest.ResponseRecorder) []string {
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	out, _ := io.ReadAll(w.Body)
	var labels []string
	json.Unmarshal(out, &labels)
	return labels
}
