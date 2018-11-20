package summaries

import (
	"bytes"
	"path/filepath"
	"runtime"
	"text/template"
)

var templates *template.Template

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	templates = template.Must(template.ParseGlob(dir + "/*.md"))
}

// Completed is the struct for completed.md
type Completed struct {
	DiffURL  string // URL for the diff-view of the results
	HostName string // Host environment name, e.g. "wpt.fyi"
	HostURL  string // Host environment URL, e.g. "https://wpt.fyi"
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
	var dest bytes.Buffer
	if err := templates.ExecuteTemplate(&dest, t, i); err != nil {
		return "", err
	}
	return dest.String(), nil
}
