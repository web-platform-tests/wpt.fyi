// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"regexp"
)

var (
	// This is the pattern for the downloadURL field in
	// https://docs.microsoft.com/en-us/rest/api/azure/devops/build/artifacts/get?view=azure-devops-rest-4.1
	azureArtifactRegex = regexp.MustCompile(`/_apis/build/builds/[0-9]+/artifacts\?artifactName=([^&]+)`)
)

func getAzureArtifactName(url string) string {
	if match := azureArtifactRegex.FindStringSubmatch(url); len(match) > 1 {
		return match[1]
	}
	return ""
}
