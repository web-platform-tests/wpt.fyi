// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/w3c/wptdashboard/metrics"
	base "github.com/w3c/wptdashboard/shared"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/remote_api"
	"log"
	"time"
)

// populate_dev_data.go populates a local running webapp instance with some
// of the latest production entities, so that there's data to view.
//
// It uses the AppEngine Remote API, which requires credentials; see:
// https://cloud.google.com/appengine/docs/standard/go/tools/remoteapi/
// https://developers.google.com/identity/protocols/application-default-credentials
//
// Usage (from util/):
// go run populate_dev_data.go
func main() {
	ctx, err := getRemoteApiContext()
	if err != nil {
		log.Fatal(err)
	}

	emptySecretToken := []interface{}{&base.Token{}}
	staticDataTime, _ := time.Parse(time.RFC3339, "2017-10-18T00:00:00Z")

	// Follow pattern established in run/*.py data collection code.
	const staticRunSHA = "b952881825"
	const summaryUrlFmtString = "/static/wptd/" + staticRunSHA + "/%s"
	staticTestRuns := []base.TestRun{
		{
			BrowserName:    "chrome",
			BrowserVersion: "63.0",
			OSName:         "linux",
			OSVersion:      "3.16",
			Revision:       staticRunSHA,
			ResultsURL:     fmt.Sprintf(summaryUrlFmtString, "chrome-63.0-linux-summary.json.gz"),
			CreatedAt:      staticDataTime,
		},
		{
			BrowserName:    "edge",
			BrowserVersion: "15",
			OSName:         "windows",
			OSVersion:      "10",
			Revision:       staticRunSHA,
			ResultsURL:     fmt.Sprintf(summaryUrlFmtString, "edge-15-windows-10-sauce-summary.json.gz"),
			CreatedAt:      staticDataTime,
		},
		{
			BrowserName:    "firefox",
			BrowserVersion: "57.0",
			OSName:         "linux",
			OSVersion:      "*",
			Revision:       staticRunSHA,
			ResultsURL:     fmt.Sprintf(summaryUrlFmtString, "firefox-57.0-linux-summary.json.gz"),
			CreatedAt:      staticDataTime,
		},
		{
			BrowserName:    "safari",
			BrowserVersion: "10",
			OSName:         "macos",
			OSVersion:      "10.12",
			Revision:       staticRunSHA,
			ResultsURL:     fmt.Sprintf(summaryUrlFmtString, "safari-10-macos-10.12-sauce-summary.json.gz"),
			CreatedAt:      staticDataTime,
		},
	}

	timeZero := time.Unix(0, 0)
	// Follow pattern established in metrics/run/*.go data collection code.
	// Use unzipped JSON for local dev.
	const metricsUrlFmtString = "/static/wptd-metrics/0-0/%s.json"
	staticTestRunMetadata := make([]interface{}, len(staticTestRuns))
	for i := range staticTestRuns {
		staticTestRunMetadata[i] = &staticTestRuns[i]
	}
	staticPassRateMetadata := []interface{}{
		&metrics.PassRateMetadata{
			StartTime: timeZero,
			EndTime:   timeZero,
			TestRuns:  staticTestRuns,
			DataUrl:   fmt.Sprintf(metricsUrlFmtString, "pass-rates"),
		},
	}

	staticFailuresMetadata := []interface{}{
		&metrics.FailuresMetadata{
			StartTime:   timeZero,
			EndTime:     timeZero,
			TestRuns:    staticTestRuns,
			DataUrl:     fmt.Sprintf(metricsUrlFmtString, "chrome-failures"),
			BrowserName: "chrome",
		},
		&metrics.FailuresMetadata{
			StartTime:   timeZero,
			EndTime:     timeZero,
			TestRuns:    staticTestRuns,
			DataUrl:     fmt.Sprintf(metricsUrlFmtString, "edge-failures"),
			BrowserName: "edge",
		},
		&metrics.FailuresMetadata{
			StartTime:   timeZero,
			EndTime:     timeZero,
			TestRuns:    staticTestRuns,
			DataUrl:     fmt.Sprintf(metricsUrlFmtString, "firefox-failures"),
			BrowserName: "firefox",
		},
		&metrics.FailuresMetadata{
			StartTime:   timeZero,
			EndTime:     timeZero,
			TestRuns:    staticTestRuns,
			DataUrl:     fmt.Sprintf(metricsUrlFmtString, "safari-failures"),
			BrowserName: "safari",
		},
	}

	tokenKindName := "Token"
	testRunKindName := "TestRun"
	passRateMetadataKindName := metrics.GetDatastoreKindName(
		metrics.PassRateMetadata{})
	failuresMetadataKindName := metrics.GetDatastoreKindName(
		metrics.FailuresMetadata{})

	log.Print("Adding local mock data (static/)...")
	addData(ctx, tokenKindName, emptySecretToken)
	addData(ctx, testRunKindName, staticTestRunMetadata)
	addData(ctx, passRateMetadataKindName, staticPassRateMetadata)
	addData(ctx, failuresMetadataKindName, staticFailuresMetadata)

	log.Print("Adding latest production TestRun data...")
	prodTestRuns := base.FetchLatestRuns("wpt.fyi")
	latestProductionTestRunMetadata := make([]interface{}, len(prodTestRuns))
	for i := range prodTestRuns {
		latestProductionTestRunMetadata[i] = &prodTestRuns[i]
	}
	addData(ctx, testRunKindName, latestProductionTestRunMetadata)
}

func addData(ctx context.Context, kindName string, data []interface{}) {
	keys := make([]*datastore.Key, len(data))
	for i := range data {
		keys[i] = datastore.NewIncompleteKey(ctx, kindName, nil)
	}
	if _, err := datastore.PutMulti(ctx, keys, data); err != nil {
		log.Fatalf("Failed to add %s entities: %s", kindName, err.Error())
	}
	log.Printf("Added %v %s entities", len(data), kindName)
}

func getRemoteApiContext() (context.Context, error) {
	const host = "localhost:9999"
	ctx := context.Background()

	hc, err := google.DefaultClient(ctx,
		"https://www.googleapis.com/auth/appengine.apis",
	)
	if err != nil {
		log.Fatal(err)
	}
	var remoteContext context.Context
	remoteContext, err = remote_api.NewRemoteContext(host, hc)
	return remoteContext, err
}
