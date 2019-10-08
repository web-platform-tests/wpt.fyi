// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate packr2

package webapp

import (
	"net/http"
	"text/template"

	"github.com/gobuffalo/packr/v2"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

var componentTemplates *template.Template

func init() {
	box := packr.New("dynamic components", "./dynamic-components/")
	componentTemplates = template.New("all.js")
	for _, t := range box.List() {
		tmpl := componentTemplates.New(t)
		body, err := box.FindString(t)
		if err != nil {
			panic(err)
		} else if _, err = tmpl.Parse(body); err != nil {
			panic(err)
		}
	}
}

func flagsComponentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/javascript")
	ctx := shared.NewAppEngineContext(r)
	ds := shared.NewAppEngineDatastore(ctx, false)
	flags, err := shared.GetFeatureFlags(ds)
	if err != nil {
		// Errors aren't a big deal; log them and ignore.
		log := shared.GetLogger(ctx)
		log.Errorf("Error loading flags: %s", err.Error())
	}
	data := struct{ Flags []shared.Flag }{flags}
	componentTemplates.ExecuteTemplate(w, "wpt-env-flags.js", data)
}
