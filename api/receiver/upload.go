// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
)

// CreateRun takes the given requirements and issues a POST request to collect the
// given reportURLs
func CreateRun(
	client *http.Client,
	aeAPI shared.AppEngineAPI,
	sha,
	username string,
	password string,
	reportURLs []string,
	labels []string) error {
	// https://github.com/web-platform-tests/wpt.fyi/blob/master/api/README.md#url-payload
	payload := make(url.Values)
	// Not to be confused with `revision` in the wpt.fyi TestRun model, this
	// parameter is the full revision hash.
	if sha != "" {
		payload.Add("revision", sha)
	}
	for _, url := range reportURLs {
		payload.Add("result_url", url)
	}
	if labels != nil {
		payload.Add("labels", strings.Join(labels, ","))
	}
	// Ensure we call back to this appengine version instance.
	host := aeAPI.GetVersionedHostname()
	payload.Add("callback_url", fmt.Sprintf("https://%s/api/results/create", host))

	// https://github.com/web-platform-tests/wpt.fyi/blob/master/api/README.md#results-creation
	uploadURL := fmt.Sprintf("https://%s/api/results/upload", appengine.DefaultVersionHostname(aeAPI.Context()))
	req, err := http.NewRequest("POST", uploadURL, strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("API error: HTTP %v: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
