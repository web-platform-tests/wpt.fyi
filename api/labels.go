// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api //nolint:revive

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// LabelsHandler is an http.Handler for the /api/labels endpoint.
type LabelsHandler struct {
	ctx context.Context // nolint:containedctx // TODO: Fix containedctx lint error
}

// apiLabelsHandler is responsible for emitting just all labels used for test runs.
func apiLabelsHandler(w http.ResponseWriter, r *http.Request) {
	// Serve cached with 5 minute expiry. Delegate to LabelsHandler on cache miss.
	ctx := r.Context()
	shared.NewCachingHandler(
		ctx,
		LabelsHandler{ctx},
		shared.NewGZReadWritable(shared.NewRedisReadWritable(ctx, 5*time.Minute)),
		shared.AlwaysCachable,
		shared.URLAsCacheKey,
		shared.CacheStatusOK,
	).ServeHTTP(w, r)
}

func (h LabelsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	store := shared.NewAppEngineDatastore(h.ctx, false)
	var runs shared.TestRuns
	_, err := store.GetAll(store.NewQuery("TestRun").Project("Labels").Distinct(), &runs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	all := mapset.NewSet()
	for _, run := range runs {
		for _, label := range run.Labels {
			all.Add(label)
		}
	}
	labels := shared.ToStringSlice(all)
	sort.Strings(labels)
	data, err := json.Marshal(labels)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	_, err = w.Write(data)
	if err != nil {
		logger := shared.GetLogger(r.Context())
		logger.Warningf("Failed to write data in api/labels handler: %s", err.Error())
	}
}
