// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package push

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/revisions/announcer"
	"github.com/web-platform-tests/wpt.fyi/revisions/api"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
	agit "github.com/web-platform-tests/wpt.fyi/revisions/git"
)

func TestGetLatestRevisions_FailedGetRevisions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	epochs := []epoch.Epoch{epoch.Hourly{}}
	latestInput := make(map[epoch.Epoch]int)
	err := errors.New("GetRevisions error")
	a.EXPECT().GetLatestGetRevisionsInput().Return(latestInput)
	ancr.EXPECT().GetRevisions(latestInput, gomock.Any()).Return(nil, err)

	_, glrErr := GetLatestRevisions(a, ancr, epochs)
	assert.Equal(t, err, glrErr)
}

func TestGetLatestRevisions_FailedLatestFromEpochs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	epochs := []epoch.Epoch{epoch.Hourly{}}
	latestInput := make(map[epoch.Epoch]int)

	// Empty list filed under epoch triggers error in LatestFromEpochs.
	//
	// TODO(markdittmer): Perhaps functions in revisions/api/types.go should be
	// wrapped in an interface or struct so that they can be mocked.
	revs := map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}: []agit.Revision{},
	}

	a.EXPECT().GetLatestGetRevisionsInput().Return(latestInput)

	ancr.EXPECT().GetRevisions(latestInput, gomock.Any()).Return(revs, nil)

	_, err := GetLatestRevisions(a, ancr, epochs)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(strings.ToLower(err.Error()), "missing"))
}

func TestDiffLatest(t *testing.T) {
	t0 := time.Now()
	t1 := api.UTCTime(t0)
	t2 := api.UTCTime(t0.Add(time.Hour))
	t3 := api.UTCTime(t0.Add(time.Hour * 2))
	r1 := api.Revision{"01", t1}
	r2 := api.Revision{"02", t2}
	r3 := api.Revision{"03", t3}
	prev := map[string]api.Revision{
		api.FromEpoch(epoch.Weekly{}).ID: r1,
		api.FromEpoch(epoch.Daily{}).ID:  r2,
	}
	next := map[string]api.Revision{
		api.FromEpoch(epoch.Daily{}).ID:  r3,
		api.FromEpoch(epoch.Hourly{}).ID: r3,
	}
	epochs := []epoch.Epoch{
		epoch.Weekly{},
		epoch.Daily{},
		epoch.Hourly{},
	}
	expected := []api.Diff{
		api.Diff{api.FromEpoch(epoch.Weekly{}).ID, &r1, nil},
		api.Diff{api.FromEpoch(epoch.Daily{}).ID, &r2, &r3},
		api.Diff{api.FromEpoch(epoch.Hourly{}).ID, nil, &r3},
	}

	actual := DiffLatest(prev, next, epochs)
	assert.Equal(t, len(expected), len(actual))
	for i := range expected {
		assert.Equal(t, expected[i].Epoch, actual[i].Epoch)
		if expected[i].Prev == nil {
			assert.Nil(t, actual[i].Prev)
		} else {
			assert.Equal(t, expected[i].Prev, actual[i].Prev)
		}
		if expected[i].Next == nil {
			assert.Nil(t, actual[i].Next)
		} else {
			assert.Equal(t, expected[i].Next, actual[i].Next)
		}
	}
}
