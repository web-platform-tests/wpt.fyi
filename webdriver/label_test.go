// +build large

package webdriver

import (
	"fmt"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestLabelParam_Results(t *testing.T) {
	runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
		// Local static data only have 2 experimental browsers, and neither has aligned
		// experimental runs.
		if *staging {
			testLabel(t, wd, app, "/", "experimental", "wpt-results", 4, false)
		} else {
			testLabel(t, wd, app, "/", "experimental", "wpt-results", 3, false)
		}
	})

}

func TestLabelParam_Interop(t *testing.T) {
	for _, aligned := range []bool{true, false} {
		t.Run(fmt.Sprintf("?aligned=%v", aligned), func(t *testing.T) {
			runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
				testLabel(t, wd, app, "/interop/", shared.StableLabel, "wpt-interop", 4, aligned)
			})
		})
	}
}

func testLabel(
	t *testing.T,
	wd selenium.WebDriver,
	app AppServer,
	path, label, elementName string,
	runs int,
	aligned bool) {
	// Navigate to the wpt.fyi homepage.
	filters := shared.TestRunFilter{
		Labels:  mapset.NewSetWith(label),
		Aligned: &aligned,
	}
	url := fmt.Sprintf("%s?%s", path, filters.ToQuery().Encode())
	if err := wd.Get(app.GetWebappURL(url)); err != nil {
		assert.FailNow(t, fmt.Sprintf("Failed to load %s: %s", url, err.Error()))
	}

	// Wait for the results view to load.
	runsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		testRuns, err := getTestRunElements(wd, elementName)
		if err != nil {
			return false, err
		}
		return len(testRuns) > 0, nil
	}
	if err := wd.WaitWithTimeout(runsLoadedCondition, time.Second*10); err != nil {
		assert.FailNow(t, fmt.Sprintf("Error waiting for test runs: %s", err.Error()))
	}

	// Check loaded test runs
	testRuns, err := getTestRunElements(wd, elementName)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("Failed to get test runs: %s", err.Error()))
	}
	assert.Lenf(t, testRuns, runs, "Expected exactly %v TestRuns search result.", runs)
	if aligned {
		assertAligned(t, wd, testRuns)
	}

	// Check tab URLs propagate label
	tabs, err := getTabElements(wd, elementName)
	assert.Len(t, tabs, 2)
	for _, tab := range tabs {
		a, err := tab.FindElement(selenium.ByTagName, "a")
		assert.Nil(t, err)
		assert.NotNil(t, a)
		href, err := a.GetAttribute("href")
		assert.Nil(t, err)
		assert.Contains(t, href, "label="+label)
	}
}

func getTestRunElements(wd selenium.WebDriver, element string) ([]selenium.WebElement, error) {
	e, err := wd.FindElement(selenium.ByCSSSelector, element)
	if err != nil {
		return nil, err
	}
	return FindShadowElements(wd, e, "test-run")
}

func getTabElements(wd selenium.WebDriver, element string) ([]selenium.WebElement, error) {
	e, err := wd.FindElement(selenium.ByCSSSelector, element)
	if err != nil {
		return nil, err
	}
	return FindShadowElements(wd, e, "results-tabs", "paper-tab")
}

func assertAligned(t *testing.T, wd selenium.WebDriver, testRuns []selenium.WebElement) {
	if len(testRuns) < 2 {
		return
	}
	args := []interface{}{testRuns[0]}
	shaProp := "return arguments[0].testRun.revision"
	sha, _ := wd.ExecuteScriptRaw(shaProp, args)
	assert.NotEqual(t, sha, "")
	for i := 1; i < len(testRuns); i++ {
		args = []interface{}{testRuns[0]}
		otherSHA, _ := wd.ExecuteScriptRaw(shaProp, args)
		assert.Equal(t, sha, otherSHA)
	}
}
