// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"embed"
	"net/http"
	"text/template"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

var componentTemplates *template.Template
//go:embed dynamic-components/templates/*
var dcTemplates embed.FS

func init() {
	componentTemplates = template.New("all.js")
	_, err := componentTemplates.ParseFS(dcTemplates, "dynamic-components/templates/*")
	if err != nil {
		panic(err)
	}
}

func flagsComponentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/javascript")
	ctx := r.Context()
	ds := shared.NewAppEngineDatastore(ctx, false)
	flags, err := shared.GetFeatureFlags(ds)
	if err != nil {
		// Errors aren't a big deal; log them and ignore.
		log := shared.GetLogger(ctx)
		log.Errorf("Error loading flags: %s", err.Error())
	}
	data := struct{ Flags []shared.Flag }{flags}
	if componentTemplates.ExecuteTemplate(w, "wpt-env-flags.js", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
