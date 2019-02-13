// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package client is a package for simplifying the upload request made by a
// client to the results receiver upload endpoint (/api/results/upload).
package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// UploadTimeout is the timeout to upload results to the results receiver.
const UploadTimeout = time.Minute

// Client is the interface for the client.
type Client interface {
	CreateRun(
		sha,
		username string,
		password string,
		reportURLs []string,
		labels []string) error
}

// NewClient returns a client impl.
func NewClient(aeAPI shared.AppEngineAPI) Client {
	return client{
		aeAPI: aeAPI,
	}
}

type client struct {
	aeAPI shared.AppEngineAPI
}

// CreateRun takes the given requirements and issues a POST request to collect the
// given reportURLs
func (c client) CreateRun(
	sha,
	username,
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
	host := c.aeAPI.GetVersionedHostname()
	payload.Add("callback_url", fmt.Sprintf("https://%s/api/results/create", host))

	uploadURL := c.aeAPI.GetResultsUploadURL()
	req, err := http.NewRequest("POST", uploadURL.String(), strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)

	slowClient, cancel := c.aeAPI.GetSlowHTTPClient(UploadTimeout)
	defer cancel()
	resp, err := slowClient.Do(req)
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
