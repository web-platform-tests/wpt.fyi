// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deckarep/golang-set"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/api/checks/summaries"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
)

// updateCheckHandler handles /api/checks/[commit] POST requests.
func updateCheckHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sha := vars["commit"]
	if len(sha) != 40 {
		http.Error(w, fmt.Sprintf("Invalid commit: %s", sha), http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)
	filter, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if len(filter.Products) < 1 {
		http.Error(w, "product param is missing", http.StatusBadRequest)
		return
	}
	filter.SHA = sha[:10]
	one := 1
	runs, err := shared.LoadTestRuns(ctx, filter.Products, filter.Labels, sha[:10], filter.From, filter.To, &one, nil)
	allRuns := runs.AllRuns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(allRuns) < 1 {
		http.NotFound(w, r)
		return
	} else if len(allRuns) > 1 {
		http.Error(w, fmt.Sprintf("Expected exactly 1 run, but found %v", len(runs)), http.StatusBadRequest)
		return
	}

	// Get a master run to compare.
	labels := filter.Labels
	if labels == nil {
		labels = mapset.NewSet()
	}
	labels.Add("master")
	masterRuns, err := shared.LoadTestRuns(ctx, filter.Products, labels, "", nil, nil, nil, nil)
	allMasterRuns := masterRuns.AllRuns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(allMasterRuns) < 1 {
		http.Error(w, "No master run found to compare differences", http.StatusNotFound)
		return
	}

	beforeJSON, err := shared.FetchRunResultsJSON(ctx, r, allRuns[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch 'before' results: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	afterJSON, err := shared.FetchRunResultsJSON(ctx, r, allMasterRuns[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch 'after' results: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	product, _ := shared.ParseProductSpec(runs[0].Product.String())
	checkState := summaries.CheckState{
		Product:    product,
		HeadSHA:    sha,
		Title:      getCheckTitle(product),
		DetailsURL: getMasterDiffURL(ctx, sha, product),
		Status:     "in_progress",
	}
	summaryData := getDiffSummary(ctx, beforeJSON, afterJSON, checkState)

	updated, err := updateCheckRun(ctx, summaryData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if updated {
		w.Write([]byte("Check(s) updated"))
	} else {
		w.Write([]byte("No check(s) updated"))
	}
}

func getDiffSummary(ctx context.Context, before, after map[string][]int, checkState summaries.CheckState) summaries.Summary {
	diffFilter := shared.DiffFilterParam{Added: true, Changed: true, Unchanged: true}
	diff := shared.GetResultsDiff(before, after, diffFilter, nil, nil)

	regressed := false
	for _, d := range diff {
		if d[1] != 0 {
			regressed = true
			break
		}
	}
	if !regressed || !shared.IsFeatureEnabled(ctx, "failChecksOnRegression") {
		host := shared.NewAppEngineAPI(ctx).GetHostname()
		data := summaries.Completed{
			CheckState: checkState,
			HostName:   host,
			HostURL:    fmt.Sprintf("https://%s/", host),
			DiffURL:    getMasterDiffURL(ctx, checkState.HeadSHA, checkState.Product).String(),
			SHAURL:     getURL(ctx, shared.TestRunFilter{SHA: checkState.HeadSHA[:10]}).String(),
		}
		neutral := "neutral"
		data.CheckState.Conclusion = &neutral
		return data
	}

	data := summaries.Regressed{
		CheckState:  checkState,
		Regressions: make(map[string]summaries.BeforeAndAfter),
	}
	failure := "failure"
	data.CheckState.Conclusion = &failure
	for path, d := range diff {
		if d[1] != 0 {
			if len(data.Regressions) <= 10 {
				ba := summaries.BeforeAndAfter{}
				if b, ok := before[path]; ok {
					ba.PassingBefore = b[0]
					ba.TotalBefore = b[1]
				}
				if a, ok := after[path]; ok {
					ba.PassingAfter = a[0]
					ba.TotalAfter = a[1]
				}
				data.Regressions[path] = ba
			} else {
				data.More++
			}
		}
	}
	return data
}
