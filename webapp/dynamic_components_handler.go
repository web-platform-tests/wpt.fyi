package webapp

import (
	"net/http"
	"text/template"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

var componentTemplates = template.Must(template.ParseGlob("dynamic-components/*.js"))

func flagsComponentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/javascript")
	ctx := shared.NewAppEngineContext(r)
	flags, err := shared.GetFeatureFlags(ctx)
	if err != nil {
		// Errors aren't a big deal; log them and ignore.
		log := shared.GetLogger(ctx)
		log.Errorf("Error loading flags: %s", err.Error())
	}
	data := struct{ Flags []shared.Flag }{flags}
	componentTemplates.ExecuteTemplate(w, "wpt-env-flags.js", data)
}
