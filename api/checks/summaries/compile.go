package summaries

import (
	"bytes"
	"net/url"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

var templates *template.Template

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	templates = template.Must(template.ParseGlob(dir + "/*.md"))
}

// Summary is the generic interface of a summary template data type.
type Summary interface {
	// GetCheckState returns the info needed to update a check.
	GetCheckState() CheckState

	// GetSummary compiles the summary markdown template.
	GetSummary() (string, error)
}

// CheckState represents all the status fields for updating a check.
type CheckState struct {
	Product    shared.ProductSpec
	HeadSHA    string
	DetailsURL *url.URL
	Title      string
	Status     string  // The current status. Can be one of "queued", "in_progress", or "completed". Default: "queued". (Optional.)
	Conclusion *string // Can be one of "success", "failure", "neutral", "cancelled", "timed_out", or "action_required". (Optional. Required if you provide a status of "completed".)
}

// Completed is the struct for completed.md
type Completed struct {
	CheckState

	DiffURL  string // URL for the diff-view of the results
	HostName string // Host environment name, e.g. "wpt.fyi"
	HostURL  string // Host environment URL, e.g. "https://wpt.fyi"
	SHAURL   string // URL for the latest results for the same SHA
}

// GetCheckState returns the info needed to update a check.
func (c Completed) GetCheckState() CheckState {
	return c.CheckState
}

// GetSummary executes the template for the data.
func (c Completed) GetSummary() (string, error) {
	return compile(&c, "completed.md")
}

// Pending is the struct for pending.md
type Pending struct {
	CheckState

	HostName string // Host environment name
	RunsURL  string // URL for the list of test runs
}

// GetCheckState returns the info needed to update a check.
func (c Pending) GetCheckState() CheckState {
	return c.CheckState
}

// GetSummary executes the template for the data.
func (c Pending) GetSummary() (string, error) {
	return compile(&c, "pending.md")
}

func compile(i interface{}, t string) (string, error) {
	var dest bytes.Buffer
	if err := templates.ExecuteTemplate(&dest, t, i); err != nil {
		return "", err
	}
	return dest.String(), nil
}

// BeforeAndAfter is a struct summarizing pass rates before and after in a diff.
type BeforeAndAfter struct {
	PassingBefore int
	PassingAfter  int
	TotalBefore   int
	TotalAfter    int
}

// Regressed is the struct for regressed.md
type Regressed struct {
	CheckState

	HostName    string
	HostURL     string
	DiffURL     string
	Regressions map[string]BeforeAndAfter
	More        int
}

// GetCheckState returns the info needed to update a check.
func (r Regressed) GetCheckState() CheckState {
	return r.CheckState
}

// GetSummary executes the template for the data.
func (r Regressed) GetSummary() (string, error) {
	return compile(&r, "regressed.md")
}
