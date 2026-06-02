// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -build_flags=--mod=mod -destination sharedtest/fetch_bsf_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared FetchBSF

package shared

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"
)

const (
	// experimentalBSFURL is the GitHub URL for fetching the experimental BSF data
	// for Chrome, Firefox and Safari. First party tests only.
	experimentalBSFURL = "https://raw.githubusercontent.com/web-platform-tests/results-analysis/gh-pages/data/experimental-browser-specific-failures.csv"
	// stableBSFURL is the GitHub URL for fetching the stable BSF data
	// for Chrome, Firefox and Safari. First party tests only.
	stableBSFURL = "https://raw.githubusercontent.com/web-platform-tests/results-analysis/gh-pages/data/stable-browser-specific-failures.csv"
	// experimentalBSFWithThirdPartyURL is the GitHub URL for fetching the experimental BSF data
	// for Chrome, Firefox and Safari. All tests (first and third party) included.
	experimentalBSFWithThirdPartyURL = "https://raw.githubusercontent.com/web-platform-tests/results-analysis/gh-pages/data/experimental-browser-specific-failures-with-third-party.csv"
	// stableBSFWithThirdPartyURL is the GitHub URL for fetching the stable BSF data
	// for Chrome, Firefox and Safari. All tests (first and third party) included.
	stableBSFWithThirdPartyURL = "https://raw.githubusercontent.com/web-platform-tests/results-analysis/gh-pages/data/stable-browser-specific-failures-with-third-party.csv"
)

// BSFData stores BSF data of the latest WPT revision.
type BSFData struct {
	// The latest WPT Revision updated in this BSF data.
	LastUpdateRevision string `json:"lastUpdateRevision"`
	// Fields correspond to the fields (columns) in this BSF data table.
	Fields []string `json:"fields"`
	// BSF data table, defined by the fields.
	Data [][]string `json:"data"`
}

// FilterandExtractBSFData filters rawBSFdata based on query filters `from` and `to`,
// and generates BSFData for bsf_handler. rawBSFdata is [][]string with the 0th index
// as fields and the rest as the BSF data table in chronological order; e.g.
// [[sha,date,chrome-version,chrome,firefox-version,firefox,safari-version,safari],
// [eea0b54014e970a2f94f1c35ec6e18ece76beb76,2018-08-07,70.0.3510.0 dev,602.0505256721168,63.0a1,1617.1788882804883,12.1,2900.3438625831423],
// [203c34855f6871d6e55eaf7b55b50dad563f781f,2018-08-18,70.0.3521.2 dev,605.3869030161061,63.0a1,1521.908686731921,12.1,2966.686195133767],
// ...]
func FilterandExtractBSFData(rawBSFdata [][]string, from *time.Time, to *time.Time) BSFData {
	if len(rawBSFdata) == 0 {
		return BSFData{}
	}

	var response BSFData
	response.Fields = rawBSFdata[0]

	var dateIndex int
	for i, field := range response.Fields {
		if field == "date" {
			dateIndex = i
			break
		}
	}

	var data [][]string
	for i, row := range rawBSFdata {
		// The 0 row is fields.
		if i == 0 {
			continue
		}

		updated, e := time.Parse("2006-01-02", row[dateIndex])
		if e != nil {
			continue
		}

		// from is inclusive.
		if from != nil && updated.Before(*from) {
			continue
		}

		// to is exclusive.
		if to != nil && (updated.After(*to) || updated.Equal(*to)) {
			continue
		}

		data = append(data, row)
	}

	if len(data) == 0 {
		return BSFData{}
	}

	// The lateset revision should be the last row at the 0th index.
	response.LastUpdateRevision = data[len(data)-1][0]
	response.Data = data
	return response
}

// FetchBSF encapsulates the Fetch(isExperimental, includeThirdParty bool) method for testing.
type FetchBSF interface {
	Fetch(isExperimental, includeThirdParty bool) ([][]string, error)
}

type fetchBSF struct{}

// Fetch() fetches BSF Data in CSV from GitHub given query options, in chronological order.
func (f fetchBSF) Fetch(isExperimental, includeThirdParty bool) ([][]string, error) {
	url := ""
	if isExperimental {
		if includeThirdParty {
			url = experimentalBSFWithThirdPartyURL
		} else {
			url = experimentalBSFURL
		}
	} else {
		if includeThirdParty {
			url = stableBSFWithThirdPartyURL
		} else {
			url = stableBSFURL
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("BSF fetch from %q failed: %w", url, err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Non-OK HTTP status code of %d from "%s"`, resp.StatusCode, url)
	}

	data, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("BSF CSV parse from %q failed: %w", url, err)
	}

	return data, nil
}

// NewFetchBSF returns an instance of FetchBSF for apiBSFHandler.
func NewFetchBSF() FetchBSF {
	return fetchBSF{}
}
