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
)

// updateCheckHandler handles /api/checks/[commit] POST requests.
func updateCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)

	vars := mux.Vars(r)
	sha := vars["commit"]
	if len(sha) != 40 {
		msg := fmt.Sprintf("Invalid commit: %s", sha)
		log.Warningf(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Warningf("Failed to parse form: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter, err := shared.ParseTestRunFilterParams(r.Form)
	if err != nil {
		log.Warningf("Failed to parse params: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(filter.Products) < 1 {
		msg := "product param is missing"
		log.Warningf(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	filter.SHA = sha[:10]
	one := 1
	runs, err := shared.LoadTestRuns(ctx, filter.Products, filter.Labels, sha[:10], filter.From, filter.To, &one, nil)
	allRuns := runs.AllRuns()
	if err != nil {
		log.Errorf("Failed to load test runs: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(allRuns) < 1 {
		log.Debugf("No runs found for %s @ %s", filter.Products[0].String(), sha[:7])
		http.NotFound(w, r)
		return
	} else if len(allRuns) > 1 {
		log.Errorf("Failed to load test runs: %s", err.Error())
		http.Error(w, fmt.Sprintf("Expected exactly 1 run, but found %v", len(runs)), http.StatusBadRequest)
		return
	}

	// Get the most recent master run to compare.
	labels := filter.Labels
	if labels == nil {
		labels = mapset.NewSet()
	}
	labels.Add("master")
	masterRuns, err := shared.LoadTestRuns(ctx, filter.Products, labels, "", nil, nil, &one, nil)
	allMasterRuns := masterRuns.AllRuns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(allMasterRuns) < 1 {
		log.Debugf("No masters runs found for %s @ %s", filter.Products[0].String(), sha[:7])
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
	neutral := "neutral"
	checkState.Conclusion = &neutral
	checksCanFailAndPass := shared.IsFeatureEnabled(ctx, "failChecksOnRegression")

	var summary summaries.Summary
	if !regressed {
		host := shared.NewAppEngineAPI(ctx).GetHostname()
		data := summaries.Completed{
			CheckState: checkState,
			HostName:   host,
			HostURL:    fmt.Sprintf("https://%s/", host),
			DiffURL:    getMasterDiffURL(ctx, checkState.HeadSHA, checkState.Product).String(),
			SHAURL:     getURL(ctx, shared.TestRunFilter{SHA: checkState.HeadSHA[:10]}).String(),
		}
		if checksCanFailAndPass {
			success := "success"
			data.CheckState.Conclusion = &success
		}
		summary = data
	} else {
		data := summaries.Regressed{
			CheckState:  checkState,
			Regressions: make(map[string]summaries.BeforeAndAfter),
		}
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
		if checksCanFailAndPass {
			failure := "failure"
			data.CheckState.Conclusion = &failure
		}
		summary = data
	}
	return summary
}
