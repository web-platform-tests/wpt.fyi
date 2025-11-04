//go:build cloud

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// TestAuthenticateUploader relies on the setup of
// TestCloudSecretManagerGetSecret in
// shared/secret_manager_cloud_cloud_test.go
func TestAuthenticateUploader(t *testing.T) {
	ctx := context.Background()
	err := shared.Clients.Init(ctx)
	require.NoError(t, err)
	a := NewAPI(ctx)

	req := httptest.NewRequest("", "/api/foo", &bytes.Buffer{})
	assert.Equal(t, "", AuthenticateUploader(a, req))

	// Case 1: Try to get an uploader that does not exist
	req.SetBasicAuth("bad-test-secret", "bad-value")
	assert.Equal(t, "", AuthenticateUploader(a, req))

	// Case 2: Try with correct username and password
	req.SetBasicAuth("test-secret", "test-secret-value")
	assert.Equal(t, "test-secret", AuthenticateUploader(a, req))

	// Case 3: Try with correct username but bad password
	req.SetBasicAuth("test-secret", "456")
	assert.Equal(t, "", AuthenticateUploader(a, req))
}
