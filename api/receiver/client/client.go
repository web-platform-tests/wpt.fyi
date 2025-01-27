// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package client is a package for simplifying the upload request made by a
// client to the results receiver upload endpoint (/api/results/upload).
package client

import (
	"context"
	"fmt"
	"io"
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
		resultURLs []string,
		screenshotURLs []string,
		archiveURLs []string,
		labels []string) error
}

// NewClient returns a client impl.
// nolint:ireturn // TODO: Fix ireturn lint error
func NewClient(aeAPI shared.AppEngineAPI) Client {
	return client{
		aeAPI: aeAPI,
	}
}

type client struct {
	aeAPI shared.AppEngineAPI
}

// CreateRun issues a POST request to the results receiver with the given payload.
func (c client) CreateRun(
	sha,
	username,
	password string,
	resultURLs []string,
	screenshotURLs []string,
	archiveURLs []string,
	labels []string) error {
	// https://github.com/web-platform-tests/wpt.fyi/blob/main/api/README.md#url-payload
	payload := make(url.Values)
	// Not to be confused with `revision` in the wpt.fyi TestRun model, this
	// parameter is the full revision hash.
	if sha != "" {
		payload.Add("revision", sha)
	}
	for _, url := range resultURLs {
		payload.Add("result_url", url)
	}
	for _, url := range screenshotURLs {
		payload.Add("screenshot_url", url)
	}
	for _, url := range archiveURLs {
		payload.Add("archive_url", url)
	}
	if labels != nil {
		payload.Add("labels", strings.Join(labels, ","))
	}
	// Ensure we call back to this appengine version instance.
	host := c.aeAPI.GetVersionedHostname()
	payload.Add("callback_url", fmt.Sprintf("https://%s/api/results/create", host))

	uploadURL := c.aeAPI.GetResultsUploadURL()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		uploadURL.String(),
		strings.NewReader(payload.Encode()),
	)
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	hc := c.aeAPI.GetHTTPClientWithTimeout(UploadTimeout)
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("API error: HTTP %v: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
