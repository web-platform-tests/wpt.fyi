package webapp

import (
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

var componentTemplates *template.Template

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	glob := path.Join(dir, "dynamic-components/*.js")
	componentTemplates = template.Must(template.ParseGlob(glob))
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
