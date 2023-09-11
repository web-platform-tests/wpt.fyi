// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"fmt"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

func analyzerHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	screenshots := q["screenshot"]
	before, after := q.Get("before"), q.Get("after")
	if before != "" {
		screenshots = append(screenshots, before)
	}
	if after != "" {
		screenshots = append(screenshots, after)
	}
	if len(screenshots) != 2 {
		http.Error(w, "Expected exactly 2 screenshots (before + after)", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	aeAPI := shared.NewAppEngineAPI(ctx)
	/* #nosec G101 */
	bucket := "wptd-screenshots-staging"
	if aeAPI.GetHostname() == "wpt.fyi" {
		/* #nosec G101 */
		bucket = "wptd-screenshots"
	}

	data := struct {
		Before string
		After  string
	}{
		fmt.Sprintf("https://storage.googleapis.com/%s/%s.png", bucket, screenshots[0]),
		fmt.Sprintf("https://storage.googleapis.com/%s/%s.png", bucket, screenshots[1]),
	}
	RenderTemplate(w, r, "analyzer.html", data)
}
