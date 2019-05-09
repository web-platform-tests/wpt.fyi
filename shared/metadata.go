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
const MetadataArchiveURL = "https://api.github.com/repos/web-platform-tests/wpt-metadata/tarball"

// MetadataTestingURL is the URL for testing.
var MetadataTestingURL = ""

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
	Product  ProductSpec
	TestPath string `yaml:"test"`
	URL      string
}

func (m MetadataResults) Len() int           { return len(m) }
func (m MetadataResults) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m MetadataResults) Less(i, j int) bool { return m[i].Test < m[j].Test }

// GetMetadatURL returns the URL that fetches Metadata information.
// For testing, it returns a MetadataTestingURL set by users.
func GetMetadatURL(isTesting bool) string {
	if isTesting {
		return MetadataTestingURL
	}
	return MetadataArchiveURL
}

// GetMetadataResponse retrieves the response to a WPT Metadata query.
func GetMetadataResponse(testRuns []TestRun, client *http.Client, log Logger, url string) (MetadataResults, error) {
	var productAtRevisions = make([]ProductAtRevision, len(testRuns))
	for i, run := range testRuns {
		productAtRevisions[i] = run.ProductAtRevision
	}
	return getMetadataResponseOnProductRevisions(productAtRevisions, client, log, url)
}

// GetMetadataResponseOnProducts constructs the response to a WPT Metadata query, given ProductSpecs.
func GetMetadataResponseOnProducts(productSpecs ProductSpecs, client *http.Client, log Logger, url string) (MetadataResults, error) {
	var productAtRevisions = make([]ProductAtRevision, len(productSpecs))
	for i, productSpec := range productSpecs {
		productAtRevisions[i] = productSpec.ProductAtRevision
	}
	return getMetadataResponseOnProductRevisions(productAtRevisions, client, log, url)
}

func getMetadataResponseOnProductRevisions(productAtRevisions []ProductAtRevision, client *http.Client, log Logger, url string) (MetadataResults, error) {
	metadataByteMap, err := util.CollectMetadata(client, url)
	if err != nil {
		return nil, err
	}
	metadata := parseMetadata(metadataByteMap, log)
	return constructMetadataResponse(productAtRevisions, metadata), nil
}

// parseMetadata collects and parses all META.yml files from
// wpt-metadata reposiroty.
func parseMetadata(metadataByteMap map[string][]byte, log Logger) map[string]Metadata {
	var metadataMap = make(map[string]Metadata)
	for path, data := range metadataByteMap {
		var metadata Metadata
		err := yaml.Unmarshal(data, &metadata)
		if err != nil {
			log.Warningf("Failed to unmarshal %s.yml.", path)
			continue
		}
		metadataMap[path] = metadata
	}
	return metadataMap
}

// constructMetadataResponse constructs the response to a WPT Metadata query, given ProductSpecs.
func constructMetadataResponse(productAtRevisions []ProductAtRevision, metadata map[string]Metadata) MetadataResults {
	res := MetadataResults{}
	for folderPath, data := range metadata {
		testMap := make(map[string][]string)

		for _, link := range data.Links {
			var urls []string

			//TODO: Concatenate test path on WPT Metadata repository instead of here.
			var fullTestName = "/" + folderPath + "/" + link.TestPath

			if _, ok := testMap[fullTestName]; !ok {
				testMap[fullTestName] = make([]string, len(productAtRevisions))
			}
			urls = testMap[fullTestName]

			for i, productAtRevision := range productAtRevisions {
				// Matches browser type if a version is not specified.
				if link.Product.MatchesProductAtRevision(productAtRevision) {
					urls[i] = link.URL
				} else if link.Product.BrowserName == "" && urls[i] == "" {
					// Matches to all browsers if product is not specified.
					urls[i] = link.URL
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
