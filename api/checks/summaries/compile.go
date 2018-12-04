// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import (
	"bytes"
	"net/url"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/google/go-github/github"
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

	// GetActions returns the actions that can be taken by the user.
	GetActions() []*github.CheckRunAction

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
	Actions    []github.CheckRunAction
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
