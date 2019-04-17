// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package metrics

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// SubTest models a single test within a WPT test file.
type SubTest struct {
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	Message *string `json:"message"`
}

// TestResults captures the results of running the tests in a WPT test file.
type TestResults struct {
	Test     string    `json:"test"`
	Status   string    `json:"status"`
	Message  *string   `json:"message"`
	Subtests []SubTest `json:"subtests"`
}

// RunInfo is an alias of ProductAtRevision with a custom marshaler to produce
// the "run_info" object in wptreport.json.
type RunInfo struct {
	shared.ProductAtRevision
}

// MarshalJSON is the custom JSON marshaler that produces field names matching
// the "run_info" object in wptreport.json.
func (r RunInfo) MarshalJSON() ([]byte, error) {
	m := map[string]string{
		"revision":        r.FullRevisionHash,
		"product":         r.BrowserName,
		"browser_version": r.BrowserVersion,
		"os":              r.OSName,
	}
	// Optional field:
	if r.OSVersion != "" {
		m["os_version"] = r.OSVersion
	}
	return json.Marshal(m)
}

// TestResultsReport models the `wpt run` results report JSON file format.
type TestResultsReport struct {
	Results []*TestResults `json:"results"`
	RunInfo RunInfo        `json:"run_info,omitempty"`
}

// TestRunsMetadata is a struct for metadata derived from a group of TestRun entities.
type TestRunsMetadata struct {
	// TestRuns are the TestRun entities, loaded from the TestRunIDs
	TestRuns   shared.TestRuns   `json:"test_runs,omitempty" datastore:"-"`
	TestRunIDs shared.TestRunIDs `json:"-"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	DataURL    string            `json:"url"`
}

// TODO(lukebjerring): Remove TestRunLegacy when old format migrated.

// TestRunLegacy is a copy of the TestRun struct, before the `Labels` field
// was added (which causes an array of array and breaks datastore).
type TestRunLegacy struct {
	ID int64 `json:"id" datastore:"-"`

	shared.ProductAtRevision

	// URL for summary of results, which is derived from raw results.
	ResultsURL string `json:"results_url"`

	// Time when the test run metadata was first created.
	CreatedAt time.Time `json:"created_at"`

	// Time when the test run started.
	TimeStart time.Time `json:"time_start"`

	// Time when the test run ended.
	TimeEnd time.Time `json:"time_end"`

	// URL for raw results JSON object. Resembles the JSON output of the
	// wpt report tool.
	RawResultsURL string `json:"raw_results_url"`

	// Legacy format's Labels are (necessarily) ignored by datastore.
	Labels []string `datastore:"-" json:"labels"`
}

// ConvertRuns converts TestRuns into the legacy format.
func ConvertRuns(runs shared.TestRuns) (converted []TestRunLegacy, err error) {
	if serialized, err := json.Marshal(runs); err != nil {
		return nil, err
	} else if err = json.Unmarshal(serialized, &converted); err != nil {
		return nil, err
	}
	return converted, nil
}

// TODO(lukebjerring): Remove TestRunsMetadataLegacy when old format migrated.

// TestRunsMetadataLegacy is a struct for loading legacy TestRunMetadata entities,
// which may have nested TestRun entities.
type TestRunsMetadataLegacy struct {
	TestRuns   []TestRunLegacy   `json:"test_runs"`
	TestRunIDs shared.TestRunIDs `json:"-"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	DataURL    string            `json:"url"`
}

// LoadTestRuns fetches the TestRun entities for the PassRateMetadata's TestRunIDs.
func (metadata *TestRunsMetadata) LoadTestRuns(ctx context.Context) (err error) {
	ds := shared.NewAppEngineDatastore(ctx, false)
	metadata.TestRuns, err = metadata.TestRunIDs.LoadTestRuns(ds)
	return err
}

// LoadTestRuns fetches the TestRun entities for the PassRateMetadata's TestRunIDs.
func (metadata *TestRunsMetadataLegacy) LoadTestRuns(ctx context.Context) (err error) {
	if len(metadata.TestRuns) == 0 {
		ds := shared.NewAppEngineDatastore(ctx, false)
		newRuns, err := metadata.TestRunIDs.LoadTestRuns(ds)
		if err != nil {
			return err
		}
		metadata.TestRuns, err = ConvertRuns(newRuns)
	}
	return err
}

// PassRateMetadata constitutes metadata capturing:
// - When metric run was performed;
// - What test runs are part of the metric run;
// - Where the metric run results reside (a URL).
type PassRateMetadata struct {
	TestRunsMetadata
}

// TODO(lukebjerring): Remove PassRateMetadataLegacy when old format migrated.

// PassRateMetadataLegacy is a struct for storing a PassRateMetadata entry in the
// datastore, avoiding nested arrays. PassRateMetadata is the legacy format, used for
// loading the entity, for backward compatibility.
type PassRateMetadataLegacy struct {
	TestRunsMetadataLegacy
}

// GetDatastoreKindName gets the full (namespaced) data type name for the given
// interface (whether a pointer or not).
func GetDatastoreKindName(data interface{}) string {
	dataType := reflect.TypeOf(data)
	for dataType.Kind() == reflect.Ptr {
		dataType = reflect.Indirect(reflect.ValueOf(
			data)).Type()
	}
	// This package was originally in another repo. We need to hard
	// code the original repo name here to avoid changing the Kind.
	return "github.com.web-platform-tests.results-analysis.metrics." +
		dataType.Name()
}
