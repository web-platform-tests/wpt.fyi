package webapp

import (
	"html/template"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
)

var componentTemplates = template.Must(template.ParseGlob("dynamic-components/*.html"))

func flagsComponentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	flags, _ := shared.GetFeatureFlags(ctx) // Errors aren't a big deal.
	data := struct{ Flags []shared.Flag }{flags}
	componentTemplates.ExecuteTemplate(w, "wpt-env-flags.html", data)
}
