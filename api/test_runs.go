// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/datastore"
)

const nextPageTokenHeaderName = "wpt-next-page"
const paginationTokenFeatureFlagName = "paginationTokens"

// apiTestRunsHandler is responsible for emitting test-run JSON for all the runs at a given SHA.
//
// URL Params:
//     sha: SHA[0:10] of the repo when the tests were executed (or 'latest')
func apiTestRunsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)
	store := shared.NewAppEngineDatastore(ctx, true)
	aeAPI := shared.NewAppEngineAPI(ctx)
	q := r.URL.Query()
	ids, err := shared.ParseRunIDsParam(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pr, err := shared.ParsePRParam(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var testRuns shared.TestRuns
	var nextPageToken string
	if len(ids) > 0 {
		testRuns, err = ids.LoadTestRuns(store)
		if err == datastore.ErrNoSuchEntity {
			w.WriteHeader(http.StatusNotFound)
			err = nil
		}
	} else {
		var filters shared.TestRunFilter
		filters, err = shared.ParseTestRunFilterParams(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if pr != nil && aeAPI.IsFeatureEnabled("runsByPRNumber") {
			filters.SHAs = getPRCommits(aeAPI, *pr)
			if len(filters.SHAs) < 1 {
				log.Warningf("PR %v returned no commits from GitHub", *pr)
			} else {
				log.Infof("PR %v returned %v commits: %s", *pr, len(filters.SHAs), strings.Join(filters.SHAs.ShortSHAs(), ","))
			}
		}
		var runsByProduct shared.TestRunsByProduct
		runsByProduct, err = LoadTestRunsForFilters(store, filters)

		if err == nil {
			testRuns = runsByProduct.AllRuns()
			if aeAPI.IsFeatureEnabled(paginationTokenFeatureFlagName) {
				nextPage := filters.NextPage(runsByProduct)
				if nextPage != nil {
					nextPageToken, _ = nextPage.Token()
				}
			}
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(testRuns) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("[]"))
		return
	}

	if nextPageToken != "" {
		w.Header().Add(nextPageTokenHeaderName, nextPageToken)
	}

	testRunsBytes, err := json.Marshal(testRuns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(testRunsBytes)
}

// LoadTestRunKeysForFilters deciphers the filters and executes a corresponding
// query to load the TestRun keys.
func LoadTestRunKeysForFilters(store shared.Datastore, filters shared.TestRunFilter) (result shared.KeysByProduct, err error) {
	q := store.TestRunQuery()
	limit := filters.MaxCount
	offset := filters.Offset
	from := filters.From
	// Default to a single, latest run when not using any "more than one results" filters.
	if limit == nil && from == nil && len(filters.SHAs) < 2 {
		one := 1
		limit = &one
	}
	products := filters.GetProductsOrDefault()

	// When ?aligned=true, make sure to show results for the same aligned run (executed for all browsers).
	if filters.SHAs.EmptyOrLatest() && filters.Aligned != nil && *filters.Aligned {
		shas, shaKeys, err := q.GetAlignedRunSHAs(products, filters.Labels, from, filters.To, limit, filters.Offset)
		if err != nil {
			return result, err
		}
		if len(shas) < 1 {
			// Bail out early - can't find any complete runs.
			return result, nil
		}
		keys := make(shared.KeysByProduct, len(products))
		for _, sha := range shas {
			for i := range shaKeys[sha] {
				keys[i].Keys = append(keys[i].Keys, shaKeys[sha][i].Keys...)
			}
		}
		return keys, err
	}
	return q.LoadTestRunKeys(products, filters.Labels, filters.SHAs, from, filters.To, limit, offset)
}

// LoadTestRunsForFilters deciphers the filters and executes a corresponding query to load
// the TestRuns.
func LoadTestRunsForFilters(store shared.Datastore, filters shared.TestRunFilter) (result shared.TestRunsByProduct, err error) {
	var keys shared.KeysByProduct
	if keys, err = LoadTestRunKeysForFilters(store, filters); err != nil {
		return nil, err
	}
	return store.TestRunQuery().LoadTestRunsByKeys(keys)
}

func getPRCommits(aeAPI shared.AppEngineAPI, pr int) shared.SHAs {
	log := shared.GetLogger(aeAPI.Context())

	githubClient, err := aeAPI.GetGitHubClient()
	if err != nil {
		log.Errorf("Failed to get github client: %s", err.Error())
		return nil
	}
	commits, _, err := githubClient.PullRequests.ListCommits(aeAPI.Context(), shared.WPTRepoOwner, shared.WPTRepoName, pr, nil)
	if err != nil || commits == nil {
		log.Errorf("Failed to fetch PR #%v: %s", pr, err.Error())
		return nil
	}
	shas := make([]string, len(commits))
	for i := range commits {
		shas[i] = commits[i].GetSHA()
	}
	return shas
}
