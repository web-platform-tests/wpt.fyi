// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"net/http"
	"sort"

	"github.com/go-yaml/yaml"
	"github.com/web-platform-tests/wpt-metadata/util"
)

// MetadataArchiveURL is the URL that retrieves an archive of wpt-metadata repository.
var MetadataArchiveURL = "https://api.github.com/repos/web-platform-tests/wpt-metadata/tarball"

// ShowMetadataParam determines whether Metadata Information returns along
// with a test result query request.
const ShowMetadataParam = "metadataInfo"

// MetadataResults is a helper type for a MetadataResult slice.
type MetadataResults []MetadataResult

// MetadataResult mimics the structure of SearchResult and is the response
// to the wpt.fyi result page.
type MetadataResult struct {
	// Test is the name of a test; this often corresponds to a test file path in
	// the WPT source reposiory.
	Test string `json:"test"`
	// URLs represents a list of bug urls that are associated with
	// this test.
	URLs []string `json:"urls,omitempty"`
}

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
	Product ProductSpec          `yaml:"product"`
	URL     string               `yaml:"url"`
	Results []MetadataTestResult `yaml:"results"`
}

// MetadataTestResult is a filter for test results to which the metadata link
// should apply.
type MetadataTestResult struct {
	TestPath    string     `yaml:"test"`
	SubtestName string     `yaml:"subtest"`
	Status      TestStatus `yaml:"status"`
}

func (m MetadataResults) Len() int           { return len(m) }
func (m MetadataResults) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m MetadataResults) Less(i, j int) bool { return m[i].Test < m[j].Test }

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
	res := MetadataResults{}
	for folderPath, data := range metadata {
		testMap := make(map[string][]string)

		for _, link := range data.Links {
			var urls []string

			for _, result := range link.Results {
				//TODO(kyleju): Concatenate test path on WPT Metadata repository instead of here.
				var fullTestName = "/" + folderPath + "/" + result.TestPath

				if _, ok := testMap[fullTestName]; !ok {
					testMap[fullTestName] = make([]string, len(productSpecs))
				}
				urls = testMap[fullTestName]

				for i, productSpec := range productSpecs {
					// Matches browser type if a version is not specified.
					if link.Product.MatchesProductSpec(productSpec) {
						urls[i] = link.URL
					} else if link.Product.BrowserName == "" && urls[i] == "" {
						// Matches to all browsers if product is not specified.
						urls[i] = link.URL
					}
				}
			}
		}
		for nameKey, urlsVal := range testMap {
			isMatches := false

			for _, url := range urlsVal {
				if url != "" {
					isMatches = true
				}
			}

			// No matching testRuns.
			if !isMatches {
				continue
			}

			linkResult := MetadataResult{Test: nameKey, URLs: urlsVal}
			res = append(res, linkResult)
		}
	}
	sort.Sort(res)
	return res
}

// PrepareLinkFilter maps a MetadataResult test name to its URLs.
func PrepareLinkFilter(metadata MetadataResults) map[string][]string {
	metadataMap := make(map[string][]string)
	for _, data := range metadata {
		metadataMap[data.Test] = data.URLs
	}
	return metadataMap
}
