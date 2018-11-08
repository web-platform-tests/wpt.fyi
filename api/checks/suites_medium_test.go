// +build medium

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
	"google.golang.org/appengine/datastore"
)

func TestGetOrCreateCheckSuite(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	sha := strings.Repeat("abcdef012345", 4)
	suite, err := getOrCreateCheckSuite(ctx, sha, "owner", "repo", 123)
	assert.Nil(t, err)
	assert.NotNil(t, suite)

	suite2, err := getOrCreateCheckSuite(ctx, sha, "owner", "repo", 123)
	assert.Nil(t, err)
	assert.NotNil(t, suite2)
	assert.Equal(t, *suite, *suite2)
	suites := []shared.CheckSuite{}
	datastore.NewQuery("CheckSuite").GetAll(ctx, &suites)
	assert.Len(t, suites, 1)
}
