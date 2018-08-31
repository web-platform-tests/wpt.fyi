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
func FetchLatestRuns(wptdHost string) (TestRuns, error) {
	return FetchRuns(wptdHost, TestRunFilter{})
}

// FetchRuns fetches the TestRun metadata for the given sha / labels, using the
// API on the given host.
func FetchRuns(wptdHost string, filter TestRunFilter) (TestRuns, error) {
	url := "https://" + wptdHost + "/api/runs"
	url += "?" + filter.ToQuery(true).Encode()

	var runs TestRuns
	err := FetchJSON(url, &runs)
	return runs, err
}

// FetchJSON fetches the given URL, which is expected to be JSON, and unmarshals
// it into the given value pointer, fatally logging any errors.
func FetchJSON(url string, value interface{}) error {
	log.Printf("Fetching %s...", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("Bad response code from " + url + ": " +
			strconv.Itoa(resp.StatusCode))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, value)
}
