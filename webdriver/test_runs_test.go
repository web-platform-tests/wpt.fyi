// +build large

package webdriver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func TestTestRuns(t *testing.T) {
	app, err := NewWebserver()
	if err != nil {
		panic(err)
	}
	defer app.Close()

	service, wd, err := GetWebDriver()
	defer service.Stop()
	defer wd.Quit()

	// Navigate to the wpt.fyi homepage.
	if err := wd.Get(app.GetWebappURL("/test-runs")); err != nil {
		panic(err)
	}

	// Wait for the results view to load.
	runsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		rows, err := getRunRowElements(wd)
		if err != nil {
			return false, err
		}
		return len(rows) > 1, nil
	}
	err = wd.WaitWithTimeout(runsLoadedCondition, time.Second*10)
	assert.Nil(t, err)
}

func getRunRowElements(wd selenium.WebDriver) ([]selenium.WebElement, error) {
	switch *browser {
	case "firefox":
		return wd.FindElements(selenium.ByCSSSelector, "wpt-runs tr")
	default:
		e, err := wd.FindElement(selenium.ByCSSSelector, "wpt-runs")
		if err != nil {
			return nil, err
		}
		return FindShadowElements(wd, e, "tr")
	}
}
