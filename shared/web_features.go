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
// a given manifest file.
type WebFeaturesData map[string]map[string]interface{}

// errUnknownWebFeaturesManifestVersion indicates that the parser does not know how to parse
// this version of the web features file.
var errUnknownWebFeaturesManifestVersion = errors.New("unknown web features manifest version")

// errBadWebFeaturesManifestJSON indicates that there was an error parsing the given
// v1 manifest file.
var errBadWebFeaturesManifestJSON = errors.New("invalid json when reading web features manifest")

// errUnexpectedWebFeaturesManifestV1Format indicates that there was an error parsing the given
// v1 manifest file.
var errUnexpectedWebFeaturesManifestV1Format = errors.New("unexpected web features manifest v1 format")

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
type webFeaturesManifestV1Data map[string][]string

// webFeaturesManifestJSONParser is parser that can interpret a Web Features Manifest in a JSON format.
type webFeaturesManifestJSONParser struct{}

// Parse parses a JSON file into a WebFeaturesData instance.
func (p webFeaturesManifestJSONParser) Parse(_ context.Context, r io.ReadCloser) (WebFeaturesData, error) {
	defer r.Close()
	file := new(webFeaturesManifestFile)
	err := json.NewDecoder(r).Decode(file)
	if err != nil {
		return nil, errors.Join(errBadWebFeaturesManifestJSON, err)
	}

	switch file.Version {
	case 1:
		data := new(webFeaturesManifestV1Data)
		err = json.Unmarshal(file.Data, data)
		if err != nil {
			return nil, errors.Join(errUnexpectedWebFeaturesManifestV1Format, err)
		}

		return data.prepareTestWebFeatureFilter(), nil
	}

	return nil, fmt.Errorf("bad version %d %w", file.Version, errUnknownWebFeaturesManifestVersion)
}

// prepareTestWebFeatureFilter maps a MetadataResult test name to its web features.
func (d webFeaturesManifestV1Data) prepareTestWebFeatureFilter() WebFeaturesData {
	// Create a map where the value is effectively a set (map[string]interface{})
	testToWebFeaturesMap := make(map[string]map[string]interface{})
	for webFeature, tests := range d {
		for _, test := range tests {
			key := test
			value := strings.ToLower(webFeature)
			if set, found := testToWebFeaturesMap[key]; found {
				set[value] = nil
				testToWebFeaturesMap[key] = set
			} else {
				testToWebFeaturesMap[key] = map[string]interface{}{value: nil}
			}
		}
	}

	return testToWebFeaturesMap
}
