package webapp

import (
	"html/template"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

var componentTemplates = template.Must(template.ParseGlob("dynamic-components/*.html"))

func flagsComponentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	flags, err := shared.GetFeatureFlags(ctx)
	if err != nil {
		// Errors aren't a big deal; log them and ignore.
		log := shared.GetLogger(ctx)
		log.Debugf("Error loading flags: %s", err.Error())
	}
	data := struct{ Flags []shared.Flag }{flags}
	componentTemplates.ExecuteTemplate(w, "wpt-env-flags.html", data)
}
