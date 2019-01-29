// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

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
