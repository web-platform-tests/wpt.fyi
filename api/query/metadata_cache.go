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
	// TODO(kyleju): plumb the SHA of a returned wpt-metadata to solve
	// the race condition; see https://github.com/web-platform-tests/wpt.fyi/issues/1890.
	res, err = shared.CollectMetadataWithURL(netClient, f.url, nil)
	return nil, res, err
}
