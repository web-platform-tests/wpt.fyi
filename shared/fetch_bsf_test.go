// +build small
// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilterandExtractBSFData_WithoutFilter(t *testing.T) {
	var rawBSFData [][]string
	fieldsRow := []string{"sha", "date", "chrome-version", "chrome", "firefox-version", "firefox", "safari-version", "safari"}
	dataRow := []string{"1", "2018-08-18", "70.0.3521.2 dev", "605.3869030161061", "63.0a1", "1521.908686731921", "12.1", "2966.686195133767"}
	rawBSFData = append(rawBSFData, fieldsRow)
	rawBSFData = append(rawBSFData, dataRow)

	result := FilterandExtractBSFData(rawBSFData, nil, nil)

	assert.Equal(t, "1", result.LastUpdateRevision)
	assert.Equal(t, fieldsRow, result.Fields)
	assert.Equal(t, 1, len(result.Data))
	assert.Equal(t, dataRow, result.Data[0])
}

func TestFilterandExtractBSFData_WithFilter(t *testing.T) {
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
	from, _ := time.Parse(time.RFC3339, "2018-04-19T00:00:00Z")
	to, _ := time.Parse(time.RFC3339, "2018-07-19T00:00:00Z")

	result := FilterandExtractBSFData(rawBSFData, &from, &to)

	assert.Equal(t, "3", result.LastUpdateRevision)
	assert.Equal(t, fieldsRow, result.Fields)
	assert.Equal(t, 2, len(result.Data))
	assert.Equal(t, dataRow2, result.Data[0])
	assert.Equal(t, dataRow3, result.Data[1])
}

func TestFilterandExtractBSFData_EmptyData(t *testing.T) {
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
	to, _ := time.Parse(time.RFC3339, "2017-07-19T00:00:00Z")

	result := FilterandExtractBSFData(rawBSFData, nil, &to)

	assert.Equal(t, BSFData{}, result)
}
