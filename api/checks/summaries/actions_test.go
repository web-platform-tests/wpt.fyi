//go:build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package summaries

import (
	"testing"

	"github.com/google/go-github/v77/github"
	"github.com/stretchr/testify/assert"
)

// https://developer.github.com/v3/checks/runs/#actions-object
func TestActionCharacterLimits(t *testing.T) {
	actions := []*github.CheckRunAction{
		RecomputeAction(),
		IgnoreAction(),
		CancelAction(),
	}
	for _, action := range actions {
		assert.True(t, len(action.Identifier) <= 20, "Action %s's ID is too long", action.Identifier)
		assert.True(t, len(action.Description) <= 40, "Action %s's desc is too long", action.Identifier)
		assert.True(t, len(action.Label) <= 20, "Action %s's label is too long", action.Identifier)
	}
}
