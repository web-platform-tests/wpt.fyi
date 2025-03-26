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

// --- Test Data ---

// Sample (valid) WebFeaturesData
var sampleWebFeaturesData = WebFeaturesData{
	"test1": {"featureA": nil, "featureB": nil},
}

type mockDownloader struct {
	returnData io.ReadCloser
	returnErr  error
}

func (m *mockDownloader) Download(ctx context.Context) (io.ReadCloser, error) {
	return m.returnData, m.returnErr
}

// --- Mocked Parser ---

type mockParser struct {
	returnData WebFeaturesData
	returnErr  error
}

func (m *mockParser) Parse(ctx context.Context, manifest io.ReadCloser) (WebFeaturesData, error) {
	return m.returnData, m.returnErr
}

func TestGetWPTWebFeaturesManifest(t *testing.T) {
	testCases := []struct {
		name           string
		mockDownloader *mockDownloader
		mockParser     *mockParser
		expectedData   WebFeaturesData
		expectedError  error
	}{
		{
			name:           "Success",
			mockDownloader: &mockDownloader{returnData: io.NopCloser(strings.NewReader(`{}`))},
			mockParser:     &mockParser{returnData: sampleWebFeaturesData},
			expectedData:   sampleWebFeaturesData,
			expectedError:  nil,
		},
		{
			name:           "Downloader Error",
			mockDownloader: &mockDownloader{returnErr: errors.New("download failed")},
			// ... (mockParser doesn't matter in this case)
			expectedError: errors.New("download failed"),
		},
		{
			name:           "Parser Error",
			mockDownloader: &mockDownloader{returnData: io.NopCloser(strings.NewReader(`{}`))},
			mockParser:     &mockParser{returnErr: errors.New("parsing failed")},
			expectedError:  errors.New("parsing failed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := getWPTWebFeaturesManifest(context.Background(), tc.mockDownloader, tc.mockParser)
			assert.Equal(t, tc.expectedData, result)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
