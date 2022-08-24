// +build large

package webdriver

import (
	"fmt"
	"testing"

	mapset "github.com/deckarep/golang-set"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestLabelParam_Results(t *testing.T) {
	runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
		aligned := false
		testLabel(t, wd, app, "/", "experimental", "wpt-results", 4, aligned)
	})

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
	if err := wd.WaitWithTimeout(runsLoadedCondition, LongTimeout); err != nil {
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
