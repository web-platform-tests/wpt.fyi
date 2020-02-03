// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// MetadataMapCached is the local cache of Metadata in searchcache.
var MetadataMapCached map[string][]byte = nil

type searchcacheMetadataFetcher struct {
	client *http.Client
	url    string
}

func (f searchcacheMetadataFetcher) Fetch() (res map[string][]byte, err error) {
	if MetadataMapCached != nil {
		return MetadataMapCached, nil
	}

	return shared.CollectMetadataWithURL(f.client, f.url)
}
