// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/deckarep/golang-set"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/api/checks/summaries"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// CheckProcessingQueue is the name of the TaskQueue that handles processing and
// interpretation of TestRun results, in order to update the GitHub checks.
const CheckProcessingQueue = "check-processing"

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
		log.Errorf("Found more that one test run")
		http.Error(w, fmt.Sprintf("Expected exactly 1 run, but found %v", len(runs)), http.StatusBadRequest)
		return
	}
	prRun := allRuns[0]

	// Get the most recent master run to compare.
	labels := filter.Labels
	if labels == nil {
		labels = mapset.NewSet()
	}
	labels.Add("master")
	masterRuns, err := shared.LoadTestRuns(ctx, filter.Products, labels, "", nil, nil, &one, nil)
	masterRun := masterRuns.First()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if masterRun == nil {
		log.Debugf("No masters runs found for %s @ %s", filter.Products[0].String(), sha[:7])
		http.Error(w, "No master run found to compare differences", http.StatusNotFound)
		return
	}

	aeAPI := shared.NewAppEngineAPI(ctx)
	diffAPI := shared.NewDiffAPI(ctx)
	summaryData, err := getDiffSummary(aeAPI, diffAPI, *masterRun, prRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	suites, err := NewAPI(ctx).GetSuitesForSHA(sha)
	if err != nil {
		log.Warningf("Failed to load CheckSuites for %s: %s", sha, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if len(suites) < 1 {
		log.Debugf("No CheckSuites found for %s", sha)
	}

	updated, err := updateCheckRun(ctx, summaryData, suites...)
	if err != nil {
		log.Errorf("Failed to update check_run: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if updated {
		w.Write([]byte("Check(s) updated"))
	} else {
		w.Write([]byte("No check(s) updated"))
	}
}

func getDiffSummary(aeAPI shared.AppEngineAPI, diffAPI shared.DiffAPI, masterRun, prRun shared.TestRun) (summaries.Summary, error) {
	diffFilter := shared.DiffFilterParam{Added: true, Changed: true, Unchanged: true}
	diff, err := diffAPI.GetRunsDiff(masterRun, prRun, diffFilter, nil)
	if err != nil {
		return nil, err
	}

	product, _ := shared.ParseProductSpec(prRun.Product.BrowserName)
	diffURL := diffAPI.GetDiffURL(masterRun, prRun, &diffFilter)
	checkState := summaries.CheckState{
		Product:    product,
		HeadSHA:    prRun.FullRevisionHash,
		Title:      getCheckTitle(product),
		DetailsURL: diffURL,
		Status:     "completed",
	}

	regressions := diff.Regressions()
	neutral := "neutral"
	checkState.Conclusion = &neutral
	checksCanFailAndPass := aeAPI.IsFeatureEnabled("failChecksOnRegression")

	var summary summaries.Summary
	host := aeAPI.GetHostname()
	if regressions.Cardinality() > 0 {
		data := summaries.Completed{
			CheckState: checkState,
			HostName:   host,
			HostURL:    fmt.Sprintf("https://%s/", host),
			DiffURL:    diffURL.String(),
			SHAURL:     aeAPI.GetRunsURL(shared.TestRunFilter{SHA: checkState.HeadSHA[:10]}).String(),
		}
		if checksCanFailAndPass {
			success := "success"
			data.CheckState.Conclusion = &success
		}
		summary = data
	} else {
		data := summaries.Regressed{
			MasterRun:     masterRun,
			PRRun:         prRun,
			CheckState:    checkState,
			HostName:      host,
			HostURL:       fmt.Sprintf("https://%s/", host),
			DiffURL:       diffURL.String(),
			MasterDiffURL: diffAPI.GetMasterDiffURL(checkState.HeadSHA, checkState.Product).String(),
			Regressions:   make(map[string]summaries.BeforeAndAfter),
		}
		tests := shared.ToStringSlice(regressions)
		sort.Strings(tests)
		for _, path := range tests {
			if len(data.Regressions) <= 10 {
				ba := summaries.BeforeAndAfter{}
				if b, ok := diff.BeforeSummary[path]; ok {
					ba.PassingBefore = b[0]
					ba.TotalBefore = b[1]
				}
				if a, ok := diff.AfterSummary[path]; ok {
					ba.PassingAfter = a[0]
					ba.TotalAfter = a[1]
				}
				data.Regressions[path] = ba
			} else {
				data.More++
			}
		}
		if checksCanFailAndPass {
			failure := "failure"
			data.CheckState.Conclusion = &failure
		}
		summary = data
	}
	return summary, nil
}
