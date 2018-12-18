// +build small

package checks

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/api/checks/summaries"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestGetDiffSummary_Regressed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	before, after := getBeforeAndAfterRuns()
	runDiff := shared.RunDiff{
		Differences: shared.ResultsDiff{"/foo.html": shared.TestDiff{0, 1, 0}},
	}

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
	aeAPI.EXPECT().IsFeatureEnabled(failChecksOnRegressionFeature).Return(false)
	aeAPI.EXPECT().GetHostname()
	diffAPI := sharedtest.NewMockDiffAPI(mockCtrl)
	diffAPI.EXPECT().GetRunsDiff(before, after, gomock.Any(), gomock.Any()).Return(runDiff, nil)
	diffURL, _ := url.Parse("https://wpt.fyi/results?diff")
	diffAPI.EXPECT().GetDiffURL(before, after, gomock.Any()).Return(diffURL)
	diffAPI.EXPECT().GetMasterDiffURL(after, sharedtest.SameDiffFilter("ACU")).Return(diffURL)
	suite := shared.CheckSuite{
		PRNumbers: []int{123},
	}

	summary, err := getDiffSummary(aeAPI, diffAPI, suite, before, after)
	assert.Nil(t, err)
	_, ok := summary.(summaries.Regressed)
	assert.True(t, ok)
	assert.Equal(t, suite.PRNumbers, summary.GetCheckState().PRNumbers)
}

func TestGetDiffSummary_Completed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	before, after := getBeforeAndAfterRuns()
	runDiff := shared.RunDiff{
		Differences: shared.ResultsDiff{"/foo.html": shared.TestDiff{1, 0, 1}},
	}

	aeAPI := sharedtest.NewMockAppEngineAPI(mockCtrl)
	aeAPI.EXPECT().Context().AnyTimes().Return(context.Background())
	aeAPI.EXPECT().IsFeatureEnabled(failChecksOnRegressionFeature).Return(false)
	aeAPI.EXPECT().GetHostname()
	diffAPI := sharedtest.NewMockDiffAPI(mockCtrl)
	diffAPI.EXPECT().GetRunsDiff(before, after, gomock.Any(), gomock.Any()).Return(runDiff, nil)
	diffURL, _ := url.Parse("https://wpt.fyi/results?diff")
	diffAPI.EXPECT().GetDiffURL(before, after, gomock.Any()).Return(diffURL)
	diffAPI.EXPECT().GetMasterDiffURL(after, sharedtest.SameDiffFilter("ACU")).Return(diffURL)
	suite := shared.CheckSuite{
		PRNumbers: []int{123},
	}

	summary, err := getDiffSummary(aeAPI, diffAPI, suite, before, after)
	assert.Nil(t, err)
	_, ok := summary.(summaries.Completed)
	assert.True(t, ok)
	assert.Equal(t, suite.PRNumbers, summary.GetCheckState().PRNumbers)
}

func getBeforeAndAfterRuns() (before, after shared.TestRun) {
	before.FullRevisionHash = strings.Repeat("0", 40)
	before.BrowserName = "chrome"
	before.Labels = []string{shared.PRBaseLabel}
	after.FullRevisionHash = strings.Repeat("1", 40)
	after.BrowserName = "chrome"
	after.Labels = []string{shared.PRHeadLabel}
	return before, after
}

func TestCollapseSummary_Nesting(t *testing.T) {
	diff := shared.ResultsSummary{
		"/foo/test.html":             shared.TestSummary{1, 1},
		"/foo/bar/test.html":         shared.TestSummary{1, 1},
		"/foo/bar/baz/test.html":     shared.TestSummary{1, 1},
		"/foo/bar/baz/qux/test.html": shared.TestSummary{1, 1},
	}
	assert.Equal(t, diff, collapseSummary(diff, 4))
	assert.Equal(t, shared.ResultsSummary{
		"/foo/test.html":     shared.TestSummary{1, 1},
		"/foo/bar/test.html": shared.TestSummary{1, 1},
		"/foo/bar/baz/":      shared.TestSummary{2, 2},
	}, collapseSummary(diff, 3))
	assert.Equal(t, shared.ResultsSummary{
		"/foo/test.html": shared.TestSummary{1, 1},
		"/foo/bar/":      shared.TestSummary{3, 3},
	}, collapseSummary(diff, 2))
	assert.Equal(t, shared.ResultsSummary{
		"/foo/": shared.TestSummary{4, 4},
	}, collapseSummary(diff, 1))
}

func TestCollapseSummary_ManyFiles(t *testing.T) {
	diff := shared.ResultsSummary{}
	for i := 1; i <= 20; i++ {
		diff[fmt.Sprintf("/foo/test%v.html", i)] = shared.TestSummary{1, 1}
		diff[fmt.Sprintf("/bar/test%v.html", i)] = shared.TestSummary{1, 1}
		diff[fmt.Sprintf("/baz/test%v.html", i)] = shared.TestSummary{1, 1}
	}
	assert.Equal(t, shared.ResultsSummary{
		"/foo/": shared.TestSummary{20, 20},
		"/bar/": shared.TestSummary{20, 20},
		"/baz/": shared.TestSummary{20, 20},
	}, collapseSummary(diff, 10))
	// A number too small still does its best to collapse.
	assert.Equal(t, shared.ResultsSummary{
		"/foo/": shared.TestSummary{20, 20},
		"/bar/": shared.TestSummary{20, 20},
		"/baz/": shared.TestSummary{20, 20},
	}, collapseSummary(diff, 1))
}

func TestCollapseDiff_Nesting(t *testing.T) {
	diff := shared.ResultsDiff{
		"/foo/test.html":             shared.TestDiff{1, 1, 1},
		"/foo/bar/test.html":         shared.TestDiff{1, 1, 1},
		"/foo/bar/baz/test.html":     shared.TestDiff{1, 1, 1},
		"/foo/bar/baz/qux/test.html": shared.TestDiff{1, 1, 1},
	}
	assert.Equal(t, diff, collapseDiff(diff, 4))
	assert.Equal(t, shared.ResultsDiff{
		"/foo/test.html":     shared.TestDiff{1, 1, 1},
		"/foo/bar/test.html": shared.TestDiff{1, 1, 1},
		"/foo/bar/baz/":      shared.TestDiff{2, 2, 2},
	}, collapseDiff(diff, 3))
	assert.Equal(t, shared.ResultsDiff{
		"/foo/test.html": shared.TestDiff{1, 1, 1},
		"/foo/bar/":      shared.TestDiff{3, 3, 3},
	}, collapseDiff(diff, 2))
	assert.Equal(t, shared.ResultsDiff{
		"/foo/": shared.TestDiff{4, 4, 4},
	}, collapseDiff(diff, 1))
}

func TestCollapseDiff_ManyFiles(t *testing.T) {
	diff := shared.ResultsDiff{}
	for i := 1; i <= 20; i++ {
		diff[fmt.Sprintf("/foo/test%v.html", i)] = shared.TestDiff{1, 1, 1}
		diff[fmt.Sprintf("/bar/test%v.html", i)] = shared.TestDiff{1, 1, 1}
		diff[fmt.Sprintf("/baz/test%v.html", i)] = shared.TestDiff{1, 1, 1}
	}
	assert.Equal(t, shared.ResultsDiff{
		"/foo/": shared.TestDiff{20, 20, 20},
		"/bar/": shared.TestDiff{20, 20, 20},
		"/baz/": shared.TestDiff{20, 20, 20},
	}, collapseDiff(diff, 10))
	// A number too small still does its best to collapse.
	assert.Equal(t, shared.ResultsDiff{
		"/foo/": shared.TestDiff{20, 20, 20},
		"/bar/": shared.TestDiff{20, 20, 20},
		"/baz/": shared.TestDiff{20, 20, 20},
	}, collapseDiff(diff, 1))
}
