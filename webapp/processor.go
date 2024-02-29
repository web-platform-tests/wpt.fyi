// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"
)

// processorStatusHandler handles GET requests to /processor
func processorStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is supported.", http.StatusMethodNotAllowed)
		return
	}

	RenderTemplate(w, r, "processor.html", nil)
}
