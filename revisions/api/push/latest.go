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

// DiffLatest returns an api.LatestResponse that contains the new values for
// given epochs that changed in value between prev and next.
func DiffLatest(prev *api.LatestResponse, next *api.LatestResponse, epochs []epoch.Epoch) []api.Diff {
	changes := make([]api.Diff, 0, len(epochs))

	if prev == nil && next == nil {
		return changes
	}

	if prev == nil {
		for _, epoch := range epochs {
			e := api.FromEpoch(epoch)
			key := e.ID
			r := next.Revisions[key]
			changes = append(changes, api.Diff{e.ID, nil, &r})
		}

		return changes
	}

	if next == nil {
		for _, epoch := range epochs {
			e := api.FromEpoch(epoch)
			key := e.ID
			r := prev.Revisions[key]
			changes = append(changes, api.Diff{e.ID, &r, nil})
		}

		return changes
	}

	for _, epoch := range epochs {
		e := api.FromEpoch(epoch)
		key := e.ID
		if prev.Revisions[key] != next.Revisions[key] {
			pr := prev.Revisions[key]
			nr := next.Revisions[key]
			changes = append(changes, api.Diff{e.ID, &pr, &nr})
		}
	}

	return changes
}
