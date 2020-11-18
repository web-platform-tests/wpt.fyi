// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate packr2

package webapp

import (
	"html/template"
	"net/http"

	"github.com/gobuffalo/packr/v2"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// templates contains all of the template objects parsed from templates/*,
// created once at startup. Use RenderTemplate to render responses.
var templates *template.Template

func init() {
	box := packr.New("html templates", "./templates/")
	templates = template.New("all.html")
	for _, t := range box.List() {
		tmpl := templates.New(t)
		body, err := box.FindString(t)
		if err != nil {
			panic(err)
		} else if _, err = tmpl.Parse(body); err != nil {
			panic(err)
		}
	}
}

type templateData struct {
	Data interface{}
	User *shared.User
}

// RenderTemplate renders an HTML template to a response. The provided data
// will be available in the Data field in the template. There are some
// additional fields extracted from the request (e.g. User) available in the
// template if the request is not nil.
// If an error is encountered, appropriate error codes and messages will be set
// on the response; do not write additional data to the response after calling
// this function.
func RenderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	tdata := templateData{Data: data}
	if r != nil {
		ctx := r.Context()
		ds := shared.NewAppEngineDatastore(ctx, false)
		tdata.User, _ = shared.GetUserFromCookie(ctx, ds, r)
	}
	if err := templates.ExecuteTemplate(w, name, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
