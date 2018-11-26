//+build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine/taskqueue"
)

func TestScheduleResultsTask(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(false)
	if err != nil {
		assert.FailNowf(t, "Failed to create aetest context: %s", err.Error())
	}
	defer done()

	stats, err := taskqueue.QueueStats(ctx, []string{""})
	assert.Nil(t, err)
	assert.Equal(t, stats[0].Tasks, 0)

	a := suitesAPIImpl{
		ctx:   ctx,
		queue: "",
	}
	chrome, _ := shared.ParseProductSpec("chrome[stable]")
	sha := strings.Repeat("0123456789", 4)
	err = a.ScheduleResultsProcessing(sha, chrome)
	assert.Nil(t, err)

	stats, err = taskqueue.QueueStats(ctx, []string{""})
	assert.Nil(t, err)
	assert.Equal(t, stats[0].Tasks, 1)
}
