// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import (
	"bytes"
	"embed"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	mapset "github.com/deckarep/golang-set"

	"github.com/google/go-github/v74/github"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// nolint:gochecknoglobals // TODO: Fix gochecknoglobals lint error
var templates *template.Template

//go:embed templates/*.md
var mdFiles embed.FS

func init() {
	templates = template.New("all.md").
		Funcs(template.FuncMap{
			"escapeMD": escapeMD,
		})
	_, err := templates.ParseFS(mdFiles, "templates/*.md")
	if err != nil {
		panic(err)
	}
}

// escapeMD returns the escaped MD equivalent of the plain text data s.
func escapeMD(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

// Summary is the generic interface of a summary template data type.
type Summary interface {
	// GetCheckState returns the info needed to update a check.
	GetCheckState() CheckState

	// GetActions returns the actions that can be taken by the user.
	GetActions() []*github.CheckRunAction

	// GetSummary compiles the summary markdown template.
	GetSummary() (string, error)
}

// CheckState represents all the status fields for updating a check.
type CheckState struct {
	HostName   string          // The host (e.g. wpt.fyi)
	TestRun    *shared.TestRun // The (completed) TestRun, if applicable.
	Product    shared.ProductSpec
	HeadSHA    string
	DetailsURL *url.URL
	// The current status. Can be one of "queued", "in_progress", or "completed". Default: "queued". (Optional.)
	Status string
	// Can be one of "success", "failure", "neutral", "cancelled", "timed_out", or "action_required".
	// (Optional. Required if you provide a status of "completed".)
	Conclusion *string
	Actions    []github.CheckRunAction
	PRNumbers  []int
}

// Name returns the check run's name, based on the product.
func (c CheckState) Name() string {
	host := c.HostName
	if host == "" {
		host = "wpt.fyi"
	}
	spec := shared.ProductSpec{} // nolint:exhaustruct // TODO: Fix exhaustruct lint error
	spec.BrowserName = c.Product.BrowserName
	if c.Product.IsExperimental() {
		spec.Labels = mapset.NewSetWith(shared.ExperimentalLabel)
	}

	return fmt.Sprintf("%s - %s", host, spec.String())
}

// Title returns the check run's title, based on the product.
func (c CheckState) Title() string {
	return fmt.Sprintf("%s results", c.Product.DisplayName())
}

// GetCheckState returns the info in the CheckState struct.
// It's a dumb placeholder since we can't define fields on interfaces.
func (c CheckState) GetCheckState() CheckState {
	return c
}

// FileIssueURL returns a URL for filing an issue in wpt.fyi repo about checks.
func (c CheckState) FileIssueURL() *url.URL {
	result, _ := url.Parse("https://github.com/web-platform-tests/wpt.fyi/issues/new")
	q := result.Query()
	q.Set("title", "Regression checks issue")
	q.Set("projects", "web-platform-tests/wpt.fyi/6")
	q.Set("template", "checks.md")
	q.Set("labels", "bug")
	result.RawQuery = q.Encode()

	return result
}

func compile(i interface{}, t string) (string, error) {
	var dest bytes.Buffer
	if err := templates.ExecuteTemplate(&dest, t, i); err != nil {
		return "", err
	}

	return dest.String(), nil
}
