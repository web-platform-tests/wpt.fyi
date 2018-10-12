// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package push

import (
	"time"

	"github.com/web-platform-tests/wpt.fyi/revisions/announcer"
	"github.com/web-platform-tests/wpt.fyi/revisions/api"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
)

// GetLatestRevisions fetches the latest revisions from ancr using configration
// from a and epochs. Note that epochs is assumed to be sorted by epoch
// duration, descending.
func GetLatestRevisions(a api.API, ancr announcer.Announcer, epochs []epoch.Epoch) (*api.LatestResponse, error) {
	now := time.Now()
	revs, err := ancr.GetRevisions(a.GetLatestGetRevisionsInput(), announcer.Limits{
		At:    now,
		Start: now.Add(-2 * epochs[0].GetData().MaxDuration),
	})
	if err != nil {
		return nil, err
	}

	response, err := api.LatestFromEpochs(revs)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// DiffLatest returns the differences between same-key values in prev and next,
// processing only those keys that correspond to epochs (by api.Epoch.ID value).
func DiffLatest(prev map[string]api.Revision, next map[string]api.Revision, epochs []epoch.Epoch) []api.Diff {
	changes := make([]api.Diff, 0, len(epochs))

	if prev == nil && next == nil {
		return changes
	}

	if prev == nil {
		for _, epoch := range epochs {
			e := api.FromEpoch(epoch)
			key := e.ID
			r := next[key]
			changes = append(changes, api.Diff{e.ID, nil, &r})
		}

		return changes
	}

	if next == nil {
		for _, epoch := range epochs {
			e := api.FromEpoch(epoch)
			key := e.ID
			r := prev[key]
			changes = append(changes, api.Diff{e.ID, &r, nil})
		}

		return changes
	}

	for _, epoch := range epochs {
		e := api.FromEpoch(epoch)
		key := e.ID
		if !prev[key].Equal(next[key]) {
			pr, pok := prev[key]
			nr, nok := next[key]
			if pr != nr {
				if !pok {
					changes = append(changes, api.Diff{e.ID, nil, &nr})
				} else if !nok {
					changes = append(changes, api.Diff{e.ID, &pr, nil})
				} else {
					changes = append(changes, api.Diff{e.ID, &pr, &nr})
				}
			}
		}
	}

	return changes
}
