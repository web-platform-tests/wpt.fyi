// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAzureArtifactName(t *testing.T) {
	url := "https://dev.azure.com/web-platform-tests/b14026b4-9423-4454-858f-bf76cf6d1faa/_apis/build/builds/4230/artifacts?artifactName=results&api-version=5.0&%24format=zip"
	a := getAzureArtifactName(url)
	assert.Equal(t, "results", a)
}
