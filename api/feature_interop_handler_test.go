// Copyright 2026 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build small

package api //nolint:revive

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"go.uber.org/mock/gomock"
)

func TestFeatureInteropHandler_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	r := httptest.NewRequest("GET", "/api/interop-features", nil)
	w := httptest.NewRecorder()
	mockFetcher := sharedtest.NewMockFetchFeatureInterop(mockCtrl)

	fieldsRow := []string{"sha", "date", "manifest", "chrome-version", "chrome", "interop"}
	dataRow := []string{"9a1b2c3d", "2026-02-28", "merge_pr_54999", "146.0.7680.31", "88.1", "66.2"}
	featureFieldsRow := []string{"feature", "chrome", "interop", "tests"}
	featureRow := []string{"grid", "100.0", "91.3", "412"}
	mockFetcher.EXPECT().Fetch(gomock.Any(), false).Return(
		[][]string{fieldsRow, dataRow},
		[][]string{featureFieldsRow, featureRow},
		nil)

	FeatureInteropHandler{mockFetcher}.ServeHTTP(w, r)

	var data shared.FeatureInteropData
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal([]byte(w.Body.String()), &data))
	assert.Equal(t, "9a1b2c3d", data.LastUpdateRevision)
	assert.Equal(t, "merge_pr_54999", data.Manifest)
	assert.Equal(t, "2026-02-28", data.Date)
	assert.Equal(t, fieldsRow, data.Fields)
	assert.Equal(t, [][]string{dataRow}, data.Data)
	assert.Equal(t, featureFieldsRow, data.FeatureFields)
	assert.Equal(t, [][]string{featureRow}, data.Features)
}

func TestFeatureInteropHandler_Success_Experimental(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	r := httptest.NewRequest("GET", "/api/interop-features?experimental=true", nil)
	w := httptest.NewRecorder()
	mockFetcher := sharedtest.NewMockFetchFeatureInterop(mockCtrl)

	fieldsRow := []string{"sha", "date", "manifest", "chrome-version", "chrome", "interop"}
	dataRow := []string{"9a1b2c3d", "2026-02-28", "merge_pr_54999", "148.0.7700.0", "89.0", "67.1"}
	// The channel decides which pair of CSVs is fetched, so the param reaching
	// the fetcher is the whole of what this asserts.
	mockFetcher.EXPECT().Fetch(gomock.Any(), true).Return([][]string{fieldsRow, dataRow}, nil, nil)

	FeatureInteropHandler{mockFetcher}.ServeHTTP(w, r)

	var data shared.FeatureInteropData
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal([]byte(w.Body.String()), &data))
	assert.Equal(t, "2026-02-28", data.Date)
	assert.Empty(t, data.Features)
}

func TestFeatureInteropHandler_FetchError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	r := httptest.NewRequest("GET", "/api/interop-features", nil)
	w := httptest.NewRecorder()
	mockFetcher := sharedtest.NewMockFetchFeatureInterop(mockCtrl)

	mockFetcher.EXPECT().Fetch(gomock.Any(), false).Return(nil, nil, errors.New("gh-pages is down"))

	FeatureInteropHandler{mockFetcher}.ServeHTTP(w, r)

	// The graph says the data is unavailable rather than drawing an empty one,
	// so the status has to distinguish a failed fetch from a scored-nothing one.
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "gh-pages is down")
}
