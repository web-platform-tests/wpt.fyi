// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// templates contains all of the template objects parsed from templates/*,
// created once at startup. Use RenderTemplate to render responses.
var templates *template.Template
//go:embed templates/*.html
var htmlTemplates embed.FS

func init() {
	templates = template.New("all.html")
	_, err := templates.ParseFS(htmlTemplates, "templates/*.html")
	if err != nil {
		panic(err)
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
