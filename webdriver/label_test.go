// +build large

package webdriver

import (
	"fmt"
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

	if *staging {
		// We have all 4 experimental browsers on staging.wpt.fyi.
		testLabel(t, wd, app, "/", "experimental", "wpt-results", 4)
	} else {
		// Local static data only have 2 experimental browsers.
		testLabel(t, wd, app, "/", "experimental", "wpt-results", 2)
	}
	testLabel(t, wd, app, "/interop", "stable", "wpt-interop", 4)
}

func testLabel(
	t *testing.T,
	wd selenium.WebDriver,
	app AppServer,
	path, label, elementName string,
	runs int) {
	// Navigate to the wpt.fyi homepage.
	url := fmt.Sprintf("%s?label=%s", path, label)
	if err := wd.Get(app.GetWebappURL(url)); err != nil {
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

	// Check loaded test runs
	testRuns, err := getTestRunElements(wd, elementName)
	if err != nil {
		panic(err)
	}
	assert.Lenf(t, testRuns, runs, "Expected exactly %v TestRuns search result.", runs)

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

func getTabElements(wd selenium.WebDriver, element string) ([]selenium.WebElement, error) {
	switch *browser {
	case "firefox":
		return wd.FindElements(selenium.ByCSSSelector, "results-navigation paper-tab")
	default:
		e, err := wd.FindElement(selenium.ByCSSSelector, element)
		if err != nil {
			return nil, err
		}
		return FindShadowElements(wd, e, "results-navigation", "paper-tab")
	}
}
