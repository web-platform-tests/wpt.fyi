// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"net/http"

	"google.golang.org/appengine"

	"github.com/web-platform-tests/wpt.fyi/api/receiver"
)

func apiResultsReceiveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	ctx := appengine.NewContext(r)
	a := receiver.NewAppEngineAPI(ctx)
	receiver.HandleResultsUpload(a, w, r)
}
