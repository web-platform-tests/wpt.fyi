// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package metrics

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"cloud.google.com/go/datastore"
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

// Load is part of the datastore.PropertyLoadSaver interface.
// We use it to reset all time to UTC and trim their monotonic clock.
func (t *TestRunsMetadata) Load(ps []datastore.Property) error {
	if err := datastore.LoadStruct(t, ps); err != nil {
		return err
	}
	t.StartTime = t.StartTime.UTC().Round(0)
	t.EndTime = t.EndTime.UTC().Round(0)
	return nil
}

// Save is part of the datastore.PropertyLoadSaver interface.
// Delegate to the default behaviour.
func (t *TestRunsMetadata) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(t)
}

// LoadTestRuns fetches the TestRun entities for the PassRateMetadata's TestRunIDs.
func (t *TestRunsMetadata) LoadTestRuns(ctx context.Context) (err error) {
	ds := shared.NewAppEngineDatastore(ctx, false)
	t.TestRuns, err = t.TestRunIDs.LoadTestRuns(ds)
	return err
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

// Load is part of the datastore.PropertyLoadSaver interface.
// We use it to reset all time to UTC and trim their monotonic clock.
func (r *TestRunLegacy) Load(ps []datastore.Property) error {
	if err := datastore.LoadStruct(r, ps); err != nil {
		return err
	}
	r.CreatedAt = r.CreatedAt.UTC().Round(0)
	r.TimeStart = r.TimeStart.UTC().Round(0)
	r.TimeEnd = r.TimeEnd.UTC().Round(0)
	return nil
}

// Save is part of the datastore.PropertyLoadSaver interface.
// Delegate to the default behaviour.
func (r *TestRunLegacy) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(r)
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
func (t *TestRunsMetadataLegacy) LoadTestRuns(ctx context.Context) (err error) {
	if len(t.TestRuns) == 0 {
		ds := shared.NewAppEngineDatastore(ctx, false)
		newRuns, err := t.TestRunIDs.LoadTestRuns(ds)
		if err != nil {
			return err
		}
		t.TestRuns, err = ConvertRuns(newRuns)
	}
	return err
}

// Load is part of the datastore.PropertyLoadSaver interface.
// We use it to reset all time to UTC and trim their monotonic clock.
func (t *TestRunsMetadataLegacy) Load(ps []datastore.Property) error {
	if err := datastore.LoadStruct(t, ps); err != nil {
		return err
	}
	t.StartTime = t.StartTime.UTC().Round(0)
	t.EndTime = t.EndTime.UTC().Round(0)
	return nil
}

// Save is part of the datastore.PropertyLoadSaver interface.
// Delegate to the default behaviour.
func (t *TestRunsMetadataLegacy) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(t)
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
