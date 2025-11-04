//go:build small

// Copyright 2023 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestHistoryHandler(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	sampleRun := shared.TestHistoryEntry{
		BrowserName: "chrome",
		RunID:       "123",
		Date:        "2022-06-02T06:02:55.000Z",
		TestName:    "test name",
		SubtestName: "subtest",
		Status:      "PASS",
	}

	body :=
		`{
			"test_name": "test name"
		}`

	store := shared.NewAppEngineDatastore(ctx, false)
	key := store.NewIncompleteKey("TestHistoryEntry")
	_, err = store.Put(key, &sampleRun)
	assert.Nil(t, err)

	bodyReader := strings.NewReader(body)
	r := httptest.NewRequest("POST", "/api/history", bodyReader)
	w := httptest.NewRecorder()
	testHistoryHandler(w, r)
	results := parseHistoryResponse(t, w)

	want := map[string]map[string]Browser{
		"results": {
			"chrome": {
				"subtest": {
					{
						"date":   "2022-06-02T06:02:55.000Z",
						"status": "PASS",
						"run_id": "123",
					},
				},
			},
		},
	}

	assert.Equal(t, want, results)
}

func parseHistoryResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]map[string]Browser {
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	out, _ := io.ReadAll(w.Body)
	var result map[string]map[string]Browser
	json.Unmarshal(out, &result)
	return result
}
