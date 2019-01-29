// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

type Metadata struct {
	Links MetadataLinks
}

type MetadataLinks []MetadataLink

type MetadataLink struct {
	Product  ProductSpec
	TestPath string `yaml:"test"`
	URL      string
}
