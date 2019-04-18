// +build small

// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package metrics

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestTestRunsLegacy_Convert(t *testing.T) {
	run := TestRunLegacy{
		ID: 123,
		ProductAtRevision: shared.ProductAtRevision{
			Product: shared.Product{
				BrowserName: "chrome",
			},
			Revision: "1234512345",
		},
	}
	meta := TestRunsMetadataLegacy{
		TestRunIDs: shared.TestRunIDs{
			123,
		},
		TestRuns: []TestRunLegacy{
			run,
		},
	}
	bytes, _ := json.Marshal(meta)
	var metaNew TestRunsMetadata
	json.Unmarshal(bytes, &metaNew)
	assert.Equal(t, meta.TestRunIDs, metaNew.TestRuns.GetTestRunIDs())
	converted, err := ConvertRuns(metaNew.TestRuns)
	assert.Nil(t, err)
	assert.Equal(t, meta.TestRuns, converted)
}

func TestGetDatastoreKindName(t *testing.T) {
	var m PassRateMetadata
	assert.Equal(t, "github.com.web-platform-tests.results-analysis.metrics.PassRateMetadata", GetDatastoreKindName(m))
}
