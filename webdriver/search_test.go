// +build large

package webdriver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func TestSearch(t *testing.T) {
	devAppserverInstance, err := NewWebserver()
	if err != nil {
		panic(err)
	}
	defer devAppserverInstance.Close()

	service, wd, err := FirefoxWebDriver()
	defer service.Stop()
	defer wd.Quit()

	// Navigate to the wpt.fyi homepage.
	if err := wd.Get(devAppserverInstance.GetWebappURL("/")); err != nil {
		panic(err)
	}

	// Wait for the results view to load.
	runsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		results, err := wd.FindElements(selenium.ByCSSSelector, "path-part")
		if err != nil {
			return false, err
		}
		return len(results) > 0, nil
	}
	wd.WaitWithTimeout(runsLoadedCondition, time.Second*10)

	// Run the search
	searchBox, err := wd.FindElement(selenium.ByCSSSelector, "input.query")
	if err != nil {
		panic(err)
	}

	const query = "2dcontext"
	if err := searchBox.SendKeys(query); err != nil {
		panic(err)
	}

	results, err := wd.FindElements(selenium.ByCSSSelector, "path-part")
	if err != nil {
		panic(err)
	}
	assert.Lenf(t, results, 1, "Expected exactly 1 '%s' search result.", query)
	text, err := results[0].Text()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, "2dcontext/", text)
}
