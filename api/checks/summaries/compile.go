// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/deckarep/golang-set"

	"github.com/lukebjerring/go-github/github"
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
	Status     string  // The current status. Can be one of "queued", "in_progress", or "completed". Default: "queued". (Optional.)
	Conclusion *string // Can be one of "success", "failure", "neutral", "cancelled", "timed_out", or "action_required". (Optional. Required if you provide a status of "completed".)
	Actions    []github.CheckRunAction
}

// Name returns the check run's name, based on the product.
func (c CheckState) Name() string {
	spec := shared.ProductSpec{}
	spec.BrowserName = c.Product.BrowserName
	if c.Product.IsExperimental() {
		spec.Labels = mapset.NewSetWith(shared.ExperimentalLabel)
	}
	return spec.String()
}

// Title returns the check run's title, based on the product.
func (c CheckState) Title() string {
	return fmt.Sprintf("wpt.fyi - %s results", c.Product.DisplayName())
}

// GetCheckState returns the info in the CheckState struct.
// It's a dumb placeholder since we can't define fields on interfaces.
func (c CheckState) GetCheckState() CheckState {
	return c
}

func compile(i interface{}, t string) (string, error) {
	var dest bytes.Buffer
	if err := templates.ExecuteTemplate(&dest, t, i); err != nil {
		return "", err
	}
	return dest.String(), nil
}
