// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/go-yaml/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/web-platform-tests/wpt-metadata/util"
)

// MetadataResponse is a response to a wpt-metadata query.
type MetadataResponse struct {
	Response MetadataResults
}

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
func (m MetadataResults) Less(i, j int) bool { return m[i].String() < m[j].String() }

func (m MetadataResult) String() string {
	var urlString string
	for _, url := range m.URLs {
		urlString += url
	}
	return m.Test + urlString
}

// GetMetadataResponse retrieves the response to a WPT Metadata query.
func GetMetadataResponse(testRuns []TestRun, client *http.Client) MetadataResponse {
	metadataByteMap, err := util.CollectMetadata(client)
	if err != nil {
		return MetadataResponse{}
	}
	metadata := parseMetadata(metadataByteMap)
	return constructMetadataResponse(testRuns, metadata)
}

// parseMetadata collects and parses all META.yml files from
// wpt-metadata reposiroty.
func parseMetadata(metadataByteMap map[string][]byte) map[string]Metadata {
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

// ConstructMetadataResponse constructs the response to a WPT Metadata query.
// When parsing 'link' nodes, assume there is no mising information nor duplicates;
// assume each test for each browser type is only associated with one bug.
func constructMetadataResponse(testRuns []TestRun, metadata map[string]Metadata) MetadataResponse {
	res := MetadataResults{}
	for folderPath, data := range metadata {
		testMap := make(map[string][]string)

		for _, link := range data.Links {
			var urls []string

			var fullTestName = ""
			if strings.HasPrefix(link.TestPath, folderPath) {
				fullTestName = link.TestPath
			} else {
				fullTestName = path.Join(folderPath, link.TestPath)
			}

			if _, ok := testMap[fullTestName]; !ok {
				testMap[fullTestName] = make([]string, len(testRuns))
			}
			urls = testMap[fullTestName]

			for i, run := range testRuns {
				// Only matches browser type.
				if link.Product.BrowserMatches(run) {
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
	return MetadataResponse{Response: res}

}
