// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package checks

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine/datastore"
)

func TestLoadRunsToCompare(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	testRun := shared.TestRun{
		ProductAtRevision: shared.ProductAtRevision{
			Product: shared.Product{
				BrowserName:    "chrome",
				BrowserVersion: "63.0",
				OSName:         "linux",
			},
		},
		Labels: []string{"master"},
	}
	yesterday := time.Now().AddDate(0, 0, -1)
	for i := 0; i < 2; i++ {
		testRun.Revision = strings.Repeat(strconv.Itoa(i), 10)
		testRun.TimeStart = yesterday.Add(time.Duration(i) * time.Hour)
		key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
		key, _ = datastore.Put(ctx, key, &testRun)
	}

	chrome, _ := shared.ParseProductSpec("chrome")
	filter := shared.TestRunFilter{
		SHA:      strings.Repeat("1", 10),
		Products: shared.ProductSpecs{chrome},
	}
	prRun, masterRun, err := loadRunsToCompare(ctx, filter)
	assert.Nil(t, err)
	if prRun == nil || masterRun == nil {
		assert.FailNow(t, "Nil run(s) returned")
	}
	assert.NotEqual(t, prRun.Revision, masterRun.Revision)
}
