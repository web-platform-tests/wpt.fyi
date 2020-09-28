// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"net/http"
)

func flagsHandler(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, r, "flags.html", nil)
}
