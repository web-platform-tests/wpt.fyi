// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"github.com/go-yaml/yaml"
	"github.com/web-platform-tests/wpt-metadata/util"
)

// Metadata represents a wpt-metadata META.yml file.
type Metadata struct {
	Links MetadataLinks
}

// MetadataLinks is a helper type for a MetadataLink slice.
type MetadataLinks []MetadataLink

// MetadataLink is an item in the `links` node of a wpt-metadata
// META.yml file, which lists an external reference, optionally
// filtered by product and a specific test.
type MetadataLink struct {
	Product  ProductSpec
	TestPath string `yaml:"test"`
	URL      string
}

// RetrieveMetadata collects and parses all META.yml files from
// wpt-metadata reposiroty.
func RetrieveMetadata() map[string]Metadata {
	return parseMetadata(util.CollectMetadata())
}

// parseMetadata implements the parsing logic.
func parseMetadata(metadataByteMap map[string][]byte) map[string]Metadata {
	var metadataMap = make(map[string]Metadata)

	for path, data := range metadataByteMap {
		var metadata Metadata
		err := yaml.Unmarshal(data, &metadata)
		if err != nil {
			panic(err)
		}
		metadataMap[path] = metadata
	}
	return metadataMap
}
