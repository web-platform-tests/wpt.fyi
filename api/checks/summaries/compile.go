package summaries

import (
	"bytes"
	"html/template"
	"path/filepath"
	"reflect"
	"runtime"
)

var templates *template.Template

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	templates = template.Must(template.ParseGlob(dir + "/*.md"))
}

// Completed is the struct for completed.md
type Completed struct {
	HostName string // Host environment name, e.g. "wpt.fyi"
	HostURL  string // Host environment URL, e.g. "https://wpt.fyi"
	DiffURL  string // URL for the diff-view of the results
	SHAURL   string // URL for the latest results for the same SHA
}

// Compile executes the template for the data.
func (c Completed) Compile() (string, error) {
	return compile(&c, "completed.md")
}

// Pending is the struct for pending.md
type Pending struct {
	HostName string // Host environment name
	RunsURL  string // URL for the list of test runs
}

// Compile executes the template for the data.
func (c Pending) Compile() (string, error) {
	return compile(&c, "pending.md")
}

func compile(i interface{}, t string) (string, error) {
	// Copy all the fields as template.HTML to avoid HTML escaping.
	values := make(map[string]template.HTML)
	v := reflect.Indirect(reflect.ValueOf(i))
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.CanInterface() {
			if s, ok := field.Interface().(string); ok {
				values[v.Type().Field(i).Name] = template.HTML(s)
			}
		}
	}

	var dest bytes.Buffer
	if err := templates.ExecuteTemplate(&dest, t, values); err != nil {
		return "", err
	}
	return dest.String(), nil
}
