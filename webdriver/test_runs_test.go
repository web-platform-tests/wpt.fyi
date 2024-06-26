// +build large

package webdriver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func TestTestRuns(t *testing.T) {
	runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
		// Navigate to the wpt.fyi homepage.
		if err := wd.Get(app.GetWebappURL("/test-runs")); err != nil {
			assert.FailNow(t, fmt.Sprintf("Error navigating to homepage: %s", err.Error()))
		}

		// Wait for the results view to load.
		runsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
			rows, err := getRunRowElements(wd)
			if err != nil {
				return false, err
			}
			return len(rows) > 1, nil
		}
		err := wd.WaitWithTimeout(runsLoadedCondition, LongTimeout)
		assert.Nil(t, err)
	})
}

func getRunRowElements(wd selenium.WebDriver) ([]selenium.WebElement, error) {
	e, err := wd.FindElement(selenium.ByCSSSelector, "wpt-runs")
	if err != nil {
		return nil, err
	}
	return FindShadowElements(wd, e, "tr")
}
