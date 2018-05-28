// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/web-platform-tests/wpt.fyi/revisions/announcer"
	"github.com/web-platform-tests/wpt.fyi/revisions/api"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
)

// ListHandler handles HTTP requests for listing epochal revisions.
func ListHandler(a api.API, w http.ResponseWriter, r *http.Request) {
	ancr := a.GetAnnouncer()
	if ancr == nil {
		w.WriteHeader(503)
		w.Write(a.ErrorJSON("Announcer not yet initialized"))
		return
	}

	epochs := a.GetEpochs()
	if len(epochs) == 0 {
		w.WriteHeader(500)
		w.Write(a.ErrorJSON("No epochs"))
		return
	}

	q := r.URL.Query()

	numRevisions := 1
	if nr, ok := q["num_revisions"]; ok {
		if len(nr) > 1 {
			w.WriteHeader(500)
			w.Write(a.ErrorJSON("Multiple num_revisions values"))
			return
		}
		if len(nr) == 0 {
			w.WriteHeader(500)
			w.Write(a.ErrorJSON("Empty num_revisions value"))
			return
		}
		var err error
		numRevisions, err = strconv.Atoi(nr[0])
		if err != nil {
			w.WriteHeader(500)
			w.Write(a.ErrorJSON(fmt.Sprintf("Invalid num_revisions value: %s", nr[0])))
			return
		}
	}

	getRevisions := make(map[epoch.Epoch]int)
	if eStrs, ok := q["epochs"]; ok {
		epochsMap := a.GetEpochsMap()
		for _, eStr := range eStrs {
			if e, ok := epochsMap[eStr]; ok {
				getRevisions[e] = numRevisions
			} else {
				w.WriteHeader(500)
				w.Write(a.ErrorJSON(fmt.Sprintf("Unknown epoch: %s", eStr)))
				return
			}
		}
	} else {
		latestGetRevisions := a.GetLatestGetRevisionsInput()
		for e := range latestGetRevisions {
			getRevisions[e] = numRevisions
		}
	}

	es := make([]epoch.Epoch, 0, len(getRevisions))
	for e := range getRevisions {
		es = append(es, e)
	}
	sort.Sort(epoch.ByMaxDuration(es))

	now := time.Now()
	if tStrs, ok := q["now"]; ok {
		if len(tStrs) > 1 {
			w.WriteHeader(500)
			w.Write(a.ErrorJSON("Multiple now values"))
			return
		}
		if len(tStrs) == 0 {
			w.WriteHeader(500)
			w.Write(a.ErrorJSON("Empty now value"))
			return
		}
		var err error
		now, err = time.Parse("", tStrs[0])
		if err != nil {
			w.WriteHeader(500)
			w.Write(a.ErrorJSON(fmt.Sprintf("Invalid now value: %s", tStrs[0])))
			return
		}
	}

	start := now.Add(time.Duration(-1-numRevisions) * epochs[0].GetData().MaxDuration)
	if tStrs, ok := q["start"]; ok {
		if len(tStrs) > 1 {
			w.WriteHeader(500)
			w.Write(a.ErrorJSON("Multiple start values"))
			return
		}
		if len(tStrs) == 0 {
			w.WriteHeader(500)
			w.Write(a.ErrorJSON("Empty start value"))
			return
		}
		var err error
		now, err = time.Parse("", tStrs[0])
		if err != nil {
			w.WriteHeader(500)
			w.Write(a.ErrorJSON(fmt.Sprintf("Invalid start value: %s", tStrs[0])))
			return
		}
	}

	revs, err := ancr.GetRevisions(getRevisions, announcer.Limits{
		At:    now,
		Start: start,
	})
	if revs == nil && err != nil {
		w.WriteHeader(500)
		w.Write(a.ErrorJSON(err.Error()))
		return
	}

	response := api.RevisionsFromEpochs(revs, err)
	bytes, err := a.Marshal(response)
	if err != nil {
		w.WriteHeader(500)
		w.Write(a.ErrorJSON("Failed to marshal latest epochal revisions JSON"))
		return
	}

	w.Write(bytes)
}
