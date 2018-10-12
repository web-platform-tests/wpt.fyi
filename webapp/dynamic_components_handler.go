package webapp

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
)

var componentTemplates = template.Must(template.ParseGlob("dynamic-components/*.html"))

func flagsComponentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	flags, _ := shared.GetFeatureFlags(ctx) // Errors aren't a big deal.
	flagsBytes, err := json.Marshal(flags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(flags) < 1 {
		flagsBytes = []byte("[]")
	}
	data := struct{ Flags string }{string(flagsBytes)}
	componentTemplates.ExecuteTemplate(w, "wpt-env-flags.html", data)
}
