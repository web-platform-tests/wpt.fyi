// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine/datastore"
)

func TestAuthenticateUploader(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()
	a := NewAppEngineAPI(ctx)

	assert.False(t, a.AuthenticateUploader("user", "123"))

	key := datastore.NewKey(ctx, "Uploader", "user", 0, nil)
	datastore.Put(ctx, key, &shared.Uploader{Username: "user", Password: "123"})
	assert.True(t, a.AuthenticateUploader("user", "123"))
}
