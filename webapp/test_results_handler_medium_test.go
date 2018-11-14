// +build medium

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func TestParseTestResultsUIFilter(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, _ := i.NewRequest("GET", "/results/?max-count=3", nil)
	assert.Nil(t, err)

	f, err := parseTestResultsUIFilter(r)
	assert.Nil(t, err)
	assert.True(t, f.MaxCount != nil && *f.MaxCount == 3)

	r, _ = i.NewRequest("GET", "/results/?products=chrome,safari&diff", nil)
	f, err = parseTestResultsUIFilter(r)
	assert.Nil(t, err)
	assert.Equal(t, "[\"chrome\",\"safari\"]", f.Products)
	assert.Equal(t, true, f.Diff)

	r, _ = i.NewRequest("GET", "/results/?before=chrome&after=chrome[experimental]", nil)
	f, err = parseTestResultsUIFilter(r)
	assert.Nil(t, err)
	assert.Equal(t, "[\"chrome\",\"chrome[experimental]\"]", f.Products)
	assert.Equal(t, true, f.Diff)

	r, _ = i.NewRequest("GET", "/results/?after=chrome&before=edge", nil)
	f, err = parseTestResultsUIFilter(r)
	assert.Nil(t, err)
	assert.Equal(t, "[\"edge\",\"chrome\"]", f.Products)
	assert.Equal(t, true, f.Diff)

	// MasterOnly default query.
	ctx := appengine.NewContext(r)
	datastore.Put(ctx, datastore.NewKey(ctx, "Flag", "masterRunsOnly", 0, nil), &shared.Flag{Enabled: true})
	r, _ = i.NewRequest("GET", "/results/", nil)
	f, err = parseTestResultsUIFilter(r)
	assert.Nil(t, err)
	assert.Contains(t, f.Labels, shared.MasterLabel)

	r, _ = i.NewRequest("GET", "/results/?sha=0123456789", nil)
	f, err = parseTestResultsUIFilter(r)
	assert.Nil(t, err)
	assert.NotContains(t, f.Labels, shared.MasterLabel)
}
