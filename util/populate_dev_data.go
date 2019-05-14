// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/remote_api"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/metrics"
)

var (
	localHost          = flag.String("local_host", "localhost:8080", "local dev_appserver.py webapp host")
	localRemoteAPIHost = flag.String("local_remote_api_host", "localhost:9999", "local dev_appserver.py host for the remote API")
	remoteHost         = flag.String("remote_host", "wpt.fyi", "wpt.fyi host to fetch prod runs from")
	numRemoteRuns      = flag.Int("num_remote_runs", 10, "number of remote runs to copy from host to local environment")
	staticRuns         = flag.Bool("static_runs", false, "Include runs in the /static dir")
	remoteRuns         = flag.Bool("remote_runs", true, "Include copies of remote runs")
	seenTestRunIDs     = mapset.NewSet()
)

// populate_dev_data.go populates a local running webapp instance with some
// of the latest production entities, so that there's data to view.
//
// Usage (from util/):
// go run populate_dev_data.go
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	ctx, err := getRemoteAPIContext()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Adding dev data to host %s...", *localRemoteAPIHost)

	emptySecretToken := &shared.Token{}
	enabledFlag := &shared.Flag{Enabled: true}
	staticDataTime, _ := time.Parse(time.RFC3339, "2019-05-13T00:00:00Z")

	// Follow pattern established in run/*.py data collection code.
	const staticRunSHA = "24278ab61781de72ed363b866ae6b50b86822b27"
	summaryURLFmtString := "http://%s/static/%s/%s"
	chrome := shared.TestRun{
		ProductAtRevision: shared.ProductAtRevision{
			Product: shared.Product{
				BrowserName:    "chrome",
				BrowserVersion: "74.0",
				OSName:         "linux",
				OSVersion:      "3.16",
			},
			FullRevisionHash: staticRunSHA,
			Revision:         staticRunSHA[:10],
		},
		ResultsURL: fmt.Sprintf(summaryURLFmtString, *localHost, staticRunSHA[:10], "chrome[stable].json"),
		CreatedAt:  staticDataTime,
		TimeStart:  staticDataTime,
		Labels:     []string{"chrome", shared.StableLabel},
	}
	chromeExp := chrome
	chromeExp.BrowserVersion = "76.0"
	chromeExp.Labels = []string{"chrome", shared.ExperimentalLabel}
	chromeExp.ResultsURL = strings.Replace(chrome.ResultsURL, "[stable]", "[experimental]", -1)

	edge := chrome
	edge.BrowserName = "edge"
	edge.BrowserVersion = "18"
	edge.OSName = "windows"
	edge.OSVersion = "10"
	edge.ResultsURL = fmt.Sprintf(summaryURLFmtString, *localHost, staticRunSHA[:10], "edge[stable].json")
	edge.Labels = []string{"edge", shared.StableLabel}

	firefox := chrome
	firefox.BrowserName = "firefox"
	firefox.BrowserVersion = "66"
	firefox.ResultsURL = fmt.Sprintf(summaryURLFmtString, *localHost, staticRunSHA[:10], "firefox[stable].json")
	firefox.Labels = []string{"firefox", shared.StableLabel}
	firefoxExp := firefox
	firefoxExp.BrowserVersion = "68.0"
	firefoxExp.Labels = []string{"firefox", shared.ExperimentalLabel}
	firefoxExp.ResultsURL = strings.Replace(firefox.ResultsURL, "[stable]", "[experimental]", -1)

	safari := chrome
	safari.BrowserName = "safari"
	safari.BrowserVersion = "12.1"
	safari.OSName = "mac"
	safari.OSName = "10.13"
	safari.ResultsURL = fmt.Sprintf(summaryURLFmtString, *localHost, staticRunSHA[:10], "safari[stable].json")
	safari.Labels = []string{"safari", shared.StableLabel}
	safariExp := safari
	safariExp.BrowserVersion = "81 preview"
	safariExp.Labels = []string{"safari", shared.ExperimentalLabel}
	safariExp.ResultsURL = strings.Replace(safari.ResultsURL, "[stable]", "[experimental]", -1)

	staticTestRuns := shared.TestRuns{
		chrome,
		chromeExp,
		edge,
		firefox,
		firefoxExp,
		safari,
		safariExp,
	}
	labelRuns(staticTestRuns, "test", "static", shared.MasterLabel)

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

	testRunKindName := "TestRun"
	passRateMetadataKindName := metrics.GetDatastoreKindName(
		metrics.PassRateMetadata{})

	log.Print("Adding local (empty) secrets...")
	addSecretToken(ctx, "upload-token", emptySecretToken)
	addSecretToken(ctx, "github-api-token", emptySecretToken)

	log.Print("Adding flag defaults...")
	addFlag(ctx, "queryBuilder", enabledFlag)
	addFlag(ctx, "diffFilter", enabledFlag)
	addFlag(ctx, "diffFromAPI", enabledFlag)
	addFlag(ctx, "experimentalByDefault", enabledFlag)
	addFlag(ctx, "experimentalAlignedExceptEdge", enabledFlag)
	addFlag(ctx, "structuredQueries", enabledFlag)
	addFlag(ctx, "diffRenames", enabledFlag)
	addFlag(ctx, "paginationTokens", enabledFlag)

	log.Print("Adding uploader \"test\"...")
	addData(ctx, "Uploader", []interface{}{
		&shared.Uploader{Username: "test", Password: "123"},
	})

	if *staticRuns {
		log.Print("Adding local mock data (static/)...")
		for i, key := range addData(ctx, testRunKindName, staticTestRunMetadata) {
			staticTestRuns[i].ID = key.IntID()
		}
		for i := range staticPassRateMetadata {
			md := staticPassRateMetadata[i].(*metrics.PassRateMetadata)
			md.TestRunIDs = staticTestRuns.GetTestRunIDs()
		}
		addData(ctx, passRateMetadataKindName, staticPassRateMetadata)
	}

	if *remoteRuns {
		log.Print("Adding latest production TestRun data...")
		filters := shared.TestRunFilter{
			Labels:   mapset.NewSetWith(shared.StableLabel),
			MaxCount: numRemoteRuns,
		}
		copyProdRuns(ctx, filters)

		log.Print("Adding latest master TestRun data...")
		filters.Labels = mapset.NewSetWith(shared.MasterLabel)
		copyProdRuns(ctx, filters)

		log.Print("Adding latest experimental TestRun data...")
		filters.Labels = mapset.NewSetWith(shared.ExperimentalLabel)
		copyProdRuns(ctx, filters)

		log.Print("Adding latest beta TestRun data...")
		filters.Labels = mapset.NewSetWith(shared.BetaLabel)
		copyProdRuns(ctx, filters)

		log.Print("Adding latest aligned Edge stable and Chrome/Firefox/Safari experimental data...")
		filters.Labels = mapset.NewSet(shared.MasterLabel)
		filters.Products, _ = shared.ParseProductSpecs("chrome[experimental]", "edge[stable]", "firefox[experimental]", "safari[experimental]")
		copyProdRuns(ctx, filters)

		log.Printf("Successfully copied a total of %v distinct TestRuns", seenTestRunIDs.Cardinality())
	}
}

func copyProdRuns(ctx context.Context, filters shared.TestRunFilter) {
	store := shared.NewAppEngineDatastore(ctx, false)
	q := store.TestRunQuery()
	for _, aligned := range []bool{false, true} {
		if aligned {
			filters.Aligned = &aligned
		}
		prodTestRuns, err := shared.FetchRuns(*remoteHost, filters)
		if err != nil {
			log.Print(err)
			continue
		}
		labelRuns(prodTestRuns, "prod")

		latestProductionTestRunMetadata := make([]interface{}, 0, len(prodTestRuns))
		for i := range prodTestRuns {
			if !seenTestRunIDs.Contains(prodTestRuns[i].ID) {
				seenTestRunIDs.Add(prodTestRuns[i].ID)
				latestProductionTestRunMetadata = append(latestProductionTestRunMetadata, &prodTestRuns[i])
			}
		}
		addData(ctx, "TestRun", latestProductionTestRunMetadata)

		passRateMetadataKindName := metrics.GetDatastoreKindName(metrics.PassRateMetadata{})
		filters.MaxCount = nil
		prodPassRateMetadata, err := FetchInterop(*remoteHost, filters)
		if err != nil {
			log.Printf("Failed to fetch interop (?aligned=%v).", aligned)
			continue
		}
		// Update the interop IDs to match the newly-copied local test-run IDs.
		// We re-fetch locally because we might have copied a large number of runs,
		// but only want the latest for interop.
		prodPassRateMetadata.TestRunIDs = make([]int64, len(prodPassRateMetadata.TestRuns))
		one := 1
		sha := shared.LatestSHA
		var localRunCopies shared.TestRuns
		if aligned {
			var shas []string
			var keys map[string]shared.KeysByProduct
			if shas, keys, err = q.GetAlignedRunSHAs(shared.GetDefaultProducts(), filters.Labels, nil, nil, &one, nil); err != nil {
				log.Printf("Failed to load a aligned run SHA: %s", err.Error())
				continue
			}
			if len(shas) > 0 {
				sha = shas[0]
				if loaded, err := q.LoadTestRunsByKeys(keys[sha]); err != nil {
					log.Printf("Failed to load test runs by keys: %s", err.Error())
					continue
				} else {
					localRunCopies = loaded.AllRuns()
				}
			}
		}
		if len(localRunCopies) != len(prodPassRateMetadata.TestRunIDs) {
			log.Printf("Could not find local copies for SHA %s", sha)
			continue
		}
		for i := range prodPassRateMetadata.TestRunIDs {
			prodPassRateMetadata.TestRunIDs[i] = localRunCopies[i].ID
		}
		addData(ctx, passRateMetadataKindName, []interface{}{&prodPassRateMetadata})
	}
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

func addFlag(ctx context.Context, id string, data interface{}) {
	key := datastore.NewKey(ctx, "Flag", id, 0, nil)
	if _, err := datastore.Put(ctx, key, data); err != nil {
		log.Fatalf("Failed to add %s flag: %s", id, err.Error())
	}
	log.Printf("Added %s flag", id)
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
	remoteContext, err := remote_api.NewRemoteContext(*localRemoteAPIHost, http.DefaultClient)
	return remoteContext, err
}

// FetchInterop fetches the PassRateMetadata for the given sha / labels, using
// the API on the given host.
// TODO(lukebjerring): Migrate to results-analysis
func FetchInterop(wptdHost string, filter shared.TestRunFilter) (metrics.PassRateMetadata, error) {
	url := "https://" + wptdHost + "/api/interop"
	url += "?" + filter.OrDefault().ToQuery().Encode()

	var interop metrics.PassRateMetadata
	err := shared.FetchJSON(url, &interop)
	return interop, err
}
