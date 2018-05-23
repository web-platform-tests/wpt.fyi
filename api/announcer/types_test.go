// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package announcer_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/announcer"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
	agit "github.com/web-platform-tests/wpt.fyi/revisions/git"
	"github.com/web-platform-tests/wpt.fyi/revisions/test"
)

var today = time.Now()
var yesterday = today.Add(-26 * time.Hour)

func TestFromEpoch(t *testing.T) {
	epochIn := epoch.Daily{}
	dataIn := epochIn.GetData()

	epochOut := announcer.FromEpoch(epochIn)

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

			assert.NotEqual(t, announcer.FromEpoch(e1).ID, announcer.FromEpoch(e2).ID)
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

	_, err := announcer.LatestFromEpochs(revs)
	assert.Equal(t, announcer.GetErMissingRevision(), err)
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
	expected := announcer.LatestResponse{
		Revisions: map[string]announcer.Revision{
			announcer.FromEpoch(epoch.Hourly{}).ID:      announcer.FromRevision(a),
			announcer.FromEpoch(epoch.TwoHourly{}).ID:   announcer.FromRevision(c),
			announcer.FromEpoch(epoch.FourHourly{}).ID:  announcer.FromRevision(d),
			announcer.FromEpoch(epoch.EightHourly{}).ID: announcer.FromRevision(f),
		},
		Epochs: []announcer.Epoch{
			announcer.FromEpoch(epoch.Hourly{}),
			announcer.FromEpoch(epoch.TwoHourly{}),
			announcer.FromEpoch(epoch.FourHourly{}),
			announcer.FromEpoch(epoch.EightHourly{}),
		},
	}

	actual, err := announcer.LatestFromEpochs(revs)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestRevisionsFromEpochs_err(t *testing.T) {
	err := errors.New("Announcer error")
	response := announcer.RevisionsFromEpochs(map[epoch.Epoch][]agit.Revision{}, err)
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

	expected := announcer.RevisionsResponse{
		Revisions: map[string][]announcer.Revision{
			announcer.FromEpoch(epoch.Hourly{}).ID: []announcer.Revision{
				announcer.FromRevision(a),
				announcer.FromRevision(b),
			},
			announcer.FromEpoch(epoch.TwoHourly{}).ID: []announcer.Revision{
				announcer.FromRevision(c),
			},
			announcer.FromEpoch(epoch.FourHourly{}).ID: []announcer.Revision{
				announcer.FromRevision(d),
				announcer.FromRevision(e),
			},
			announcer.FromEpoch(epoch.EightHourly{}).ID: []announcer.Revision{
				announcer.FromRevision(f),
			},
		},
		Epochs: []announcer.Epoch{
			announcer.FromEpoch(epoch.Hourly{}),
			announcer.FromEpoch(epoch.TwoHourly{}),
			announcer.FromEpoch(epoch.FourHourly{}),
			announcer.FromEpoch(epoch.EightHourly{}),
		},
	}

	actual := announcer.RevisionsFromEpochs(revs, nil)

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

	expected := announcer.RevisionsResponse{
		Revisions: map[string][]announcer.Revision{
			announcer.FromEpoch(epoch.Hourly{}).ID: []announcer.Revision{
				announcer.FromRevision(a),
				announcer.FromRevision(b),
			},
			announcer.FromEpoch(epoch.TwoHourly{}).ID: []announcer.Revision{
				announcer.FromRevision(c),
			},
			announcer.FromEpoch(epoch.FourHourly{}).ID: []announcer.Revision{
				announcer.FromRevision(d),
				announcer.FromRevision(e),
			},
			announcer.FromEpoch(epoch.EightHourly{}).ID: []announcer.Revision{
				announcer.FromRevision(f),
			},
		},
		Epochs: []announcer.Epoch{
			announcer.FromEpoch(epoch.Hourly{}),
			announcer.FromEpoch(epoch.TwoHourly{}),
			announcer.FromEpoch(epoch.FourHourly{}),
			announcer.FromEpoch(epoch.EightHourly{}),
		},
		Error: err.Error(),
	}

	actual := announcer.RevisionsFromEpochs(revs, err)

	assert.Equal(t, expected, actual)
}
