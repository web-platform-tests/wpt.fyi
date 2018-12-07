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

	mapset "github.com/deckarep/golang-set"

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

	if len(filter.Products) != 1 {
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
			log.Errorf(msg)
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
		log.Errorf("Failed to update check_run(s): %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if updated {
		w.Write([]byte("Check(s) updated"))
	} else {
		w.Write([]byte("No check(s) updated"))
	}
}

func loadRunsToCompare(ctx context.Context, filter shared.TestRunFilter) (headRun, baseRun *shared.TestRun, err error) {
	one := 1
	runs, err := shared.LoadTestRuns(ctx, filter.Products, filter.Labels, filter.SHA, filter.From, filter.To, &one, nil)
	if err != nil {
		return nil, nil, err
	}
	run := runs.First()
	if run == nil {
		return nil, nil, fmt.Errorf("no test run found for %s @ %s", filter.Products[0].String(), filter.SHA[:7])
	}

	labels := run.LabelsSet()
	if labels.Contains(shared.MasterLabel) {
		headRun = run
		baseRun, err = loadMasterRunBefore(ctx, filter, headRun)
	} else if labels.Contains(shared.PRBaseLabel) {
		baseRun = run
		headRun, err = loadPRRun(ctx, filter, shared.PRHeadLabel)
	} else if labels.Contains(shared.PRHeadLabel) {
		headRun = run
		baseRun, err = loadPRRun(ctx, filter, shared.PRBaseLabel)
	} else {
		return nil, nil, fmt.Errorf("test run %d doesn't have pr_base, pr_head or master label", run.ID)
	}

	return headRun, baseRun, err
}

func loadPRRun(ctx context.Context, filter shared.TestRunFilter, extraLabel string) (*shared.TestRun, error) {
	// Find the corresponding pr_base or pr_head run.
	one := 1
	labels := mapset.NewSetWith(extraLabel)
	runs, err := shared.LoadTestRuns(ctx, filter.Products, labels, filter.SHA, nil, nil, &one, nil)
	run := runs.First()
	if err != nil {
		return nil, err
	}
	if run == nil {
		err = fmt.Errorf("no test run found for %s @ %s with label %s",
			filter.Products[0].String(), filter.SHA, extraLabel)
	}
	return run, err
}

func loadMasterRunBefore(ctx context.Context, filter shared.TestRunFilter, headRun *shared.TestRun) (*shared.TestRun, error) {
	// Get the most recent, but still earlier, master run to compare.
	one := 1
	to := headRun.TimeStart.Add(-time.Millisecond)
	labels := mapset.NewSetWith(shared.MasterLabel)
	runs, err := shared.LoadTestRuns(ctx, filter.Products, labels, shared.LatestSHA, nil, &to, &one, nil)
	baseRun := runs.First()
	if err != nil {
		return nil, err
	}
	if baseRun == nil {
		err = fmt.Errorf("no master run found for %s before %s",
			filter.Products[0].String(), filter.SHA)
	}
	return baseRun, err
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
