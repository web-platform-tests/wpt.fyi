//go:build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api //nolint:revive

import (
	"bytes"
	"compress/gzip"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/manifest/mock_manifest"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestGetGitHubReleaseAsset_Caches(t *testing.T) {
	ctx := sharedtest.NewTestContext()
	log := shared.GetLogger(ctx)
	errNotFound := errors.New("not found")

	fullSHA := strings.Repeat("1234567890", 4)
	content := "latest data"
	data := getManifestPayload(content)

	t.Run("cache missed", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockMC := sharedtest.NewMockReadWritable(mockCtrl)
		mockWC := sharedtest.NewMockWriteCloser(t)
		mockLatestMC := sharedtest.NewMockReadWritable(mockCtrl)
		mockLatestWC := sharedtest.NewMockWriteCloser(t)
		manifestAPI := mock_manifest.NewMockAPI(mockCtrl)
		manifestAPI.EXPECT().NewRedis(time.Hour * 48).Return(mockMC)
		manifestAPI.EXPECT().NewRedis(time.Minute * 5).Return(mockLatestMC)
		gomock.InOrder(
			mockLatestMC.EXPECT().NewReadCloser("MANIFEST-latest").Return(nil, errNotFound),
			manifestAPI.EXPECT().GetManifestForSHA("").Return(fullSHA, data, nil),
			mockMC.EXPECT().NewWriteCloser("MANIFEST-"+fullSHA).Return(mockWC, nil),
			mockLatestMC.EXPECT().NewWriteCloser("MANIFEST-latest").Return(mockLatestWC, nil),
		)

		sha, latestManifest, err := getManifest(log, manifestAPI, "", nil)
		assert.Nil(t, err)
		assert.Equal(t, fullSHA, sha)
		assert.Equal(t, content, string(latestManifest))
		// Should be added to cache
		assert.True(t, mockLatestWC.IsClosed())
		assert.Equal(t, fullSHA, string(mockLatestWC.Bytes()))
		assert.True(t, mockWC.IsClosed())
		assert.Equal(t, data, mockWC.Bytes())
	})

	t.Run("cache hit (latest)", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockMC := sharedtest.NewMockReadWritable(mockCtrl)
		mockRC := sharedtest.NewMockReadCloser(t, []byte(data))
		mockLatestMC := sharedtest.NewMockReadWritable(mockCtrl)
		mockLatestRC := sharedtest.NewMockReadCloser(t, []byte(fullSHA))
		manifestAPI := mock_manifest.NewMockAPI(mockCtrl)
		manifestAPI.EXPECT().NewRedis(time.Hour * 48).AnyTimes().Return(mockMC)
		manifestAPI.EXPECT().NewRedis(time.Minute * 5).AnyTimes().Return(mockLatestMC)
		gomock.InOrder(
			mockLatestMC.EXPECT().NewReadCloser("MANIFEST-latest").Return(mockLatestRC, nil),
			mockMC.EXPECT().NewReadCloser("MANIFEST-"+fullSHA).Return(mockRC, nil),
		)

		// Load from cache without touching API.
		sha, latestManifest, err := getManifest(log, manifestAPI, "latest", nil)
		assert.Nil(t, err)
		assert.Equal(t, fullSHA, sha)
		assert.Equal(t, content, string(latestManifest))
	})

	t.Run("cache hit (full SHA)", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockMC := sharedtest.NewMockReadWritable(mockCtrl)
		mockRC := sharedtest.NewMockReadCloser(t, []byte(data))
		mockLatestMC := sharedtest.NewMockReadWritable(mockCtrl)
		manifestAPI := mock_manifest.NewMockAPI(mockCtrl)
		manifestAPI.EXPECT().NewRedis(time.Hour * 48).Return(mockMC)
		manifestAPI.EXPECT().NewRedis(time.Minute * 5).Return(mockLatestMC)
		mockMC.EXPECT().NewReadCloser("MANIFEST-"+fullSHA).Return(mockRC, nil)

		// Load from cache without touching API.
		sha, latestManifest, err := getManifest(log, manifestAPI, fullSHA, nil)
		assert.Nil(t, err)
		assert.Equal(t, fullSHA, sha)
		assert.Equal(t, content, string(latestManifest))
	})
}

func getManifestPayload(data string) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write([]byte(data))
	zw.Close()
	return buf.Bytes()
}
