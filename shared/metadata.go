// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"net/http"

	"github.com/web-platform-tests/wpt-metadata/util"
	"gopkg.in/yaml.v2"
)

// MetadataArchiveURL is the URL that retrieves an archive of wpt-metadata repository.
var MetadataArchiveURL = "https://api.github.com/repos/web-platform-tests/wpt-metadata/tarball"

// ShowMetadataParam determines whether Metadata Information returns along
// with a test result query request.
const ShowMetadataParam = "metadataInfo"

// MetadataResults is a map from test paths to all of the links under that test path.
// It represents a flattened copy of the wpt-metadata repository, which has metadata
// sharded across as large number of files in a directory structure.
type MetadataResults map[string]MetadataLinks

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
	Product ProductSpec          `yaml:"product" json:"product,omitempty"`
	URL     string               `yaml:"url"     json:"url"`
	Results []MetadataTestResult `yaml:"results" json:"results,omitempty"`
}

// MetadataTestResult is a filter for test results to which the metadata link
// should apply.
type MetadataTestResult struct {
	TestPath    string      `yaml:"test"    json:"test,omitempty"`
	SubtestName *string     `yaml:"subtest" json:"subtest,omitempty"`
	Status      *TestStatus `yaml:"status"  json:"status,omitempty"`
}

// GetMetadataResponse retrieves the response to a WPT Metadata query.
func GetMetadataResponse(testRuns []TestRun, client *http.Client, log Logger, url string) (MetadataResults, error) {
	var productSpecs = make([]ProductSpec, len(testRuns))
	for i, run := range testRuns {
		productSpecs[i] = ProductSpec{ProductAtRevision: run.ProductAtRevision, Labels: run.LabelsSet()}
	}

	metadata, err := GetMetadataByteMap(client, log, url)
	if err != nil {
		return nil, err
	}

	return constructMetadataResponse(productSpecs, metadata), nil
}

// GetMetadataResponseOnProducts constructs the response to a WPT Metadata query, given ProductSpecs.
func GetMetadataResponseOnProducts(productSpecs ProductSpecs, client *http.Client, log Logger, url string) (MetadataResults, error) {
	metadata, err := GetMetadataByteMap(client, log, url)
	if err != nil {
		return nil, err
	}

	return constructMetadataResponse(productSpecs, metadata), nil
}

// GetMetadataByteMap collects and parses all META.yml files from
// wpt-metadata reposiroty.
func GetMetadataByteMap(client *http.Client, log Logger, url string) (map[string]Metadata, error) {
	metadataByteMap, err := util.CollectMetadataWithURL(client, url)
	if err != nil {
		log.Errorf("Error from CollectMetadataWithURL: %s", err.Error())
		return nil, err
	}

	metadata := parseMetadata(metadataByteMap, log)
	return metadata, nil
}

func parseMetadata(metadataByteMap map[string][]byte, log Logger) map[string]Metadata {
	var metadataMap = make(map[string]Metadata)
	for path, data := range metadataByteMap {
		var metadata Metadata
		err := yaml.Unmarshal(data, &metadata)
		if err != nil {
			log.Warningf("Failed to unmarshal %s.", path)
			continue
		}
		metadataMap[path] = metadata
	}
	return metadataMap
}

// constructMetadataResponse constructs the response to a WPT Metadata query, given ProductSpecs.
func constructMetadataResponse(productSpecs ProductSpecs, metadata map[string]Metadata) MetadataResults {
	res := make(MetadataResults)
	for folderPath, data := range metadata {
		for i := range data.Links {
			link := data.Links[i]
			for _, result := range link.Results {
				//TODO(kyleju): Concatenate test path on WPT Metadata repository instead of here.
				var fullTestName = "/" + folderPath + "/" + result.TestPath
				for _, productSpec := range productSpecs {
					// Matches browser type if a version is not specified.
					if link.Product.MatchesProductSpec(productSpec) ||
						// Matches to all browsers if product is not specified.
						link.Product.BrowserName == "" {
						output := MetadataLink{
							Product: link.Product,
							URL:     link.URL,
						}
						if result.SubtestName != nil || result.Status != nil {
							output.Results = []MetadataTestResult{
								{
									SubtestName: result.SubtestName,
									Status:      result.Status,
									// TestPath is redundant (it's the map key)
								},
							}
						}
						if _, ok := res[fullTestName]; !ok {
							res[fullTestName] = MetadataLinks{output}
						} else {
							res[fullTestName] = append(res[fullTestName], output)
						}
						break
					}
				}
			}
		}
	}
	return res
}

// PrepareLinkFilter maps a MetadataResult test name to its URLs.
func PrepareLinkFilter(metadata MetadataResults) map[string][]string {
	metadataMap := make(map[string][]string)
	for test, links := range metadata {
		for _, link := range links {
			if urls, ok := metadataMap[test]; !ok {
				metadataMap[test] = []string{link.URL}
			} else {
				metadataMap[test] = append(urls, link.URL)
			}
		}
	}
	return metadataMap
}
