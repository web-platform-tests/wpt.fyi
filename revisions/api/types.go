// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"errors"
	"reflect"
	"sort"
	"time"

	strcase "github.com/stoewer/go-strcase"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
	agit "github.com/web-platform-tests/wpt.fyi/revisions/git"
)

var errMissingRevision = errors.New("Missing required revision")

// GetErMissingRevision produces the error produced when a required revision is not provided.
func GetErMissingRevision() error {
	return errMissingRevision
}

// Epoch is the representation of an epoch exposed via the public API.
type Epoch struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	MinDuration float32 `json:"min_duration_sec"`
	MaxDuration float32 `json:"max_duration_sec"`
}

// FromEpoch converts an epoch.Epoch to an epoch exposed via the public API.
func FromEpoch(e epoch.Epoch) Epoch {
	t := reflect.TypeOf(e)
	v := reflect.ValueOf(e)
	for t.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
		t = v.Type()
	}
	id := strcase.SnakeCase(t.Name())
	d := e.GetData()
	minDuration := float32(d.MinDuration.Seconds())
	maxDuration := float32(d.MaxDuration.Seconds())
	return Epoch{
		id,
		d.Label,
		d.Description,
		minDuration,
		maxDuration,
	}
}

// Revision is the representation of a git revision exposed via the public API.
type Revision struct {
	Hash       string    `json:"hash"`
	CommitTime time.Time `json:"commit_time"`
}

// FromRevision converts an agit.Revision to a revision exposed via the public API.
func FromRevision(rev agit.Revision) Revision {
	return Revision{
		Hash:       rev.GetHash().String(),
		CommitTime: rev.GetCommitTime(),
	}
}

// LatestRequest models a request for the latest announced revisions.
type LatestRequest struct{}

// LatestResponse models a response for the latest announced revisions.
type LatestResponse struct {
	Revisions map[string]Revision `json:"revisions"`
	Epochs    []Epoch             `json:"epochs"`
}

// LatestFromEpochs formats a map[epoch.Epoch][]agit.Revision from the announcer into a LatestResponse.
func LatestFromEpochs(revs map[epoch.Epoch][]agit.Revision) (LatestResponse, error) {
	epochs := make([]epoch.Epoch, 0, len(revs))
	for e := range revs {
		epochs = append(epochs, e)
	}
	sort.Sort(epoch.ByMaxDuration(epochs))
	es := make([]Epoch, 0, len(epochs))
	for _, e := range epochs {
		es = append(es, FromEpoch(e))
	}

	rs := make(map[string]Revision)

	for i := range es {
		if len(revs[epochs[i]]) == 0 {
			continue
		}
		rev := revs[epochs[i]][0]
		rs[es[i].ID] = FromRevision(rev)
	}

	latest := LatestResponse{
		rs,
		es,
	}

	if len(rs) < len(epochs) {
		return latest, errMissingRevision
	}

	return latest, nil
}

// EpochsRequest models a request for the epochs supported by the service.
type EpochsRequest struct{}

// EpochsResponse models a response for the epochs supported by the service.
type EpochsResponse []Epoch

// RevisionsRequest models a request for the announced revisions.
type RevisionsRequest struct {
	Epochs       []epoch.Epoch `json:"epochs,omitempty"`
	NumRevisions int           `json:"num_revisions,omitempty"`
	Now          time.Time     `json:"now,omitempty"`
	Start        time.Time     `json:"start,omitempty"`
}

// RevisionsResponse models a response for the announced revisions.
type RevisionsResponse struct {
	Revisions map[string][]Revision `json:"revisions"`
	Epochs    []Epoch               `json:"epochs"`
	Error     string                `json:"error,omitempty"`
}

// RevisionsFromEpochs formats a map[epoch.Epoch][]agit.Revision + internal announcer API error into a RevisionsResponse.
func RevisionsFromEpochs(revs map[epoch.Epoch][]agit.Revision, apiErr error) RevisionsResponse {
	epochs := make([]epoch.Epoch, 0, len(revs))
	for e := range revs {
		epochs = append(epochs, e)
	}
	sort.Sort(epoch.ByMaxDuration(epochs))
	es := make([]Epoch, 0, len(epochs))
	for _, e := range epochs {
		es = append(es, FromEpoch(e))
	}

	rs := make(map[string][]Revision)

	for i := range es {
		if len(revs[epochs[i]]) == 0 {
			continue
		}
		revs := revs[epochs[i]]
		apiRevs := make([]Revision, 0, len(revs))
		for _, rev := range revs {
			apiRevs = append(apiRevs, Revision{
				Hash:       rev.GetHash().String(),
				CommitTime: rev.GetCommitTime(),
			})
		}
		rs[es[i].ID] = apiRevs
	}

	var response RevisionsResponse
	if apiErr != nil {
		response = RevisionsResponse{
			rs,
			es,
			apiErr.Error(),
		}
	} else {
		response = RevisionsResponse{
			rs,
			es,
			"",
		}
	}

	return response
}
