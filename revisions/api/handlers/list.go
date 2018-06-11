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
		http.Error(w, a.ErrorJSON("Announcer not yet initialized"), http.StatusServiceUnavailable)
		return
	}

	epochs := a.GetEpochs()
	if len(epochs) == 0 {
		http.Error(w, a.ErrorJSON("No epochs"), http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()

	numRevisions := 1
	if nr, ok := q["num_revisions"]; ok {
		if len(nr) > 1 {
			http.Error(w, a.ErrorJSON("Multiple num_revisions values"), http.StatusBadRequest)
			return
		}
		if len(nr) == 0 {
			http.Error(w, a.ErrorJSON("Empty num_revisions value"), http.StatusBadRequest)
			return
		}
		var err error
		numRevisions, err = strconv.Atoi(nr[0])
		if err != nil {
			http.Error(w, a.ErrorJSON(fmt.Sprintf("Invalid num_revisions value: %s", nr[0])), http.StatusBadRequest)
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
				http.Error(w, a.ErrorJSON(fmt.Sprintf("Unknown epoch: %s", eStr)), http.StatusBadRequest)
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

	at := time.Now()
	if tStrs, ok := q["at"]; ok {
		if len(tStrs) > 1 {
			http.Error(w, a.ErrorJSON("Multiple at values"), http.StatusBadRequest)
			return
		}
		if len(tStrs) == 0 {
			http.Error(w, a.ErrorJSON("Empty at value"), http.StatusBadRequest)
			return
		}
		var err error
		at, err = time.Parse(time.RFC3339, tStrs[0])
		if err != nil {
			http.Error(w, a.ErrorJSON(fmt.Sprintf("Invalid at value: %s", tStrs[0])), http.StatusBadRequest)
			return
		}
	}

	start := at.Add(time.Duration(-1-numRevisions) * epochs[0].GetData().MaxDuration)
	if tStrs, ok := q["start"]; ok {
		if len(tStrs) > 1 {
			http.Error(w, a.ErrorJSON("Multiple start values"), http.StatusBadRequest)
			return
		}
		if len(tStrs) == 0 {
			http.Error(w, a.ErrorJSON("Empty start value"), http.StatusBadRequest)
			return
		}
		var err error
		start, err = time.Parse(time.RFC3339, tStrs[0])
		if err != nil {
			http.Error(w, a.ErrorJSON(fmt.Sprintf("Invalid start value: %s", tStrs[0])), http.StatusBadRequest)
			return
		}
	}

	if at.Before(start) {
		http.Error(w, a.ErrorJSON(fmt.Sprintf("At parameter (%v) occurs before start parameter (%v)", at, start)), http.StatusBadRequest)
		return
	}

	revs, err := ancr.GetRevisions(getRevisions, announcer.Limits{
		At:    at,
		Start: start,
	})
	if revs == nil && err != nil {
		http.Error(w, a.ErrorJSON(err.Error()), http.StatusInternalServerError)
		return
	}

	response := api.RevisionsFromEpochs(revs, err)
	bytes, err := a.Marshal(response)
	if err != nil {
		http.Error(w, a.ErrorJSON("Failed to marshal latest epochal revisions JSON"), 500)
		return
	}

	w.Write(bytes)
}
