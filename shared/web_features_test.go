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
			data:       WebFeaturesData{"test1": map[string]interface{}{"feature1": nil, "feature2": nil}},
			test:       "test1",
			webFeature: "feature1",
			expected:   true,
		},
		{
			name:       "test doesn't match with web feature",
			data:       WebFeaturesData{"test1": map[string]interface{}{"feature1": nil, "feature2": nil}},
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
		{
			name: "test name with uppercase characters matches exact case",
			data: WebFeaturesData{
				"/custom-elements/Document-createElement-customized-builtins.html": map[string]interface{}{
					"customized-built-in-elements": nil,
				},
			},
			test:       "/custom-elements/Document-createElement-customized-builtins.html",
			webFeature: "customized-built-in-elements",
			expected:   true,
		},
		{
			name: "test name case must match exactly",
			data: WebFeaturesData{
				"/custom-elements/Document-createElement-customized-builtins.html": map[string]interface{}{
					"customized-built-in-elements": nil,
				},
			},
			test:       "/custom-elements/document-createelement-customized-builtins.html",
			webFeature: "customized-built-in-elements",
			expected:   false,
		},
		{
			name: "web feature name remains case-insensitive",
			data: WebFeaturesData{
				"/custom-elements/Document-createElement-customized-builtins.html": map[string]interface{}{
					"customized-built-in-elements": nil,
				},
			},
			test:       "/custom-elements/Document-createElement-customized-builtins.html",
			webFeature: "Customized-Built-In-Elements",
			expected:   true,
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
			name:      "valid manifest version 1",
			inputJSON: `{"version": 1, "data": {"feature1": ["/test1", "/test2"], "feature2": ["/test1", "/test3"]}}`,
			expectedData: WebFeaturesData{
				"/test1": map[string]interface{}{
					"feature1": nil,
					"feature2": nil,
				},
				"/test2": map[string]interface{}{
					"feature1": nil,
				},
				"/test3": map[string]interface{}{
					"feature2": nil,
				},
			},
			expectedError: nil,
		},
		{
			name:          "invalid manifest JSON",
			inputJSON:     `{"version": 1, "data": invalid}`,
			expectedData:  nil,
			expectedError: errBadWebFeaturesManifestJSON,
		},
		{
			name:          "invalid manifest v1 JSON",
			inputJSON:     `{"version": 1, "data": "invalid"}`,
			expectedData:  nil,
			expectedError: errUnexpectedWebFeaturesManifestV1Format,
		},
		{
			name:          "unknown manifest version",
			inputJSON:     `{"version": 2}`,
			expectedData:  nil,
			expectedError: errUnknownWebFeaturesManifestVersion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parser := webFeaturesManifestJSONParser{}
			r := io.NopCloser(strings.NewReader(tc.inputJSON))
			data, err := parser.Parse(context.Background(), r)
			if !errors.Is(err, tc.expectedError) {
				t.Errorf("Parse() returned unexpected error: (%v). expected error: (%v).", err, tc.expectedError)
			}
			assert.Equal(t, tc.expectedData, data)
		})
	}
}

func TestWebFeaturesManifestV1Data_prepareTestWebFeatureFilter(t *testing.T) {
	// Test cases for prepareTestWebFeatureFilter
	data := webFeaturesManifestV1Data{
		"feature1": []string{
			"/test1",
			"/test2",
		},
		"feature2": []string{
			"/test2",
		},
		"Feature3": []string{
			"/Test3",
		}}
	expectedResult := WebFeaturesData{
		"/test1": {"feature1": nil},
		"/test2": {"feature1": nil, "feature2": nil},
		"/Test3": {"feature3": nil},
	}
	result := data.prepareTestWebFeatureFilter()
	assert.Equal(t, expectedResult, result)
}
