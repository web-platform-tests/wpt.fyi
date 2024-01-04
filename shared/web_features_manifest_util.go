// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/web_features_manifest_util_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared WebFeaturesManifestDownloader,WebFeatureManifestParser

package shared

import (
	"context"
	"io"
)

// WebFeatureManifestRepo is where the web feature manifest is published
const WebFeatureManifestRepo = "wpt"

const webFeaturesManifestFilename = "WEB_FEATURES_MANIFEST.json.gz"

// GetWPTWebFeaturesManifest is the entrypoint for handling web features manifest downloads.
// It includes two generic steps: Downloading and Parsing.
func GetWPTWebFeaturesManifest(ctx context.Context, downloader WebFeaturesManifestDownloader, parser WebFeatureManifestParser) (WebFeaturesData, error) {
	manifest, err := downloader.Download(ctx)
	if err != nil {
		return nil, nil
	}
	data, err := parser.Parse(ctx, manifest)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// WebFeaturesManifestDownloader provides an interface for downloading a manifest file.
// Typically implementers of the interface create a struct based on the location of the file.
// For example: GitHub.
type WebFeaturesManifestDownloader interface {
	Download(context.Context) (io.ReadCloser, error)
}

// WebFeatureManifestParser provides an interface on how to parse a given manifest.
// Typically implementers of the interface create a struct per type of file.
// For example: JSON file or YAML file
type WebFeatureManifestParser interface {
	// Parse reads the stream of data and returns a map.
	// The map mirrors the structure of MetadataResults where the key is the
	// test name and the value is the actual data. In this case the "data" is a
	// collection of applicable web features
	Parse(context.Context, io.ReadCloser) (WebFeaturesData, error)
}
