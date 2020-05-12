// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// MetadataMapCached is the local cache of Metadata in searchcache.
var MetadataMapCached map[string][]byte = nil

type searchcacheMetadataFetcher struct {
	url string
}

func (f searchcacheMetadataFetcher) Fetch() (sha *string, res map[string][]byte, err error) {
	if MetadataMapCached != nil {
		return nil, MetadataMapCached, nil
	}

	var netClient = &http.Client{
		Timeout: time.Second * 5,
	}
	// TODO(kyleju): utilize the SHA information here to potentially resolve the metadata cache
	// out-of-sync issue between searchcache and webapp.
	res, err = shared.CollectMetadataWithURL(netClient, f.url, nil)
	return nil, res, err
}
