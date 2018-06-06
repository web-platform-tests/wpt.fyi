// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/deckarep/golang-set"

	"github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/remote_api"
)

var (
	host = flag.String("host", "wpt.fyi", "wpt.fyi host to fetch prod runs from")
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
	flag.Parse()

	ctx, err := getRemoteAPIContext()
	if err != nil {
		log.Fatal(err)
	}

	emptySecretToken := &shared.Token{}
	staticDataTime, _ := time.Parse(time.RFC3339, "2017-10-18T00:00:00Z")

	// Follow pattern established in run/*.py data collection code.
	const staticRunSHA = "b952881825"
	const summaryURLFmtString = "/static/" + staticRunSHA + "/%s"
	staticTestRuns := shared.TestRuns{
		{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "chrome",
					BrowserVersion: "63.0",
					OSName:         "linux",
					OSVersion:      "3.16",
				},
				Revision: staticRunSHA,
			},
			ResultsURL: fmt.Sprintf(summaryURLFmtString, "chrome-63.0-linux-summary.json.gz"),
			CreatedAt:  staticDataTime,
		},
		{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "edge",
					BrowserVersion: "15",
					OSName:         "windows",
					OSVersion:      "10",
				},
				Revision: staticRunSHA,
			},
			ResultsURL: fmt.Sprintf(summaryURLFmtString, "edge-15-windows-10-sauce-summary.json.gz"),
			CreatedAt:  staticDataTime,
		},
		{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "firefox",
					BrowserVersion: "57.0",
					OSName:         "linux",
					OSVersion:      "*",
				},
				Revision: staticRunSHA,
			},
			ResultsURL: fmt.Sprintf(summaryURLFmtString, "firefox-57.0-linux-summary.json.gz"),
			CreatedAt:  staticDataTime,
		},
		{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "safari",
					BrowserVersion: "10",
					OSName:         "macos",
					OSVersion:      "10.12",
				},
				Revision: staticRunSHA,
			},
			ResultsURL: fmt.Sprintf(summaryURLFmtString, "safari-10-macos-10.12-sauce-summary.json.gz"),
			CreatedAt:  staticDataTime,
		},
	}
	labelRuns(staticTestRuns, "test", "static")

	timeZero := time.Unix(0, 0)
	// Follow pattern established in metrics/run/*.go data collection code.
	// Use unzipped JSON for local dev.
	const metricsURLFmtString = "/static/wptd-metrics/0-0/%s.json"
	staticTestRunMetadata := make([]interface{}, len(staticTestRuns))
	for i := range staticTestRuns {
		staticTestRunMetadata[i] = &staticTestRuns[i]
	}
	staticPassRateMetadata := []interface{}{
		&metrics.PassRateMetadata{
			TestRunsMetadata: metrics.TestRunsMetadata{
				StartTime: timeZero,
				EndTime:   timeZero,
				DataURL:   fmt.Sprintf(metricsURLFmtString, "pass-rates"),
			},
		},
	}

	staticFailuresMetadata := []interface{}{
		&metrics.FailuresMetadata{
			TestRunsMetadata: metrics.TestRunsMetadata{
				StartTime: timeZero,
				EndTime:   timeZero,
				DataURL:   fmt.Sprintf(metricsURLFmtString, "chrome-failures"),
			},
			BrowserName: "chrome",
		},
		&metrics.FailuresMetadata{
			TestRunsMetadata: metrics.TestRunsMetadata{
				StartTime: timeZero,
				EndTime:   timeZero,
				DataURL:   fmt.Sprintf(metricsURLFmtString, "edge-failures"),
			},
			BrowserName: "edge",
		},
		&metrics.FailuresMetadata{
			TestRunsMetadata: metrics.TestRunsMetadata{
				StartTime: timeZero,
				EndTime:   timeZero,
				DataURL:   fmt.Sprintf(metricsURLFmtString, "firefox-failures"),
			},
			BrowserName: "firefox",
		},
		&metrics.FailuresMetadata{
			TestRunsMetadata: metrics.TestRunsMetadata{
				StartTime: timeZero,
				EndTime:   timeZero,
				DataURL:   fmt.Sprintf(metricsURLFmtString, "safari-failures"),
			},
			BrowserName: "safari",
		},
	}

	testRunKindName := "TestRun"
	passRateMetadataKindName := metrics.GetDatastoreKindName(
		metrics.PassRateMetadata{})
	failuresMetadataKindName := metrics.GetDatastoreKindName(
		metrics.FailuresMetadata{})

	log.Print("Adding local (empty) secrets...")
	addSecretToken(ctx, "upload-token", emptySecretToken)
	addSecretToken(ctx, "github-api-token", emptySecretToken)

	log.Print("Adding uploader \"test\"...")
	addData(ctx, "Uploader", []interface{}{
		&shared.Uploader{Username: "test", Password: "123"},
	})

	log.Print("Adding local mock data (static/)...")
	testRunKeys := addData(ctx, testRunKindName, staticTestRunMetadata)
	for i, key := range testRunKeys {
		staticTestRuns[i].ID = key.IntID()
	}
	for i := range staticPassRateMetadata {
		md := staticPassRateMetadata[i].(*metrics.PassRateMetadata)
		md.TestRunIDs = staticTestRuns.GetTestRunIDs()
	}
	for i := range staticFailuresMetadata {
		md := staticFailuresMetadata[i].(*metrics.FailuresMetadata)
		md.TestRunIDs = staticTestRuns.GetTestRunIDs()
	}
	addData(ctx, passRateMetadataKindName, staticPassRateMetadata)
	addData(ctx, failuresMetadataKindName, staticFailuresMetadata)

	log.Print("Adding latest production TestRun data...")
	prodTestRuns := shared.FetchLatestRuns(*host)
	labelRuns(prodTestRuns, "prod")
	latestProductionTestRunMetadata := make([]interface{}, len(prodTestRuns))
	for i := range prodTestRuns {
		latestProductionTestRunMetadata[i] = &prodTestRuns[i]
	}
	addData(ctx, testRunKindName, latestProductionTestRunMetadata)

	log.Print("Adding latest experimental TestRun data...")
	prodTestRuns = shared.FetchRuns(*host, "latest", mapset.NewSet("experimental"))
	labelRuns(prodTestRuns, "prod")

	latestProductionTestRunMetadata = make([]interface{}, len(prodTestRuns))
	for i := range prodTestRuns {
		latestProductionTestRunMetadata[i] = &prodTestRuns[i]
	}
	addData(ctx, testRunKindName, latestProductionTestRunMetadata)
}

func labelRuns(runs []shared.TestRun, labels ...string) {
	for i := range runs {
		for _, label := range labels {
			runs[i].Labels = append(runs[i].Labels, label)
		}
	}
}

func addSecretToken(ctx context.Context, id string, data interface{}) {
	key := datastore.NewKey(ctx, "Token", id, 0, nil)
	if _, err := datastore.Put(ctx, key, data); err != nil {
		log.Fatalf("Failed to add %s secret: %s", id, err.Error())
	}
	log.Printf("Added %s secret", id)
}

func addData(ctx context.Context, kindName string, data []interface{}) (keys []*datastore.Key) {
	keys = make([]*datastore.Key, len(data))
	for i := range data {
		keys[i] = datastore.NewIncompleteKey(ctx, kindName, nil)
	}
	var err error
	if keys, err = datastore.PutMulti(ctx, keys, data); err != nil {
		log.Fatalf("Failed to add %s entities: %s", kindName, err.Error())
	}
	log.Printf("Added %v %s entities", len(data), kindName)
	return keys
}

func getRemoteAPIContext() (context.Context, error) {
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
