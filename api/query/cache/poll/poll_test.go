//go:build small

// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package poll

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

type testWebFeaturesGetter struct {
	data shared.WebFeaturesData
	err  error
}

func (g testWebFeaturesGetter) Get(_ context.Context) (shared.WebFeaturesData, error) {
	return g.data, g.err
}

func TestKeepWebFeaturesManifestUpdated(t *testing.T) {
	testCases := []struct {
		name         string
		initialCache shared.WebFeaturesData
		dataToReturn shared.WebFeaturesData
		errToReturn  error
		expectedData shared.WebFeaturesData
	}{
		{
			name:         "initial successful update",
			initialCache: nil,
			dataToReturn: shared.WebFeaturesData{
				"foo": {"/foo.js": nil},
			},
			errToReturn: nil,
			expectedData: shared.WebFeaturesData{
				"foo": {"/foo.js": nil},
			},
		},
		{
			name:         "initial failed update",
			initialCache: nil,
			dataToReturn: nil,
			errToReturn:  errors.New("whoops"),
			expectedData: nil,
		},
		{
			name: "successful update to existing cache",
			initialCache: shared.WebFeaturesData{
				"foo": {"/foo.js": nil},
			},
			dataToReturn: shared.WebFeaturesData{
				"bar": {"/bar.js": nil},
			},
			errToReturn: nil,
			expectedData: shared.WebFeaturesData{
				"bar": {"/bar.js": nil},
			},
		},
		{
			name: "unsuccessful update to existing cache. return existing",
			initialCache: shared.WebFeaturesData{
				"foo": {"/foo.js": nil},
			},
			dataToReturn: nil,
			errToReturn:  errors.New("whoops"),
			expectedData: shared.WebFeaturesData{
				"foo": {"/foo.js": nil},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the cache.
			query.SetWebFeaturesDataCache(tc.initialCache)
			require.Equal(t, tc.initialCache, query.GetWebFeaturesDataCache())

			getter := testWebFeaturesGetter{tc.dataToReturn, tc.errToReturn}
			ctx := context.Background()
			logger := shared.GetLogger(ctx)
			keepWebFeaturesManifestUpdated(ctx, logger, getter)

			assert.Equal(t, tc.expectedData, query.GetWebFeaturesDataCache())
		})
	}
}
