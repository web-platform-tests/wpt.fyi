// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package backfill

import (
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
)

func TestNilIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	fetcher := NewMockRunFetcher(ctrl)
	_, err := FillIndex(fetcher, nil, nil, 1, uint64(1), nil)
	assert.Equal(t, errNilIndex, err)
}

func TestFetchErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	fetcher := NewMockRunFetcher(ctrl)
	idx := index.NewMockIndex(ctrl)
	expected := errors.New("Fetch error")
	fetcher.EXPECT().FetchRuns(gomock.Any()).Return(nil, expected)
	_, err := FillIndex(fetcher, nil, nil, 1, uint64(1), idx)
	assert.Equal(t, expected, err)
}
