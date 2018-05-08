// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	models "github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

type gitHubClientImpl struct {
	Token   *models.Token
	Context context.Context
}

func (g *gitHubClientImpl) fetch(url string) ([]byte, error) {
	client := urlfetch.Client(g.Context)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != nil && g.Token.Secret != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", g.Token.Secret))
	}
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s returned HTTP status %d:\n%s", url, resp.StatusCode, string(body))
	}
	return body, nil
}
