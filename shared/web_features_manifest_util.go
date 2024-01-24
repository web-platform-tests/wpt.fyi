// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"io"
)

// getWPTWebFeaturesManifest contains the common logic to handle retrieving a
// web features manifest.
// It includes two generic steps: Downloading and Parsing.
func getWPTWebFeaturesManifest(
	ctx context.Context,
	downloader webFeaturesManifestDownloader,
	parser webFeatureManifestParser) (WebFeaturesData, error) {
	manifest, err := downloader.Download(ctx)
	if err != nil {
		return nil, err
	}
	data, err := parser.Parse(ctx, manifest)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// webFeaturesManifestDownloader provides an interface for downloading a manifest file.
// Typically implementers of the interface create a struct based on the location of the file.
// For example: GitHub.
type webFeaturesManifestDownloader interface {
	Download(context.Context) (io.ReadCloser, error)
}

// webFeatureManifestParser provides an interface on how to parse a given manifest.
// Typically implementers of the interface create a struct per type of file.
// For example: JSON file or YAML file.
type webFeatureManifestParser interface {
	// Parse reads the stream of data and returns a map.
	// The map mirrors the structure of MetadataResults where the key is the
	// test name and the value is the actual data. In this case the "data" is a
	// collection of applicable web features
	Parse(context.Context, io.ReadCloser) (WebFeaturesData, error)
}
