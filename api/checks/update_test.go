// +build small

package checks

import (
	"context"
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
	diffAPI.EXPECT().GetMasterDiffURL(after.FullRevisionHash, sharedtest.SameProductSpec(before.BrowserName), nil).Return(diffURL)

	summary, err := getDiffSummary(aeAPI, diffAPI, before, after)
	assert.Nil(t, err)
	_, ok := summary.(summaries.Regressed)
	assert.True(t, ok)
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
	diffAPI.EXPECT().GetMasterDiffURL(after.FullRevisionHash, sharedtest.SameProductSpec(before.BrowserName), nil).Return(diffURL)

	summary, err := getDiffSummary(aeAPI, diffAPI, before, after)
	assert.Nil(t, err)
	_, ok := summary.(summaries.Completed)
	assert.True(t, ok)
}

func getBeforeAndAfterRuns() (before, after shared.TestRun) {
	before.FullRevisionHash = strings.Repeat("0", 40)
	before.BrowserName = "chrome"
	after.FullRevisionHash = strings.Repeat("1", 40)
	after.BrowserName = "chrome"
	return before, after
}
