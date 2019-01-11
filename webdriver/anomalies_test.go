// +build large

package webdriver

import (
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

func TestAnomalies(t *testing.T) {
	// TODO(markdittmer): Figure out how to make this test fast.
	t.SkipNow()

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

	// Navigate to the wpt.fyi anomalies page.
	path := "/anomalies"
	if err := wd.Get(app.GetWebappURL(path)); err != nil {
		panic(err)
	}

	// Wait for the results view to load.
	runsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		pathParts, err := getAnomalyElements(wd)
		if err != nil {
			return false, err
		}
		return len(pathParts) > 0, nil
	}
	wd.WaitWithTimeout(runsLoadedCondition, time.Second*10)
}

func getAnomalyElements(wd selenium.WebDriver) ([]selenium.WebElement, error) {
	e, err := wd.FindElement(selenium.ByCSSSelector, "wpt-anomalies")
	if err != nil {
		return nil, err
	}
	return FindShadowElements(wd, e, "h2 ~ a")
}
