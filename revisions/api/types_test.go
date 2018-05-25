// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
	agit "github.com/web-platform-tests/wpt.fyi/revisions/git"
	"github.com/web-platform-tests/wpt.fyi/revisions/test"
)

var today = time.Now()
var yesterday = today.Add(-26 * time.Hour)

func TestFromEpoch(t *testing.T) {
	epochIn := epoch.Daily{}
	dataIn := epochIn.GetData()

	epochOut := api.FromEpoch(epochIn)

	assert.Equal(t, dataIn.Description, epochOut.Description)
	assert.Equal(t, dataIn.Label, epochOut.Label)
	assert.Equal(t, float32(dataIn.MinDuration.Seconds()), epochOut.MinDuration)
	assert.Equal(t, float32(dataIn.MaxDuration.Seconds()), epochOut.MaxDuration)
}

func TestFromEpoch_ID(t *testing.T) {
	epochs := []epoch.Epoch{
		epoch.Weekly{},
		epoch.Daily{},
		epoch.EightHourly{},
		epoch.FourHourly{},
		epoch.TwoHourly{},
		epoch.Hourly{},
	}

	for i, e1 := range epochs {
		for j, e2 := range epochs {
			if i == j {
				continue
			}

			assert.NotEqual(t, api.FromEpoch(e1).ID, api.FromEpoch(e2).ID)
		}
	}
}

func TestLatestFromEpochs_missing(t *testing.T) {
	a := agit.RevisionData{
		Hash:       test.NewHash("0a"),
		CommitTime: today,
	}
	revs := map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}:    []agit.Revision{a},
		epoch.TwoHourly{}: []agit.Revision{},
	}

	_, err := api.LatestFromEpochs(revs)
	assert.Equal(t, api.GetErMissingRevision(), err)
}

func TestLatestFromEpochs(t *testing.T) {
	a := agit.RevisionData{
		Hash:       test.NewHash("0a"),
		CommitTime: today,
	}
	b := agit.RevisionData{
		Hash:       test.NewHash("0b"),
		CommitTime: yesterday,
	}
	c := agit.RevisionData{
		Hash:       test.NewHash("0c"),
		CommitTime: today,
	}
	d := agit.RevisionData{
		Hash:       test.NewHash("0d"),
		CommitTime: yesterday,
	}
	e := agit.RevisionData{
		Hash:       test.NewHash("0e"),
		CommitTime: yesterday,
	}
	f := agit.RevisionData{
		Hash:       test.NewHash("0f"),
		CommitTime: today,
	}
	revs := map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}:      []agit.Revision{a, b},
		epoch.FourHourly{}:  []agit.Revision{d, e},
		epoch.EightHourly{}: []agit.Revision{f},
		epoch.TwoHourly{}:   []agit.Revision{c},
	}
	expected := api.LatestResponse{
		Revisions: map[string]api.Revision{
			api.FromEpoch(epoch.Hourly{}).ID:      api.FromRevision(a),
			api.FromEpoch(epoch.TwoHourly{}).ID:   api.FromRevision(c),
			api.FromEpoch(epoch.FourHourly{}).ID:  api.FromRevision(d),
			api.FromEpoch(epoch.EightHourly{}).ID: api.FromRevision(f),
		},
		Epochs: []api.Epoch{
			api.FromEpoch(epoch.Hourly{}),
			api.FromEpoch(epoch.TwoHourly{}),
			api.FromEpoch(epoch.FourHourly{}),
			api.FromEpoch(epoch.EightHourly{}),
		},
	}

	actual, err := api.LatestFromEpochs(revs)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestRevisionsFromEpochs_err(t *testing.T) {
	err := errors.New("Announcer error")
	response := api.RevisionsFromEpochs(map[epoch.Epoch][]agit.Revision{}, err)
	assert.Equal(t, err.Error(), response.Error)
}

func TestRevisionsFromEpochs_several(t *testing.T) {
	a := agit.RevisionData{
		Hash:       test.NewHash("0a"),
		CommitTime: today,
	}
	b := agit.RevisionData{
		Hash:       test.NewHash("0b"),
		CommitTime: yesterday,
	}
	c := agit.RevisionData{
		Hash:       test.NewHash("0c"),
		CommitTime: today,
	}
	d := agit.RevisionData{
		Hash:       test.NewHash("0d"),
		CommitTime: yesterday,
	}
	e := agit.RevisionData{
		Hash:       test.NewHash("0e"),
		CommitTime: yesterday,
	}
	f := agit.RevisionData{
		Hash:       test.NewHash("0f"),
		CommitTime: today,
	}
	revs := map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}:      []agit.Revision{a, b},
		epoch.FourHourly{}:  []agit.Revision{d, e},
		epoch.EightHourly{}: []agit.Revision{f},
		epoch.TwoHourly{}:   []agit.Revision{c},
	}

	expected := api.RevisionsResponse{
		Revisions: map[string][]api.Revision{
			api.FromEpoch(epoch.Hourly{}).ID: []api.Revision{
				api.FromRevision(a),
				api.FromRevision(b),
			},
			api.FromEpoch(epoch.TwoHourly{}).ID: []api.Revision{
				api.FromRevision(c),
			},
			api.FromEpoch(epoch.FourHourly{}).ID: []api.Revision{
				api.FromRevision(d),
				api.FromRevision(e),
			},
			api.FromEpoch(epoch.EightHourly{}).ID: []api.Revision{
				api.FromRevision(f),
			},
		},
		Epochs: []api.Epoch{
			api.FromEpoch(epoch.Hourly{}),
			api.FromEpoch(epoch.TwoHourly{}),
			api.FromEpoch(epoch.FourHourly{}),
			api.FromEpoch(epoch.EightHourly{}),
		},
	}

	actual := api.RevisionsFromEpochs(revs, nil)

	assert.Equal(t, expected, actual)
}

func TestRevisionsFromEpochs_several_err(t *testing.T) {
	a := agit.RevisionData{
		Hash:       test.NewHash("0a"),
		CommitTime: today,
	}
	b := agit.RevisionData{
		Hash:       test.NewHash("0b"),
		CommitTime: yesterday,
	}
	c := agit.RevisionData{
		Hash:       test.NewHash("0c"),
		CommitTime: today,
	}
	d := agit.RevisionData{
		Hash:       test.NewHash("0d"),
		CommitTime: yesterday,
	}
	e := agit.RevisionData{
		Hash:       test.NewHash("0e"),
		CommitTime: yesterday,
	}
	f := agit.RevisionData{
		Hash:       test.NewHash("0f"),
		CommitTime: today,
	}
	revs := map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}:      []agit.Revision{a, b},
		epoch.FourHourly{}:  []agit.Revision{d, e},
		epoch.EightHourly{}: []agit.Revision{f},
		epoch.TwoHourly{}:   []agit.Revision{c},
	}
	err := errors.New("Announcer error")

	expected := api.RevisionsResponse{
		Revisions: map[string][]api.Revision{
			api.FromEpoch(epoch.Hourly{}).ID: []api.Revision{
				api.FromRevision(a),
				api.FromRevision(b),
			},
			api.FromEpoch(epoch.TwoHourly{}).ID: []api.Revision{
				api.FromRevision(c),
			},
			api.FromEpoch(epoch.FourHourly{}).ID: []api.Revision{
				api.FromRevision(d),
				api.FromRevision(e),
			},
			api.FromEpoch(epoch.EightHourly{}).ID: []api.Revision{
				api.FromRevision(f),
			},
		},
		Epochs: []api.Epoch{
			api.FromEpoch(epoch.Hourly{}),
			api.FromEpoch(epoch.TwoHourly{}),
			api.FromEpoch(epoch.FourHourly{}),
			api.FromEpoch(epoch.EightHourly{}),
		},
		Error: err.Error(),
	}

	actual := api.RevisionsFromEpochs(revs, err)

	assert.Equal(t, expected, actual)
}
