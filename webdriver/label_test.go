// +build large

package webdriver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func TestLabelParam(t *testing.T) {
	app, err := NewWebserver()
	if err != nil {
		panic(err)
	}
	defer app.Close()

	service, wd, err := GetWebDriver()
	defer service.Stop()
	defer wd.Quit()

	testLabel(t, wd, app, "/?label=experimental", "wpt-results", 2)
	testLabel(t, wd, app, "/interop/?label=stable", "wpt-interop", 4)
}

func testLabel(t *testing.T, wd selenium.WebDriver, app AppServer, path string, elementName string, runs int) {
	// Navigate to the wpt.fyi homepage.
	if err := wd.Get(app.GetWebappURL(path)); err != nil {
		panic(err)
	}

	// Wait for the results view to load.
	runsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		testRuns, err := getTestRunElements(wd, elementName)
		if err != nil {
			return false, err
		}
		return len(testRuns) > 0, nil
	}
	wd.WaitWithTimeout(runsLoadedCondition, time.Second*10)

	// Run the search
	testRuns, err := getTestRunElements(wd, elementName)
	if err != nil {
		panic(err)
	}
	assert.Lenf(t, testRuns, runs, "Expected exactly %v TestRuns search result.", runs)
}

func getTestRunElements(wd selenium.WebDriver, element string) ([]selenium.WebElement, error) {
	switch *browser {
	case "firefox":
		return wd.FindElements(selenium.ByCSSSelector, "test-run")
	default:
		e, err := wd.FindElement(selenium.ByCSSSelector, element)
		if err != nil {
			return nil, err
		}
		return FindShadowElements(wd, e, "test-run")
	}
}
