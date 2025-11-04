//go:build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package backfill

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/query/cache/index"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	gomock "go.uber.org/mock/gomock"
)

func TestNilIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := sharedtest.NewMockDatastore(ctrl)
	_, err := FillIndex(store, nil, nil, 1, uint(10), uint64(1), 0.0, nil)
	assert.Equal(t, errNilIndex, err)
}

func TestFetchErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := sharedtest.NewMockDatastore(ctrl)
	query := sharedtest.NewMockTestRunQuery(ctrl)
	idx := index.NewMockIndex(ctrl)
	expected := errors.New("fetch error")
	store.EXPECT().TestRunQuery().Return(query)
	query.EXPECT().LoadTestRuns(gomock.Any(), nil, nil, nil, nil, gomock.Any(), nil).Return(nil, expected)
	_, err := FillIndex(store, nil, nil, 1, uint(10), uint64(1), 0.0, idx)
	assert.Equal(t, expected, err)
}
