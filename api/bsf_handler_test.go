//go:build small

// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api //nolint:revive

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"go.uber.org/mock/gomock"
)

func TestBSFHandler_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	r := httptest.NewRequest("GET", "/api/bsf?", nil)
	w := httptest.NewRecorder()
	mockBSFFetcher := sharedtest.NewMockFetchBSF(mockCtrl)

	var rawBSFData [][]string
	fieldsRow := []string{"sha", "date", "chrome-version", "chrome", "firefox-version", "firefox", "safari-version", "safari"}
	dataRow := []string{"1", "2018-08-18", "70.0.3521.2 dev", "605.3869030161061", "63.0a1", "1521.908686731921", "12.1", "2966.686195133767"}
	rawBSFData = append(rawBSFData, fieldsRow)
	rawBSFData = append(rawBSFData, dataRow)
	mockBSFFetcher.EXPECT().Fetch(false).Return(rawBSFData, nil)

	BSFHandler{mockBSFFetcher}.ServeHTTP(w, r)

	var bsfData shared.BSFData
	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal([]byte(w.Body.String()), &bsfData)
	assert.Equal(t, "1", bsfData.LastUpdateRevision)
	assert.Equal(t, fieldsRow, bsfData.Fields)
	assert.Equal(t, 1, len(bsfData.Data))
	assert.Equal(t, dataRow, bsfData.Data[0])
}

func TestBSFHandler_Success_WithParams(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	r := httptest.NewRequest("GET", "/api/bsf?experimental=true&from=2018-04-01T00%3A00%3A00Z&to=2018-07-01T00%3A00%3A00Z", nil)
	w := httptest.NewRecorder()
	mockBSFFetcher := sharedtest.NewMockFetchBSF(mockCtrl)

	var rawBSFData [][]string
	fieldsRow := []string{"sha", "date", "chrome-version", "chrome", "firefox-version", "firefox", "safari-version", "safari"}
	dataRow1 := []string{"1", "2018-03-18", "70.0.3521.2 dev", "605.3869030161061", "63.0a1", "1521.908686731921", "12.1", "2966.686195133767"}
	dataRow2 := []string{"2", "2018-05-17", "70.0.3521.2 dev", "605.3869030161061", "63.0a1", "1521.908686731921", "12.1", "2966.686195133767"}
	dataRow3 := []string{"3", "2018-05-19", "70.0.3521.2 dev", "605.3869030161061", "63.0a1", "1521.908686731921", "12.1", "2966.686195133767"}
	dataRow4 := []string{"4", "2018-11-18", "70.0.3521.2 dev", "605.3869030161061", "63.0a1", "1521.908686731921", "12.1", "2966.686195133767"}
	rawBSFData = append(rawBSFData, fieldsRow)
	rawBSFData = append(rawBSFData, dataRow1)
	rawBSFData = append(rawBSFData, dataRow2)
	rawBSFData = append(rawBSFData, dataRow3)
	rawBSFData = append(rawBSFData, dataRow4)
	mockBSFFetcher.EXPECT().Fetch(true).Return(rawBSFData, nil)

	BSFHandler{mockBSFFetcher}.ServeHTTP(w, r)

	var bsfData shared.BSFData
	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal([]byte(w.Body.String()), &bsfData)
	assert.Equal(t, "3", bsfData.LastUpdateRevision)
	assert.Equal(t, fieldsRow, bsfData.Fields)
	assert.Equal(t, 2, len(bsfData.Data))
	assert.Equal(t, dataRow2, bsfData.Data[0])
	assert.Equal(t, dataRow3, bsfData.Data[1])
}
