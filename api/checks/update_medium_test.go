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

func TestLoadRunsToCompare_master(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	testRun := shared.TestRun{
		ProductAtRevision: shared.ProductAtRevision{
			Product: shared.Product{
				BrowserName: "chrome",
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
		SHA:      "1111111111",
		Products: shared.ProductSpecs{chrome},
	}
	headRun, baseRun, err := loadRunsToCompare(ctx, filter)

	assert.Nil(t, err)
	assert.NotNil(t, headRun)
	assert.NotNil(t, baseRun)
	assert.Equal(t, "0000000000", baseRun.Revision)
	assert.Equal(t, "1111111111", headRun.Revision)
}

func TestLoadRunsToCompare_pr_base_first(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	labelsForRuns := [][]string{{"pr_base"}, {"pr_head"}}
	yesterday := time.Now().AddDate(0, 0, -1)
	for i := 0; i < 2; i++ {
		testRun := shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName: "chrome",
				},
				Revision: "1234567890",
			},
			TimeStart: yesterday.Add(time.Duration(i) * time.Hour),
			Labels:    labelsForRuns[i],
		}
		key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
		key, _ = datastore.Put(ctx, key, &testRun)
	}

	chrome, _ := shared.ParseProductSpec("chrome")
	filter := shared.TestRunFilter{
		SHA:      "1234567890",
		Products: shared.ProductSpecs{chrome},
	}
	headRun, baseRun, err := loadRunsToCompare(ctx, filter)

	assert.Nil(t, err)
	assert.NotNil(t, headRun)
	assert.NotNil(t, baseRun)
	assert.Equal(t, []string{"pr_base"}, baseRun.Labels)
	assert.Equal(t, []string{"pr_head"}, headRun.Labels)
}

func TestLoadRunsToCompare_pr_head_first(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	labelsForRuns := [][]string{{"pr_head"}, {"pr_base"}}
	yesterday := time.Now().AddDate(0, 0, -1)
	for i := 0; i < 2; i++ {
		testRun := shared.TestRun{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName: "chrome",
				},
				Revision: "1234567890",
			},
			TimeStart: yesterday.Add(time.Duration(i) * time.Hour),
			Labels:    labelsForRuns[i],
		}
		key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
		key, _ = datastore.Put(ctx, key, &testRun)
	}

	chrome, _ := shared.ParseProductSpec("chrome")
	filter := shared.TestRunFilter{
		SHA:      "1234567890",
		Products: shared.ProductSpecs{chrome},
	}
	headRun, baseRun, err := loadRunsToCompare(ctx, filter)

	assert.Nil(t, err)
	assert.NotNil(t, headRun)
	assert.NotNil(t, baseRun)
	assert.Equal(t, []string{"pr_base"}, baseRun.Labels)
	assert.Equal(t, []string{"pr_head"}, headRun.Labels)
}
