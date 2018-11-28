// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestEvictEmpty(t *testing.T) {
	i := NewWPTIndex()
	assert.NotNil(t, i.EvictAnyRun())
}

func TestEvictNonEmpty(t *testing.T) {
	i := NewWPTIndex()
	i.IngestRun(shared.TestRun{})
	assert.Nil(t, i.EvictAnyRun())
}
