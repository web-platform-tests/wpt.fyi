// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build small

package shared

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestWebFeaturesData_TestMatchesWithWebFeature(t *testing.T) {
	// Test cases for TestMatchesWithWebFeature
	tests := []struct {
		name       string
		data       WebFeaturesData
		test       string
		webFeature string
		expected   bool
	}{
		{
			name:       "test matches with web feature",
			data:       WebFeaturesData{"test1": []string{"feature1", "feature2"}},
			test:       "test1",
			webFeature: "feature1",
			expected:   true,
		},
		{
			name:       "test doesn't match with web feature",
			data:       WebFeaturesData{"test1": []string{"feature1", "feature2"}},
			test:       "test1",
			webFeature: "feature3",
			expected:   false,
		},
		{
			name:       "test not present in data",
			data:       WebFeaturesData{},
			test:       "test1",
			webFeature: "feature1",
			expected:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.data.TestMatchesWithWebFeature(tc.test, tc.webFeature)
			assert.Equal(t, result, tc.expected)
		})
	}
}

func TestWebFeaturesManifestJSONParser_Parse(t *testing.T) {
	// Test cases for Parse
	tests := []struct {
		name          string
		inputJSON     string
		expectedData  WebFeaturesData
		expectedError error
	}{
		{
			name:          "valid manifest version 1",
			inputJSON:     `{"version": 1, "data": {"feature1": ["test1", "test2"]}}`,
			expectedData:  WebFeaturesData{"test1": []string{"feature1"}, "test2": []string{"feature1"}},
			expectedError: nil,
		},
		{
			name:          "invalid manifest JSON",
			inputJSON:     `{"version": 1, "data": invalid}`,
			expectedData:  nil,
			expectedError: ErrBadWebFeaturesManifestJSON,
		},
		{
			name:          "invalid manifest v1 JSON",
			inputJSON:     `{"version": 1, "data": "invalid"}`,
			expectedData:  nil,
			expectedError: ErrUnexpectedWebFeaturesManifestV1Format,
		},
		{
			name:          "unknown manifest version",
			inputJSON:     `{"version": 2}`,
			expectedData:  nil,
			expectedError: ErrUnknownWebFeaturesManifestVersion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parser := WebFeaturesManifestJSONParser{}
			r := io.NopCloser(strings.NewReader(tc.inputJSON))
			data, err := parser.Parse(context.Background(), r)
			if !errors.Is(err, tc.expectedError) {
				t.Errorf("Parse() returned unexpected error: (%v). expected error: (%v).", err, tc.expectedError)
			}
			assert.Equal(t, data, tc.expectedData)
		})
	}
}

// Ensures order is kept for equality purposes
func sortValuesInMap(input *map[string][]string) {
	for test, features := range *input {
		slices.Sort(features)
		(*input)[test] = features
	}
}

func TestWebFeaturesManifestV1Data_prepareTestWebFeatureFilter(t *testing.T) {
	// Test cases for prepareTestWebFeatureFilter
	data := webFeaturesManifestV1Data{"feature1": []string{"test1", "test2"}, "feature2": []string{"test2"}}
	expectedResult := map[string][]string{"test1": {"feature1"}, "test2": {"feature1", "feature2"}}
	result := data.prepareTestWebFeatureFilter()
	sortValuesInMap(&result)
	assert.Equal(t, expectedResult, result)
}
