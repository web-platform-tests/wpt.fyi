// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// WebFeaturesData is the public data type that represents the data parsed from
// a given manifest file
type WebFeaturesData map[string]map[string]interface{}

// ErrUnknownWebFeaturesManifestVersion indicates that the parser does not know how to parse
// this version of the web features file.
var ErrUnknownWebFeaturesManifestVersion = errors.New("unknown web features manifest version")

// ErrBadWebFeaturesManifestJSON indicates that there was an error parsing the given
// v1 manifest file.
var ErrBadWebFeaturesManifestJSON = errors.New("invalid json when reading web features manifest")

// ErrUnexpectedWebFeaturesManifestV1Format indicates that there was an error parsing the given
// v1 manifest file.
var ErrUnexpectedWebFeaturesManifestV1Format = errors.New("unexpected web features manifest v1 format")

// TestMatchesWithWebFeature performs two checks.
// If the given test path is present in the data. If not, return false
// If it is present, check if the given web feature applies to that test.
func (d WebFeaturesData) TestMatchesWithWebFeature(test, webFeature string) bool {
	if len(d) == 0 {
		return false
	}
	if webFeatures, ok := d[test]; ok {
		_, found := webFeatures[strings.ToLower(webFeature)]

		return found
	}

	return false
}

// webFeaturesManifestFile is the base format for any manifest file.
type webFeaturesManifestFile struct {
	Version int             `json:"version,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// webFeaturesManifestV1Data represents the data section in the manifest file
// given manifest version 1.
// The data is a simple map.
// The key is the web feature key
// The value is a list of tests that are part of that web feature.
type webFeaturesManifestV1Data map[string][]webFeaturesManifestV1DataTest

type webFeaturesManifestV1DataTest struct {
	Path string  `json:"path,omitempty"`
	URL  *string `json:"url,omitempty"`
}

// WebFeaturesManifestJSONParser is parser that can interpret a Web Features Manifest in a JSON format
type WebFeaturesManifestJSONParser struct{}

// Parse parses a JSON file into a WebFeaturesData instance.
func (p WebFeaturesManifestJSONParser) Parse(ctx context.Context, r io.ReadCloser) (WebFeaturesData, error) {
	defer r.Close()
	file := webFeaturesManifestFile{}
	err := json.NewDecoder(r).Decode(&file)
	if err != nil {
		return nil, errors.Join(ErrBadWebFeaturesManifestJSON, err)
	}

	switch file.Version {
	case 1:
		data := new(webFeaturesManifestV1Data)
		err = json.Unmarshal(file.Data, data)
		if err != nil {
			return nil, errors.Join(ErrUnexpectedWebFeaturesManifestV1Format, err)
		}

		return data.prepareTestWebFeatureFilter(), nil
	}

	return nil, fmt.Errorf("bad version %d %w", file.Version, ErrUnknownWebFeaturesManifestVersion)
}

// PrepareTestWebFeatureFilter maps a MetadataResult test name to its web features.
func (d webFeaturesManifestV1Data) prepareTestWebFeatureFilter() map[string]map[string]interface{} {
	// Create a map where the value is effectively a set (map[string]interface{})
	testToWebFeaturesMap := make(map[string]map[string]interface{})
	for webFeature, tests := range d {
		for _, test := range tests {
			if test.URL == nil {
				// wpt.fyi only cares about the url. If it doesn't have it, it is probably a support file.
				continue
			}
			key := strings.ToLower(*test.URL)
			value := strings.ToLower(webFeature)
			testToWebFeaturesMap[key] = map[string]interface{}{value: nil}
		}
	}

	return testToWebFeaturesMap
}
