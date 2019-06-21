// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
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
const onlyChangesAsRegressionsFeature = "onlyChangesAsRegressions"

// updateCheckHandler handles /api/checks/[commit] POST requests.
func updateCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	log := shared.GetLogger(ctx)

	vars := mux.Vars(r)
	sha, err := shared.ParseSHA(vars["commit"])
	if err != nil {
		log.Warningf(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	filter.SHAs = shared.SHAs{sha}
	headRun, baseRun, err := loadRunsToCompare(ctx, filter)
	if err != nil {
		msg := "Could not find runs to compare"
		if err != nil {
			msg = fmt.Sprintf("%s: %s", msg, err.Error())
			log.Errorf(msg)
		}
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	sha = headRun.FullRevisionHash
	aeAPI := shared.NewAppEngineAPI(ctx)
	diffAPI := shared.NewDiffAPI(ctx)
	suites, err := NewAPI(ctx).GetSuitesForSHA(sha)
	if err != nil {
		log.Warningf("Failed to load CheckSuites for %s: %s", sha, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if len(suites) < 1 {
		log.Debugf("No CheckSuites found for %s", sha)
	}

	updatedAny := false
	for _, suite := range suites {
		summaryData, err := getDiffSummary(aeAPI, diffAPI, suite, *baseRun, *headRun)
		if err == shared.ErrRunNotInSearchCache {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		updated, updateErr := updateCheckRunSummary(ctx, summaryData, suite)
		if updateErr != nil {
			err = updateErr
		}
		updatedAny = updatedAny || updated
	}

	if err != nil {
		log.Errorf("Failed to update check_run(s): %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if updatedAny {
		w.Write([]byte("Check(s) updated"))
	} else {
		w.Write([]byte("No check(s) updated"))
	}
}

func loadRunsToCompare(ctx context.Context, filter shared.TestRunFilter) (headRun, baseRun *shared.TestRun, err error) {
	one := 1
	store := shared.NewAppEngineDatastore(ctx, false)
	runs, err := store.TestRunQuery().LoadTestRuns(filter.Products, filter.Labels, filter.SHAs, filter.From, filter.To, &one, nil)
	if err != nil {
		return nil, nil, err
	}
	run := runs.First()
	if run == nil {
		return nil, nil, fmt.Errorf("no test run found for %s @ %s",
			filter.Products[0].String(),
			shared.CropString(filter.SHAs.FirstOrLatest(), 7))
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
	store := shared.NewAppEngineDatastore(ctx, false)
	labels := mapset.NewSetWith(extraLabel)
	runs, err := store.TestRunQuery().LoadTestRuns(filter.Products, labels, filter.SHAs, nil, nil, &one, nil)
	run := runs.First()
	if err != nil {
		return nil, err
	}
	if run == nil {
		err = fmt.Errorf("no test run found for %s @ %s with label %s",
			filter.Products[0].String(), filter.SHAs.FirstOrLatest(), extraLabel)
	}
	return run, err
}

func loadMasterRunBefore(ctx context.Context, filter shared.TestRunFilter, headRun *shared.TestRun) (*shared.TestRun, error) {
	// Get the most recent, but still earlier, master run to compare.
	store := shared.NewAppEngineDatastore(ctx, false)
	one := 1
	to := headRun.TimeStart.Add(-time.Millisecond)
	labels := mapset.NewSetWith(headRun.Channel(), shared.MasterLabel)
	runs, err := store.TestRunQuery().LoadTestRuns(filter.Products, labels, nil, nil, &to, &one, nil)
	baseRun := runs.First()
	if err != nil {
		return nil, err
	}
	if baseRun == nil {
		err = fmt.Errorf("no master run found for %s before %s",
			filter.Products[0].String(), filter.SHAs.FirstOrLatest())
	}
	return baseRun, err
}

func getDiffSummary(aeAPI shared.AppEngineAPI, diffAPI shared.DiffAPI, suite shared.CheckSuite, baseRun, headRun shared.TestRun) (summaries.Summary, error) {
	diffFilter := shared.DiffFilterParam{Added: true, Changed: true, Deleted: true}
	diff, err := diffAPI.GetRunsDiff(baseRun, headRun, diffFilter, nil)
	if err != nil {
		return nil, err
	}

	checkProduct := shared.ProductSpec{
		// [browser]@[sha] is plenty specific, and avoids bad version strings.
		ProductAtRevision: shared.ProductAtRevision{
			Product:  shared.Product{BrowserName: headRun.BrowserName},
			Revision: headRun.Revision,
		},
		Labels: mapset.NewSetWith(baseRun.Channel()),
	}

	diffURL := diffAPI.GetDiffURL(baseRun, headRun, &diffFilter)
	host := aeAPI.GetHostname()
	checkState := summaries.CheckState{
		HostName:   host,
		TestRun:    &headRun,
		Product:    checkProduct,
		HeadSHA:    headRun.FullRevisionHash,
		DetailsURL: diffURL,
		Status:     "completed",
		PRNumbers:  suite.PRNumbers,
	}

	var regressions mapset.Set
	if aeAPI.IsFeatureEnabled("onlyChangesAsRegressions") {
		regressionFilter := shared.DiffFilterParam{Changed: true} // Only changed items
		changeOnlyDiff, err := diffAPI.GetRunsDiff(baseRun, headRun, regressionFilter, nil)
		if err != nil {
			return nil, err
		}
		regressions = changeOnlyDiff.Differences.Regressions()
	} else {
		regressions = diff.Differences.Regressions()
	}
	hasRegressions := regressions.Cardinality() > 0
	neutral := "neutral"
	checkState.Conclusion = &neutral
	checksCanBeNonNeutral := aeAPI.IsFeatureEnabled(failChecksOnRegressionFeature)

	// Set URL path to deepest shared dir.
	var tests []string
	if hasRegressions {
		tests = shared.ToStringSlice(regressions)
	} else {
		tests, _ = shared.MapStringKeys(diff.AfterSummary)
	}
	sharedPath := "/results" + shared.GetSharedPath(tests...)
	diffURL.Path = sharedPath

	var summary summaries.Summary

	resultsComparison := summaries.ResultsComparison{
		BaseRun: baseRun,
		HeadRun: headRun,
		HostURL: fmt.Sprintf("https://%s/", host),
		DiffURL: diffURL.String(),
	}
	if headRun.LabelsSet().Contains(shared.PRHeadLabel) {
		// Deletions are meaningless and abundant comparing to master; ignore them.
		masterDiffFilter := shared.DiffFilterParam{Added: true, Changed: true, Unchanged: true}
		masterDiffURL := diffAPI.GetMasterDiffURL(headRun, &masterDiffFilter)
		masterDiffURL.Path = sharedPath
		resultsComparison.MasterDiffURL = masterDiffURL.String()
	}

	if !hasRegressions {
		collapsed := collapseSummary(diff.AfterSummary, 10)
		data := summaries.Completed{
			CheckState:        checkState,
			ResultsComparison: resultsComparison,
			Results:           make(map[string][]int),
		}
		tests, _ := shared.MapStringKeys(collapsed)
		sort.Strings(tests)
		for _, test := range tests {
			if len(data.Results) < 10 {
				data.Results[test] = collapsed[test]
			} else {
				data.More++
			}
		}
		success := "success"
		data.CheckState.Conclusion = &success
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
		if checksCanBeNonNeutral {
			actionRequired := "action_required"
			data.CheckState.Conclusion = &actionRequired
		}
		summary = data
	}
	return summary, nil
}

type pathKeys []string

func (e pathKeys) Len() int      { return len(e) }
func (e pathKeys) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e pathKeys) Less(i, j int) bool {
	return len(strings.Split(e[i], "/")) > len(strings.Split(e[j], "/"))
}

// collapseDiff collapses a tree of file paths into a smaller tree of folders.
func collapseDiff(diff shared.ResultsDiff, limit int) shared.ResultsDiff {
	keys, _ := shared.MapStringKeys(diff)
	paths := shared.ToStringSlice(collapsePaths(keys, limit))
	result := make(shared.ResultsDiff)
	for k, v := range diff {
		for _, p := range paths {
			if strings.HasPrefix(k, p) {
				result.Add(p, v)
				break
			}
		}
	}
	return result
}

// collapseSummary collapses a tree of file paths into a smaller tree of folders.
func collapseSummary(summary shared.ResultsSummary, limit int) shared.ResultsSummary {
	keys, _ := shared.MapStringKeys(summary)
	paths := shared.ToStringSlice(collapsePaths(keys, limit))
	result := make(shared.ResultsSummary)
	for k, v := range summary {
		for _, p := range paths {
			if strings.HasPrefix(k, p) {
				result.Add(p, v)
				break
			}
		}
	}
	return result
}

func collapsePaths(keys []string, limit int) mapset.Set {
	result := shared.NewSetFromStringSlice(keys)
	// 10 iterations to avoid edge-case infinite looping risk.
	for i := 0; i < 10 && result.Cardinality() > limit; i++ {
		sort.Sort(pathKeys(keys))
		collapsed := mapset.NewSet()
		depth := -1
		for _, k := range keys {
			// Something might have already collapsed down 1 dir into this one.
			if collapsed.Contains(k) {
				continue
			}
			parts := strings.Split(k, "/")
			if parts[len(parts)-1] == "" {
				parts = parts[:len(parts)-1]
			}
			if len(parts) < depth {
				collapsed.Add(k)
				continue
			}

			path := strings.Join(parts[:len(parts)-1], "/") + "/"
			collapsed.Add(path)
			depth = len(parts)
		}
		if i > 0 && depth < 3 {
			break
		}
		keys = shared.ToStringSlice(collapsed)
		result = collapsed
	}
	return result
}
