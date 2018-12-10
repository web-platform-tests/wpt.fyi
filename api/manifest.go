// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/api/manifest"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func apiManifestHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	sha, err := shared.ParseSHAParamFull(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	paths := shared.ParsePathsParam(q)
	ctx := shared.NewAppEngineContext(r)
	manifestAPI := manifest.NewAPI(ctx)
	sha, manifestBytes, err := manifestAPI.GetManifestForSHA(sha)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Add("wpt-sha", sha)
	w.Header().Add("Content-Type", "application/json")
	if paths != nil {
		if manifestBytes, err = manifest.Filter(manifestBytes, paths); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Write(manifestBytes)
}
