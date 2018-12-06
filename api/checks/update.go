// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/deckarep/golang-set"

	"github.com/gorilla/mux"
	"github.com/web-platform-tests/wpt.fyi/api/checks/summaries"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// CheckProcessingQueue is the name of the TaskQueue that handles processing and
// interpretation of TestRun results, in order to update the GitHub checks.
const CheckProcessingQueue = "check-processing"

const failChecksOnRegressionFeature = "failChecksOnRegression"

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
	prRun, masterRun, err := loadRunsToCompare(ctx, filter)
	if prRun == nil || masterRun == nil || err != nil {
		msg := "Could not find runs to compare"
		if err != nil {
			msg = fmt.Sprintf("%s: %s", msg, err.Error())
		}
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	aeAPI := shared.NewAppEngineAPI(ctx)
	diffAPI := shared.NewDiffAPI(ctx)
	summaryData, err := getDiffSummary(aeAPI, diffAPI, *masterRun, *prRun)
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

	updated, err := updateCheckRunSummary(ctx, summaryData, suites...)
	if err != nil {
		log.Errorf("Failed to update check_run: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if updated {
		w.Write([]byte("Check(s) updated"))
	} else {
		w.Write([]byte("No check(s) updated"))
	}
}

func loadRunsToCompare(ctx context.Context, filter shared.TestRunFilter) (prRun, masterRun *shared.TestRun, err error) {
	log := shared.GetLogger(ctx)
	one := 1
	runs, err := shared.LoadTestRuns(ctx, filter.Products, filter.Labels, filter.SHA, filter.From, filter.To, &one, nil)
	prRun = runs.First()
	if err != nil {
		log.Errorf("Failed to load test runs: %s", err.Error())
		return nil, nil, err
	} else if prRun == nil {
		log.Debugf("No runs found for %s @ %s", filter.Products[0].String(), filter.SHA[:7])
		return nil, nil, err
	} else if len(runs.AllRuns()) > 1 {
		log.Errorf("Found more that one test run")
		return nil, nil, fmt.Errorf("Expected exactly 1 run, but found %v", len(runs))
	}

	// Get the most recent, but still earlier, master run to compare.
	labels := filter.Labels
	if labels == nil {
		labels = mapset.NewSet()
	}
	labels.Add("master")
	to := prRun.TimeStart.Add(-time.Millisecond)
	masterRuns, err := shared.LoadTestRuns(ctx, filter.Products, labels, "", nil, &to, &one, nil)
	masterRun = masterRuns.First()
	if err != nil {
		log.Errorf("Failed to load master run: %s", err.Error())
		return prRun, nil, err
	} else if masterRun == nil {
		log.Debugf("No masters runs found before %s @ %s", filter.Products[0].String(), filter.SHA[:7])
		return prRun, masterRun, fmt.Errorf("No master run found to compare differences")
	}
	return prRun, masterRun, nil
}

func getDiffSummary(aeAPI shared.AppEngineAPI, diffAPI shared.DiffAPI, masterRun, prRun shared.TestRun) (summaries.Summary, error) {
	diffFilter := shared.DiffFilterParam{Added: true, Changed: true, Unchanged: true}
	diff, err := diffAPI.GetRunsDiff(masterRun, prRun, diffFilter, nil)
	if err != nil {
		return nil, err
	}

	diffURL := diffAPI.GetDiffURL(masterRun, prRun, &diffFilter)
	var labels mapset.Set
	if prRun.IsExperimental() {
		labels = mapset.NewSet(shared.ExperimentalLabel)
	}
	checkState := summaries.CheckState{
		TestRun: &prRun,
		Product: shared.ProductSpec{
			ProductAtRevision: prRun.ProductAtRevision,
			Labels:            labels,
		},
		HeadSHA:    prRun.FullRevisionHash,
		DetailsURL: diffURL,
		Status:     "completed",
	}

	regressions := diff.Differences.Regressions()
	neutral := "neutral"
	checkState.Conclusion = &neutral
	checksCanFailAndPass := aeAPI.IsFeatureEnabled(failChecksOnRegressionFeature)

	var summary summaries.Summary
	host := aeAPI.GetHostname()

	resultsComparison := summaries.ResultsComparison{
		MasterRun:     masterRun,
		PRRun:         prRun,
		HostName:      host,
		HostURL:       fmt.Sprintf("https://%s/", host),
		DiffURL:       diffURL.String(),
		MasterDiffURL: diffAPI.GetMasterDiffURL(checkState.HeadSHA, checkState.Product).String(),
	}

	hasRegressions := regressions.Cardinality() > 0
	if !hasRegressions {
		data := summaries.Completed{
			CheckState:        checkState,
			ResultsComparison: resultsComparison,
			Results:           make(map[string][]int),
		}
		tests, _ := shared.MapStringKeys(diff.AfterSummary)
		sort.Strings(tests)
		for _, test := range tests {
			if len(data.Results) < 10 {
				data.Results[test] = diff.AfterSummary[test]
			} else {
				data.More++
			}
		}
		if checksCanFailAndPass {
			success := "success"
			data.CheckState.Conclusion = &success
		}
		summary = data
	} else {
		data := summaries.Regressed{
			CheckState:        checkState,
			ResultsComparison: resultsComparison,
			Regressions:       make(map[string]summaries.BeforeAndAfter),
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
