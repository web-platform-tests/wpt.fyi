// +build large

package webdriver

import (
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

func TestFileResults(t *testing.T) {
	app, err := NewWebserver()
	if err != nil {
		panic(err)
	}
	defer app.Close()

	service, wd, err := GetWebDriver()
	if err != nil {
		panic(err)
	}
	defer service.Stop()
	defer wd.Quit()

	// Navigate to an individual test file result.
	url := "/results/infrastructure/assumptions/ahem.html"
	if err := wd.Get(app.GetWebappURL(url)); err != nil {
		panic(err)
	}

	// Wait for the results view to load.
	runsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		testRuns, err := getFileResultRows(wd)
		if err != nil {
			return false, err
		}
		return len(testRuns) > 0, nil
	}
	wd.WaitWithTimeout(runsLoadedCondition, time.Second*10)
}

func getFileResultRows(wd selenium.WebDriver) ([]selenium.WebElement, error) {
	switch *browser {
	case "firefox":
		return wd.FindElements(selenium.ByCSSSelector, "tbody tr")
	default:
		e, err := wd.FindElement(selenium.ByCSSSelector, "wpt-results")
		if err != nil {
			return nil, err
		}
		return FindShadowElements(wd, e, "test-file-results", "tbody tr")
	}
}
