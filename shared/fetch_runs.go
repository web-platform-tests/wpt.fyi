// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// FetchLatestRuns fetches the TestRun metadata for the latest runs, using the
// API on the given host.
func FetchLatestRuns(wptdHost string) TestRuns {
	return FetchRuns(wptdHost, TestRunFilter{})
}

// FetchRuns fetches the TestRun metadata for the given sha / labels, using the
// API on the given host.
func FetchRuns(wptdHost string, filter TestRunFilter) TestRuns {
	url := "https://" + wptdHost + "/api/runs"
	url += "?" + filter.ToQuery(true).Encode()

	var runs TestRuns
	FetchAPIJSON(url, &runs)
	return runs
}

// FetchAPIJSON fetches the given URL and unmarshals it into the given
// value pointer, fatally logging any errors.
func FetchAPIJSON(url string, value interface{}) {
	log.Printf("Fetching %s...", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal(errors.New("Bad response code from " + url + ": " +
			strconv.Itoa(resp.StatusCode)))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(body, value); err != nil {
		log.Fatal(err)
	}
}
