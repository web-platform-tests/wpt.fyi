// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"bytes"
	"compress/gzip"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/manifest"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine/memcache"
)

func TestGetGitHubReleaseAsset_Caches(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	fullSHA := strings.Repeat("1234567890", 4)
	content := "latest data"
	data := getManifestPayload(content)

	manifestAPI := manifest.NewMockAPI(mockCtrl)
	manifestAPI.EXPECT().GetManifestForSHA("latest").Return(fullSHA, data, nil)

	// Should be added to cache
	_, latestManifest, _ := getManifest(ctx, manifestAPI, "", nil)
	assert.Equal(t, []byte(content), latestManifest)
	cached, _ := memcache.Get(ctx, manifestCacheKey("latest"))
	assert.Equal(t, fullSHA, string(cached.Value))
	cached, _ = memcache.Get(ctx, manifestCacheKey(fullSHA))
	assert.Equal(t, data, cached.Value)

	// Second time, they're loaded from cache, without touching API.
	_, latestManifest, _ = getManifest(ctx, manifestAPI, "latest", nil)
	assert.Equal(t, []byte(content), latestManifest)
	_, latestManifest, _ = getManifest(ctx, manifestAPI, fullSHA, nil)
	assert.Equal(t, []byte(content), latestManifest)
}

func getManifestPayload(data string) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write([]byte(data))
	zw.Close()
	return buf.Bytes()
}
